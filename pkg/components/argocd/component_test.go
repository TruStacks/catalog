package argocd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/functions"
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

func TestCreateOIDCClient(t *testing.T) {
	var p map[string]interface{}
	defer functions.PatchMockFunction("create-oidc-client", func(params map[string]interface{}) (interface{}, error) {
		p = params
		return map[string]interface{}{"clientId": "test-id", "clientSecret": "test-secret"}, nil
	})()
	clientId, clientSecret, err := createOIDCClient("test-provider")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, componentName, p["name"], "got an unexpected provider")
	assert.Equal(t, "test-provider", p["provider"], "got an unexpected provider")
	assert.Equal(t, "test-id", clientId, "got an unexpected client id")
	assert.Equal(t, "test-secret", clientSecret, "got an unexpected client secret")
}

func TestCreateOIDCClientSecret(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	if err := createOIDCClientSecret("test-id", "test-secret", "test", clientset); err != nil {
		t.Fatal(err)
	}
	secret, err := clientset.CoreV1().Secrets("test").Get(context.TODO(), "oidc-client", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test-id", string(secret.Data["id"]), "got an unexpected client id")
	assert.Equal(t, "test-secret", string(secret.Data["secret"]), "got an unexpected client secret")
	assert.Equal(t, "argocd", secret.Labels["app.kubernetes.io/part-of"], "got an unexpected part-of label")
}

func TestGetAdminPassword(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd-initial-admin-secret",
		},
		Data: map[string][]byte{
			"password": []byte("password123"),
		},
	}
	namespace := "test"
	_, err := clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	adminPassword, err := getAdminPassword("test", clientset)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "password123", adminPassword, "got an unexpected admin password")
}

func TestGetAPISessionToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"token": "test-session-token"}`)); err != nil {
			t.Fatal(err)
		}
	}))
	token, err := getAPISessionToken(ts.URL, "password123")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test-session-token", token, "got an unexpected session token")
}

func TestSetSystemUserPassword(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}))
	if err := setServiceAccountPassword(ts.URL, "test-session-token", "current-password", "password"); err != nil {
		t.Fatal(err)
	}
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
