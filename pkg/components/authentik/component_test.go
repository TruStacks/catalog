package authentik

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trustacks/catalog/pkg/catalog"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetChart(t *testing.T) {
	var conf *catalog.ComponentConfig
	if err := yaml.Unmarshal(config, &conf); err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/%s-%s.tgz", conf.Repo, conf.Chart, conf.Version)
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
	namespace := "test"
	_, err := clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	token, err := getAPIToken(namespace, clientset)
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
	defer ts.Close()
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
	defer ts.Close()
	pk, err := getAuthorizationFlow(ts.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "123", pk, "got an unexpected flow pk")
}

func TestGetSigningKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"results": [{"pk": "123", "name": "authentik Self-signed Certificate"}]}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()
	pk, err := getCertificateKeypair(ts.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "123", pk, "got an unexpected certificate keypair pk")
}

func TestCreateOIDCProvider(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"pk": 123}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()
	mappings := []string{
		"225abbd5-1a2b-44a8-b21d-df7f3c9be735",
		"faee6fea-8b07-40f7-bda8-0c02159cd608",
		"32c4a700-5352-44d2-880a-c5dfb08328f9",
	}
	flow := "c53f70da-aa78-42c1-950a-f0c7e7e324a1"
	signingKey := "62b33e8b-033b-4dc7-9580-0de0a3f457e6"
	pk, id, secret, err := createOIDCProvider("test", ts.URL, "test-token", flow, signingKey, mappings)
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
	defer ts.Close()
	if err := createApplication(123, "test", ts.URL, "test-token"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateAPIToken(t *testing.T) {
	defer patchAPIToken()()
	clientset := fake.NewSimpleClientset()
	namespace := "test"
	if err := createAPIToken(namespace, "test-token", clientset); err != nil {
		t.Fatal(err)
	}
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), apiTokenSecret, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test-token", strings.TrimSpace(string(secret.Data["api-token"])), "got an unexpected token value")

	// check idempotence.
	if err := createAPIToken(namespace, "test-token", clientset); err != nil {
		t.Fatal(err)
	}
}

func TestCreateGroups(t *testing.T) {
	getGroups := make([]string, 0)
	postGroups := make([]string, 0)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			name := r.URL.Query().Get("name")
			// add group to get groups.
			getGroups = append(getGroups, name)
			if name == "admins" {
				// return a result to simulate an existing 'admins' group.
				if _, err := w.Write([]byte(`{"results": [{}]}`)); err != nil {
					t.Fatal(err)
				}
			} else {
				// return no result to invoke group creation.
				if _, err := w.Write([]byte(`{"results": []}`)); err != nil {
					t.Fatal(err)
				}
			}
		case "POST":
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			g := &group{}
			if err := json.Unmarshal(data, &g); err != nil {
				t.Fatal(err)
			}
			// add group to post groups.
			postGroups = append(postGroups, g.Name)
			if _, err := w.Write([]byte(`{}`)); err != nil {
				t.Fatal(err)
			}
		}
	}))
	defer ts.Close()
	if err := createGroups(ts.URL, "test-token"); err != nil {
		t.Fatal(err)
	}
	assert.ElementsMatch(t, getGroups, []string{"admins", "editors", "viewers"})
	assert.ElementsMatch(t, postGroups, []string{"editors", "viewers"})
}

func TestHealthCheckService(t *testing.T) {
	// test the health check with a malforned URL.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := healthCheckService("http://test.trustacks.local", 1, ctx); err == nil {
		t.Fatal("expected a timeout error")
	}
	// test the health check with a valid URL.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	if err := healthCheckService(ts.URL, 1, context.TODO()); err != nil {
		t.Fatal(err)
	}
}
