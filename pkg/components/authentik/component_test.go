package authentik

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trustacks/catalog/pkg/catalog"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLoadComponent(t *testing.T) {
	var config catalog.BaseComponent
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
}

func TestGetChart(t *testing.T) {
	var config catalog.BaseComponent
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/%s-%s.tgz", config.Repo, config.Chart, config.Version)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("failed retrieving the helm chart")
	}
}

// patchAPIToken mock patches the api token secret.
func patchAPIToken() func() {
	previousApiTokenSecret := apiTokenSecret
	apiTokenSecret = "test-bootstrap"
	return func() {
		apiTokenSecret = previousApiTokenSecret
	}
}

func TestGetAPIToken(t *testing.T) {
	defer patchAPIToken()()
	clientset := fake.NewSimpleClientset()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-bootstrap",
		},
		Data: map[string][]byte{
			"api-token": []byte("test-token"),
		},
	}
	if err := os.MkdirAll("/var/run/secrets/kubernetes.io/serviceaccount", 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	namespace, err := getNamespace()
	if err != nil {
		t.Fatal(err)
	}
	_, err = clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	token, err := getAPIToken(clientset)
	if err != nil {
		t.Fatal(err)
	}
	if token != "test-token" {
		t.Fatal("got an unexpected token value")
	}
}

func TestGetPropertyMappings(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"results": [
			{"managed": "goauthentik.io/providers/oauth2/scope-email", "pk": "pk1"},
			{"managed": "goauthentik.io/providers/oauth2/scope-openid", "pk": "pk2"},
			{"managed": "goauthentik.io/providers/oauth2/scope-profile", "pk": "pk3"}
		]}`)); err != nil {
			t.Fatal(err)
		}
	}))
	pm, err := getPropertyMappings(ts.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	assert.ElementsMatch(t, pm, []string{"pk1", "pk2", "pk3"}, "got unexpected property mapping identifiers")
}

func TestGetAuthoroizationFlow(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"results": [{"pk": "123", "slug": "default-provider-authorization-explicit-consent"}]}`)); err != nil {
			t.Fatal(err)
		}
	}))
	pk, err := getAuthorizationFlow(ts.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "123", pk, "got an unexpected flow pk")
}

func TestCreateOIDCProvider(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"pk": 123}`)); err != nil {
			t.Fatal(err)
		}
	}))
	mappings := []string{
		"225abbd5-1a2b-44a8-b21d-df7f3c9be735",
		"faee6fea-8b07-40f7-bda8-0c02159cd608",
		"32c4a700-5352-44d2-880a-c5dfb08328f9",
	}
	flow := "c53f70da-aa78-42c1-950a-f0c7e7e324a1"
	pk, id, secret, err := createOIDCProvider("test", ts.URL, "test-token", flow, mappings)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, id, 40, "expected a 40 character id")
	assert.Len(t, secret, 128, "expected a 128 character secret")
	assert.Equal(t, 123, pk, "got an unexpected provider pk")
}

func TestCreateApplication(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}))
	if err := createApplication(123, "test", ts.URL, "test-token"); err != nil {
		t.Fatal(err)
	}
}
