package argocd

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

func TestCreateOIDCClient(t *testing.T) {
	var p map[string]interface{}
	defer functions.PatchMockFunction("create-oidc-client", func(params map[string]interface{}) (interface{}, error) {
		fmt.Println("called")
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
