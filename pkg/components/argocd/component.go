package argocd

import (
	"context"
	_ "embed"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/functions"
	"github.com/trustacks/catalog/pkg/hooks"
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
	return createOIDCClientSecret(clientId, clientSecret, namespace, clientset)
}

// createOIDCClient creates the concourse oidc client.
func createOIDCClient(provider string) (string, string, error) {
	result, err := functions.Call("create-oidc-client", map[string]interface{}{
		"name":     componentName,
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

// createOIDCClientSecret creates the oidc client secret.
func createOIDCClientSecret(clientId, clientSecret, namespace string, clientset kubernetes.Interface) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oidc-client",
		},
		Data: map[string][]byte{
			"id":     []byte(clientId),
			"secret": []byte(clientSecret),
		},
	}
	_, err := clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
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
		hooks.PreInstallHook: component.preInstall,
	} {
		if err := hooks.AddHook(componentName, hook, fn); err != nil {
			log.Fatal(err)
		}
	}
}
