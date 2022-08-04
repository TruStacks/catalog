package concourse

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/sethvargo/go-password/password"
	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/functions"
	"github.com/trustacks/catalog/pkg/hooks"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// componentName is the name of the component.
	componentName          = "concourse"
	systemVarsName         = "system-vars"
	systemSecretsName      = "system-secrets"
	applicationVarsName    = "application-vars"
	applicationSecretsName = "application-secrets"
)

var (
	// inClusterNamespace is the path to the in-cluster namespace.
	inClusterNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	// serviceURL is the concourse kubernetes service name.
	serviceURL = "http://concourse-web:8080"
)

type concourse struct {
	catalog.BaseComponent
}

// preInstall creates the concourse oidc client and secrets.
func (c *concourse) preInstall() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	namespace, err := getNamespace()
	if err != nil {
		return err
	}
	provider := os.Getenv("SSO_PROVIDER")
	clientId, clientSecret, err := createOIDCClient(provider)
	if err != nil {
		return err
	}
	if err := createSecrets(clientId, clientSecret, namespace, clientset); err != nil {
		return err
	}
	return nil
}

// generateRSAKeyPair creates an RSA private and public key pair.
func generateRSAKeyPair() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	var privateKeyBuf bytes.Buffer
	if err := pem.Encode(&privateKeyBuf, privateKeyBlock); err != nil {
		return nil, nil, err
	}
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return privateKeyBuf.Bytes(), ssh.MarshalAuthorizedKey(publicKey), nil
}

// createSecrets creates the web and worker secrets.
func createSecrets(clientId, clientSecret, namespace string, clientset kubernetes.Interface) error {
	hostKey, hostKeyPub, err := generateRSAKeyPair()
	if err != nil {
		return err
	}
	workerKey, workerKeyPub, err := generateRSAKeyPair()
	if err != nil {
		return err
	}
	sessionSigningKey, _, err := generateRSAKeyPair()
	if err != nil {
		return err
	}
	webSecrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "concourse-web",
		},
		Data: map[string][]byte{
			"host-key":            hostKey,
			"session-signing-key": sessionSigningKey,
			"worker-key-pub":      workerKeyPub,
			"oidc-client-id":      []byte(clientId),
			"oidc-client-secret":  []byte(clientSecret),
			"local-users":         []byte(fmt.Sprintf("trustacks:%s", password.MustGenerate(32, 10, 0, false, false))),
		},
	}
	workerSecrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "concourse-worker",
		},
		Data: map[string][]byte{
			"host-key-pub": hostKeyPub,
			"worker-key":   workerKey,
		},
	}
	for _, secret := range []*corev1.Secret{webSecrets, workerSecrets} {
		if _, err := clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// createOIDCClient creates the concourse oidc client.
func createOIDCClient(provider string) (string, string, error) {
	params := []byte(fmt.Sprintf(`{"name": "%s", "provider": "%s"}`, componentName, provider))
	result, err := functions.Call("create-oidc-client", params)
	if err != nil {
		return "", "", err
	}
	v, ok := result.(map[string]interface{})
	if !ok {
		return "", "", errors.New("error parsing the result")
	}
	return v["clientId"].(string), v["clientSecret"].(string), nil
}

// downloadFlyCLI downloads the concourse fly cli.
func downloadFlyCLI(url string) (string, error) {
	f, err := ioutil.TempFile("", "fly-cli")
	if err != nil {
		return "", err
	}
	defer f.Close()
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/cli?arch=amd64&platform=linux", url))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}
	if err := os.Chmod(f.Name(), 0755); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// createApplicationHandler downloads the fly cli and runs the
// application creation procedure.
func createApplicationHandler(params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, errors.New("name is required")
	}
	toolchain, ok := params["toolchain"].(string)
	if !ok {
		return nil, errors.New("toolchain is required")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	cli, err := downloadFlyCLI(serviceURL)
	if err != nil {
		return nil, err
	}
	defer os.Remove(cli)
	if err := createApplication(toolchain, name, clientset, cli, runFlyCmd); err != nil {
		return nil, err
	}
	return nil, nil
}

//go:embed pipeline.gotxt
var pipelineTemplate string

// createApplication creates the application pipeline.
func createApplication(toolchain, name string, clientset kubernetes.Interface, cli string, flyCmd func(cli string, args ...string) error) error {
	toolchainNamespace := fmt.Sprintf("trustacks-toolchain-%s", toolchain)
	applicationNamespace := fmt.Sprintf("trustacks-application-%s-%s", toolchain, name)
	vars, varsFrom, err := getApplicationVars(toolchainNamespace, applicationNamespace, clientset)
	if err != nil {
		return err
	}
	secrets, err := getApplicationSecrets(toolchainNamespace, applicationNamespace, clientset)
	if err != nil {
		return err
	}
	// create the pipeline template.
	var tmplBuf bytes.Buffer
	tmpl, err := template.New("pipeline").Parse(pipelineTemplate)
	if err != nil {
		return err
	}
	if err = tmpl.Execute(&tmplBuf, map[string]interface{}{
		"vars":    vars,
		"secrets": secrets,
	}); err != nil {
		return err
	}
	pipeline, err := ioutil.TempFile("", "pipeline")
	if err != nil {
		return err
	}
	defer os.Remove(pipeline.Name())
	if _, err := pipeline.Write(tmplBuf.Bytes()); err != nil {
		return err
	}
	pipeline.Close()

	// get the system user password.
	webSecrets, err := clientset.CoreV1().Secrets(toolchainNamespace).Get(context.TODO(), "concourse-web", metav1.GetOptions{})
	if err != nil {
		return err
	}
	pwd := strings.Split(string(webSecrets.Data["local-users"]), ":")[1]

	// execute fly commands.
	team := fmt.Sprintf("%s-%s", toolchain, name)
	if err := flyCmd(cli, "login", "-c", serviceURL, "--username", "trustacks", "--password", pwd); err != nil {
		return err
	}
	if err := flyCmd(cli, "sync"); err != nil {
		return err
	}
	if err := flyCmd(cli, "set-team", "--team-name", team, "--local-user", "trustacks", "--non-interactive"); err != nil {
		return err
	}
	if err := flyCmd(cli, "set-pipeline", "--team", team, "-p", name, "-c", pipeline.Name(), "--non-interactive", "--load-vars-from", varsFrom); err != nil {
		return err
	}
	return flyCmd(cli, "unpause-pipeline", "-p", name, "--team", team)
}

// getApplicationVars gets the application vars list.
func getApplicationVars(toolchainNamespace, applicationNamespace string, clientset kubernetes.Interface) ([]string, string, error) {
	systemVars, err := clientset.CoreV1().ConfigMaps(toolchainNamespace).Get(context.TODO(), systemVarsName, metav1.GetOptions{})
	if err != nil {
		return nil, "", err
	}
	patch, err := json.Marshal(systemVars.Data)
	if err != nil {
		return nil, "", err
	}
	patch = []byte(fmt.Sprintf(`{"data": %s}`, patch))
	if _, err := clientset.CoreV1().ConfigMaps(applicationNamespace).Patch(context.TODO(), applicationVarsName, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
		return nil, "", err
	}
	applicationVars, err := clientset.CoreV1().ConfigMaps(applicationNamespace).Get(context.TODO(), applicationVarsName, metav1.GetOptions{})
	if err != nil {
		return nil, "", err
	}
	f, err := ioutil.TempFile("", applicationVarsName)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	if _, err := f.Write([]byte("application-vars:\n")); err != nil {
		return nil, "", err
	}
	vars := make([]string, len(applicationVars.Data))
	i := 0
	for k, v := range applicationVars.Data {
		vars[i] = k
		if _, err := f.Write([]byte(fmt.Sprintf("  %s: \"%s\"\n", k, v))); err != nil {
			return nil, "", err
		}
		i++
	}
	return vars, f.Name(), nil
}

// getApplicationSecrets gets the application secrets list.
func getApplicationSecrets(toolchainNamespace, applicationNamespace string, clientset kubernetes.Interface) ([]string, error) {
	systemSecrets, err := clientset.CoreV1().Secrets(toolchainNamespace).Get(context.TODO(), systemSecretsName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	patch, err := json.Marshal(systemSecrets.Data)
	if err != nil {
		return nil, err
	}
	patch = []byte(fmt.Sprintf(`{"data": %s}`, patch))
	if _, err := clientset.CoreV1().Secrets(applicationNamespace).Patch(context.TODO(), applicationSecretsName, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
		return nil, err
	}
	secret, err := clientset.CoreV1().Secrets(applicationNamespace).Get(context.TODO(), applicationSecretsName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secrets := make([]string, len(secret.Data))
	i := 0
	for k := range secret.Data {
		secrets[i] = k
		i++
	}
	return secrets, nil
}

// flyCmd runs the fly command with the provided arguments.
func runFlyCmd(cli string, args ...string) error {
	args = append([]string{"-t", "default"}, args...)
	var outBuf, errBuf bytes.Buffer
	command := exec.Command(cli, args...)
	command.Stdout = &outBuf
	command.Stderr = &errBuf
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), errBuf.String())
	}
	return nil
}

// getNamespace gets the current kubernetes namespace.
func getNamespace() (string, error) {
	data, err := ioutil.ReadFile(inClusterNamespace)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), err
}

//go:embed config.yaml
var config []byte

//go:embed hooks.yaml
var hookManifests []byte

//go:embed application-hooks.yaml
var applicationHookManifests []byte

// Initialize adds the component to the catalog and configures hooks.
func Initialize(c *catalog.ComponentCatalog) {
	var conf *catalog.ComponentConfig
	if err := yaml.Unmarshal(config, &conf); err != nil {
		log.Fatal(err)
	}
	component := &concourse{
		catalog.BaseComponent{
			Repo:             conf.Repo,
			Chart:            conf.Chart,
			Version:          conf.Version,
			Values:           conf.Values,
			Hooks:            string(hookManifests),
			ApplicationHooks: string(applicationHookManifests),
		},
	}
	c.AddComponent(componentName, component)

	// configure hooks.
	for hook, fn := range map[string]func() error{
		hooks.PreInstallHook: component.preInstall,
	} {
		if err := hooks.AddHook(componentName, hook, fn); err != nil {
			log.Fatal(err)
		}
	}

	// configure functions.
	functions.AddCreateApplicationHandler("concourse", createApplicationHandler)
}
