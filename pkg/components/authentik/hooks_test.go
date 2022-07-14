package authentik

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateAPIToken(t *testing.T) {
	defer patchAPIToken()()
	if err := os.MkdirAll("/var/run/secrets/kubernetes.io/serviceaccount", 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	clientset := fake.NewSimpleClientset()
	if err := createAPIToken("test-token", clientset); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	namespace, err := getNamespace()
	if err != nil {
		t.Fatal(err)
	}
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), apiTokenSecret, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test-token", strings.TrimSpace(string(secret.Data["api-token"])), "got an unexpected token value")

	// check idempotence.
	if err := createAPIToken("test-token", clientset); err != nil {
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
	if err := createGroups(ts.URL, "test-token"); err != nil {
		t.Fatal(err)
	}
	assert.ElementsMatch(t, getGroups, []string{"admins", "editors", "viewers"})
	assert.ElementsMatch(t, postGroups, []string{"editors", "viewers"})
}
