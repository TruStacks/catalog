package inputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAddSystemVars(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	if err := AddSystemVars("test", "test", map[string]string{"name": "joe"}, clientset); err != nil {
		t.Fatal(err)
	}
	if err := AddSystemVars("test", "test", map[string]string{"age": "42"}, clientset); err != nil {
		t.Fatal(err)
	}
	cm, err := clientset.CoreV1().ConfigMaps("test").Get(context.TODO(), systemVarsConfigMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "joe", cm.Data["test.name"], "got an unexpected name variable value")
	assert.Equal(t, "42", cm.Data["test.age"], "got an unexpected age variable value")
}

func TestAddSystemSecrets(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	if err := AddSystemSecrets("test", "test", map[string][]byte{"username": []byte("joe")}, clientset); err != nil {
		t.Fatal(err)
	}
	if err := AddSystemSecrets("test", "test", map[string][]byte{"password": []byte("password")}, clientset); err != nil {
		t.Fatal(err)
	}
	secret, err := clientset.CoreV1().Secrets("test").Get(context.TODO(), systemSecretsSecretName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "joe", string(secret.Data["test.username"]), "got an unexpected username variable value")
	assert.Equal(t, "password", string(secret.Data["test.password"]), "got an unexpected password variable value")
}
