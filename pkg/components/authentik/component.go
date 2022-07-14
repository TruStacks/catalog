package authentik

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/sethvargo/go-password/password"
	"github.com/trustacks/catalog/pkg/catalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// componentName is the name of the component.
	componentName = "authentik"
	// serviceURL is the url of the authentik service.
	serviceURL = "http://authentik"
)

// apiTokenSecret is the secret where the api token is stored.
var apiTokenSecret = "authentik-bootstrap"

type authentik struct {
	catalog.BaseComponent
}

// getAPIToken gets the api token secret value.
func getAPIToken(clientset kubernetes.Interface) (string, error) {
	namespace, err := getNamespace()
	if err != nil {
		return "", err
	}
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), apiTokenSecret, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(secret.Data["api-token"])), nil
}

// getAPIResource gets the API resource at the provided path.
func getAPIResource(url, resource, token string, search string) ([]byte, error) {
	uri := fmt.Sprintf("%s/api/v3/%s/", url, resource)
	if search != "" {
		uri = fmt.Sprintf("%s?%s", uri, search)
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("'%s' get error: %s", resource, body)
	}
	return body, nil
}

// postAPIResource posts the API resource at the provided path.
func postAPIResource(url, resource, token string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/%s/", url, resource), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("'%s' post error: %s", resource, body)
	}
	return body, nil
}

// getNamespace gets the current kubernetes namespace.
func getNamespace() (string, error) {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), err
}

type propertyMapping struct {
	PK      string `json:"pk"`
	Managed string `json:"managed"`
}

type propertyMappings struct {
	Results []propertyMapping `json:"results"`
}

// getPropertyMappings gets the ids of the oauth2 scope mappings.
func getPropertyMappings(url, token string) ([]string, error) {
	scopes := []string{
		"goauthentik.io/providers/oauth2/scope-email",
		"goauthentik.io/providers/oauth2/scope-openid",
		"goauthentik.io/providers/oauth2/scope-profile",
	}
	pks := make([]string, len(scopes))
	resp, err := getAPIResource(url, "propertymappings/all", token, "")
	if err != nil {
		return nil, err
	}
	pm := &propertyMappings{}
	if err := json.Unmarshal(resp, &pm); err != nil {
		return nil, err
	}
	for _, p := range pm.Results {
		for i, scope := range scopes {
			if p.Managed == scope {
				pks[i] = p.PK
			}
		}
	}
	return pks, nil
}

type flow struct {
	PK   string `json:"pk"`
	Slug string `json:"slug"`
}

type flows struct {
	Results []flow `json:"results"`
}

// getAuthorizationFlow gets the id of the default authorization
// flow.
func getAuthorizationFlow(url, token string) (string, error) {
	resp, err := getAPIResource(url, "flows/instances", token, "")
	if err != nil {
		return "", err
	}
	f := &flows{}
	if err := json.Unmarshal(resp, &f); err != nil {
		return "", err
	}
	for _, flow := range f.Results {
		if flow.Slug == "default-provider-authorization-explicit-consent" {
			return flow.PK, nil
		}
	}
	return "", errors.New("authorization flow not found")
}

// createOIDCProvier creates a new openid connection auth provider.
func createOIDCProvider(name, url, token, flow string, mappings []string) (int, string, string, error) {
	client_id, err := password.Generate(40, 30, 0, false, true)
	if err != nil {
		return -1, "", "", err
	}
	client_secret, err := password.Generate(128, 96, 0, false, true)
	if err != nil {
		return -1, "", "", err
	}
	body := map[string]interface{}{
		"name":               name,
		"authorization_flow": flow,
		"client_type":        "confidential",
		"client_id":          client_id,
		"client_secret":      client_secret,
		"property_mappings":  mappings,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return -1, "", "", err
	}
	resp, err := postAPIResource(url, "providers/oauth2", token, data)
	if err != nil {
		return -1, "", "", err
	}
	provider := map[string]interface{}{}
	if err := json.Unmarshal(resp, &provider); err != nil {
		return -1, "", "", err
	}
	return int(provider["pk"].(float64)), client_id, client_secret, nil
}

// createApplication creates a new application.
func createApplication(provider int, name, url, token string) error {
	body := map[string]interface{}{
		"name":     name,
		"slug":     name,
		"provider": provider,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = postAPIResource(url, "core/applications", token, data)
	return err
}

// CreateOIDCCLient creates a consumable end to end oidc client.
func CreateOIDCCLient(name string) (string, string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", "", err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", err
	}
	token, err := getAPIToken(clientset)
	if err != nil {
		return "", "", err
	}
	mappings, err := getPropertyMappings(serviceURL, token)
	if err != nil {
		return "", "", err
	}
	flow, err := getAuthorizationFlow(serviceURL, token)
	if err != nil {
		return "", "", err
	}
	pk, id, secret, err := createOIDCProvider(name, serviceURL, token, flow, mappings)
	if err != nil {
		return "", "", err
	}
	if err := createApplication(pk, name, serviceURL, token); err != nil {
		return "", "", err
	}
	return id, secret, nil
}

// Initialize adds the component to the catalog and configures hooks.
func Initialize(c *catalog.ComponentCatalog) {
	config, err := catalog.LoadComponentConfig(componentName)
	if err != nil {
		log.Fatal(err)
	}
	component := &authentik{
		catalog.BaseComponent{
			Repo:    config.Repo,
			Chart:   config.Chart,
			Version: config.Version,
			Values:  config.Values,
			Hooks:   config.Hooks,
		},
	}
	c.AddComponent(componentName, component)

	for hook, fn := range map[string]func() error{
		"preInstall":  component.preInstall,
		"postInstall": component.postInstall,
	} {
		if err := catalog.AddHook(componentName, hook, fn); err != nil {
			log.Fatal(err)
		}
	}
}
