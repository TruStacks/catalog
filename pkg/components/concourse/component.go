package concourse

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/functions"
	"github.com/trustacks/catalog/pkg/hooks"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// componentName is the name of the component.
	componentName = "concourse"
	// inClusterNamespace is the path to the in-cluster namespace.
	inClusterNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type concourse struct {
	catalog.BaseComponent
}

// preInstall .
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

// generateRSAKeyPair .
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

// createSecrets .
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

// createOIDCClient .
func createOIDCClient(provider string) (string, string, error) {
	result, err := functions.Call("create-oidc-client", map[string]interface{}{
		"name":     "concourse",
		"provider": provider,
	})
	if err != nil {
		return "", "", err
	}
	v, ok := result.(map[string]interface{})
	if !ok {
		return "", "", errors.New("error parsing the result")
	}
	return v["clientId"].(string), v["clientSecret"].(string), nil
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
	component := &concourse{
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
		hooks.PreInstallHook: component.preInstall,
	} {
		if err := hooks.AddHook(componentName, hook, fn); err != nil {
			log.Fatal(err)
		}
	}
}
