package concourse

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/functions"
	"gopkg.in/yaml.v3"
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

func TestGenerateRSAKeyPair(t *testing.T) {
	private, public, err := generateRSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, string(private), "-----BEGIN RSA PRIVATE KEY-----", "got an unexpected rsa private key")
	assert.Contains(t, string(public), "ssh-rsa", "got an unexpected rsa public key")
}

func TestCreateSecrets(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	if err := createSecrets("test-id", "test-secret", "test", clientset); err != nil {
		t.Fatal(err)
	}
	webSecrets, err := clientset.CoreV1().Secrets("test").Get(context.TODO(), "concourse-web", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, string(webSecrets.Data["host-key"]), "-----BEGIN RSA PRIVATE KEY-----", "got an unexpected host key")
	assert.Contains(t, string(webSecrets.Data["session-signing-key"]), "-----BEGIN RSA PRIVATE KEY-----", "got an unexpected sessions signing key")
	assert.Contains(t, string(webSecrets.Data["worker-key-pub"]), "ssh-rsa", "got an unexpected worker public key")
	assert.Equal(t, "test-id", string(webSecrets.Data["oidc-client-id"]), "got an unexpected oidc client id")
	assert.Equal(t, "test-secret", string(webSecrets.Data["oidc-client-secret"]), "got an unexpected oidc client secret")

	workerSecrets, err := clientset.CoreV1().Secrets("test").Get(context.TODO(), "concourse-worker", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, string(workerSecrets.Data["host-key-pub"]), "ssh-rsa", "got an unexpected worker public key")
	assert.Contains(t, string(workerSecrets.Data["worker-key"]), "-----BEGIN RSA PRIVATE KEY-----", "got an unexpected worker public key")
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
