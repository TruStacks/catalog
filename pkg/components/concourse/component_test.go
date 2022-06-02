package concourse

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

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
	assert.Contains(t, string(webSecrets.Data["local-users"]), "trustacks:", "got an unexpected local user value")
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

func TestDownloadFlyCLI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("#!/bin/sh\necho 'hello, world'")); err != nil {
			t.Fatal(err)
		}
	}))
	cli, err := downloadFlyCLI(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cli)
	var outBuf bytes.Buffer
	cmd := exec.Command(cli)
	cmd.Stdout = &outBuf
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "hello, world\n", outBuf.String(), "got an unexpected command output")
}

func TestCreateApplication(t *testing.T) {
	// patch the in cluster namespace file.
	f, err := os.CreateTemp("", "in-cluster-namespace")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("test")); err != nil {
		t.Fatal(err)
	}
	previousInClusterNamespace := inClusterNamespace
	inClusterNamespace = f.Name()
	defer func() {
		os.Remove(f.Name())
		inClusterNamespace = previousInClusterNamespace
	}()
	calls := make([]string, 0)
	mockRunFlyCmd := func(cli string, args ...string) error {
		calls = append(calls, strings.Join(append([]string{cli}, args...), " "))
		return nil
	}
	clientset := fake.NewSimpleClientset()
	systemVars := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-vars",
		},
	}
	if _, err := clientset.CoreV1().ConfigMaps("trustacks-toolchain-test").Create(context.TODO(), systemVars, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	applicationVars := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "application-test-vars",
		},
	}
	if _, err := clientset.CoreV1().ConfigMaps("trustacks-toolchain-test").Create(context.TODO(), applicationVars, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	systemSecrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-secrets",
		},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-toolchain-test").Create(context.TODO(), systemSecrets, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	applicationSecrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "application-test-secrets",
		},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-toolchain-test").Create(context.TODO(), applicationSecrets, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	sopsAgeKey := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sops-age",
		},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-toolchain-test").Create(context.TODO(), sopsAgeKey, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	concourseWeb := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "concourse-web",
		},
		Data: map[string][]byte{"local-users": []byte("test:test")},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-toolchain-test").Create(context.TODO(), concourseWeb, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := createApplication("test", "test", clientset, "test-fly", mockRunFlyCmd); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test-fly login -c http://concourse-web:8080 --username trustacks --password test", calls[0], "expected call to exist")
	assert.Equal(t, "test-fly sync", calls[1], "expected call to exist")
	assert.Equal(t, "test-fly set-team --team-name test-test --local-user trustacks --non-interactive", calls[2], "expected call to exist")
	assert.Regexp(t, `test-fly set-pipeline --team test-test -p test -c /tmp/pipeline[0-9]+ --non-interactive --load-vars-from /tmp/application-vars[0-9]+`, calls[3], "expected call to exist")
	assert.Equal(t, "test-fly unpause-pipeline -p test --team test-test", calls[4], "expected call to exist")
}

func TestGetApplicationVars(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	systemVars := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-vars",
		},
		Data: map[string]string{
			"system1": "test",
			"system2": "test",
		},
	}
	if _, err := clientset.CoreV1().ConfigMaps("trustacks-toolchain-test").Create(context.TODO(), systemVars, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	applicationVars := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "application-vars",
		},
		Data: map[string]string{
			"application1": "test",
			"application2": "test",
		},
	}
	if _, err := clientset.CoreV1().ConfigMaps("trustacks-application-test-app").Create(context.TODO(), applicationVars, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	vars, path, err := getApplicationVars("test", "app", clientset)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	assert.Contains(t, vars, "system1", "got an unexpected application var")
	assert.Contains(t, vars, "system2", "got an unexpected application var")
	assert.Contains(t, vars, "application1", "got an unexpected application var")
	assert.Contains(t, vars, "application2", "got an unexpected application var")
}

func TestGetApplicationSecrets(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	systemSecrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-secrets",
		},
		Data: map[string][]byte{
			"system1": []byte("test"),
			"system2": []byte("test"),
		},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-toolchain-test").Create(context.TODO(), systemSecrets, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	applicationSecrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "application-secrets",
		},
		Data: map[string][]byte{
			"application1": []byte("test"),
			"application2": []byte("test"),
		},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-application-test-app").Create(context.TODO(), applicationSecrets, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	secrets, err := getApplicationSecrets("test", "app", clientset)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, secrets, "system1", "got an unexpected application var")
	assert.Contains(t, secrets, "system2", "got an unexpected application var")
	assert.Contains(t, secrets, "application1", "got an unexpected application var")
	assert.Contains(t, secrets, "application2", "got an unexpected application var")
}

func TestSetAgePublicKey(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	sopsAgeSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sops-age",
		},
		Data: map[string][]byte{
			"age.agepub": []byte("age13p8qsfygta3td5yqskddxgrm62zwzekjzx0690ux46tmxqxegvqswhmrl0"),
		},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-toolchain-test").Create(context.TODO(), sopsAgeSecret, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	applicationVars := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "application-vars",
		},
		Data: map[string]string{},
	}
	if _, err := clientset.CoreV1().ConfigMaps("trustacks-application-test-app").Create(context.TODO(), applicationVars, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := setAgePublicKey("test", "app", clientset); err != nil {
		t.Fatal(err)
	}
	vars, err := clientset.CoreV1().ConfigMaps("trustacks-application-test-app").Get(context.TODO(), "application-vars", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Regexp(t, `age13p8qsfygta3td5yqskddxgrm62zwzekjzx0690ux46tmxqxegvqswhmrl0`, vars.Data["agePublicKey"], "got an unexpected age private key")
}

func TestCopyApplicationInputs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	vars := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-test-vars",
			Namespace: "trustacks-toolchain-test",
		},
		Data: map[string]string{
			"test": "value",
		},
	}
	if _, err := clientset.CoreV1().ConfigMaps("trustacks-toolchain-test").Create(context.TODO(), vars, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	secrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-test-secrets",
			Namespace: "trustacks-toolchain-test",
		},
		Data: map[string][]byte{
			"test": []byte("value"),
		},
	}
	if _, err := clientset.CoreV1().Secrets("trustacks-toolchain-test").Create(context.TODO(), secrets, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := copyApplicationInputs("test", "test", clientset); err != nil {
		t.Fatal(err)
	}
	var err error
	vars, err = clientset.CoreV1().ConfigMaps("trustacks-application-test-test").Get(context.TODO(), "application-vars", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	secrets, err = clientset.CoreV1().Secrets("trustacks-application-test-test").Get(context.TODO(), "application-secrets", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "value", vars.Data["test"], "got an unexpected variable value")
	assert.Equal(t, "value", string(secrets.Data["test"]), "got an unexpected secret value")
}
