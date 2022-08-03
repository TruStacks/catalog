package argocd

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sethvargo/go-password/password"
	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/functions"
	"github.com/trustacks/catalog/pkg/hooks"
	"github.com/trustacks/catalog/pkg/inputs"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// componentName is the name of the component.
	componentName = "argo-cd"
	// inClusterNamespace is the path to the in-cluster namespace.
	inClusterNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	// serviceURL is the argo cd kubernetes service name.
	serviceURL = "http://argo-cd-argocd-server"
)

type argocd struct {
	catalog.BaseComponent
}

// preInstall creates the oidc client and secret.
func (c *argocd) preInstall() error {
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
	systemVars := map[string]string{"server": "argo-cd-argocd-server"}
	if err := inputs.AddSystemVars(componentName, namespace, systemVars, clientset); err != nil {
		return err
	}
	return createOIDCClientSecret(clientId, clientSecret, namespace, clientset)
}

// postInstall creates the ci service account.
func (c *argocd) postInstall() error {
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
	adminPassword, err := getAdminPassword(namespace, clientset)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	if err := healthCheckService(serviceURL, 2, ctx); err != nil {
		return err
	}
	token, err := getAPISessionToken(serviceURL, adminPassword)
	if err != nil {
		return err
	}
	log.Println("set service account password")
	pwd := password.MustGenerate(32, 10, 0, false, false)
	if err := setServiceAccountPassword(serviceURL, token, adminPassword, pwd); err != nil {
		return err
	}
	systemSecrets := map[string][]byte{"password": []byte(pwd)}
	return inputs.AddSystemSecrets(componentName, namespace, systemSecrets, clientset)
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

// createOIDCClientSecret creates the oidc client secret.
func createOIDCClientSecret(clientId, clientSecret, namespace string, clientset kubernetes.Interface) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oidc-client",
			Labels: map[string]string{
				"app.kubernetes.io/part-of": "argocd",
			},
		},
		Data: map[string][]byte{
			"id":     []byte(clientId),
			"secret": []byte(clientSecret),
		},
	}
	_, err := clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	return err
}

// healthCheckService checks the health of the argocd service.
func healthCheckService(url string, interval int, ctx context.Context) error {
	for {
		select {
		case <-time.After(time.Second * time.Duration(interval)):
			if _, err := http.Get(url); err != nil {
				log.Println(err)
				continue
			}
		case <-ctx.Done():
			return errors.New("service health check timeout")
		}
		break
	}
	return nil
}

// getAdminPassword gets the initial admin password.
func getAdminPassword(namespace string, clientset kubernetes.Interface) (string, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), "argocd-initial-admin-secret", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["password"]), nil
}

// getAPISessionToken creates an api session token.
func getAPISessionToken(url, password string) (string, error) {
	data := fmt.Sprintf(`{"username": "admin", "password": "%s"}`, password)
	requestBody := bytes.NewBuffer([]byte(data))
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/session", url), "application/json", requestBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	return response["token"].(string), nil
}

// setServiceAccountPassword sets the system service account password.
func setServiceAccountPassword(url, token, currentPassword, password string) error {
	data := fmt.Sprintf(`{"name": "trustacks", "currentPassword": "%s", "newPassword": "%s"}`, currentPassword, password)
	requestBody := bytes.NewBuffer([]byte(data))
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/account/password", url), requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	_, err = http.DefaultClient.Do(req)
	return err
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

// Initialize adds the component to the catalog and configures hooks.
func Initialize(c *catalog.ComponentCatalog) {
	var conf *catalog.ComponentConfig
	if err := yaml.Unmarshal(config, &conf); err != nil {
		log.Fatal(err)
	}
	component := &argocd{
		catalog.BaseComponent{
			Repo:    conf.Repo,
			Chart:   conf.Chart,
			Version: conf.Version,
			Values:  conf.Values,
			Hooks:   string(hookManifests),
		},
	}
	c.AddComponent(componentName, component)

	for hook, fn := range map[string]func() error{
		hooks.PreInstallHook:  component.preInstall,
		hooks.PostInstallHook: component.postInstall,
	} {
		if err := hooks.AddHook(componentName, hook, fn); err != nil {
			log.Fatal(err)
		}
	}
}
