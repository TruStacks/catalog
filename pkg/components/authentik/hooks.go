package authentik

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// preInstall creates the authentik admin api token.
func (c *authentik) preInstall() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	log.Printf("create admin api token")
	res, err := password.Generate(32, 10, 0, false, false)
	if err != nil {
		return err
	}
	if err := createAPIToken(res, clientset); err != nil {
		return err
	}
	return nil
}

// postInstall creates the authentik user groups.
func (c *authentik) postInstall() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	log.Println("create authentik user groups")
	token, err := getAPIToken(clientset)
	if err != nil {
		return err
	}
	if err := createGroups(serviceURL, token); err != nil {
		return err
	}
	return nil
}

// createAPIToken creates the api token secret.
func createAPIToken(token string, clientset kubernetes.Interface) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiTokenSecret,
		},
		Data: map[string][]byte{
			"api-token": []byte(token),
		},
	}
	namespace, err := getNamespace()
	if err != nil {
		return err
	}
	_, err = clientset.CoreV1().Secrets(namespace).Get(context.TODO(), apiTokenSecret, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_, err = clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
			return err
		}
		if !strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	return nil
}

// group represents an authentik group.
type group struct {
	Name        string `json:"name"`
	Users       []int  `json:"users"`
	IsSuperuser bool   `json:"is_superuser"`
	Parent      *int   `json:"parent"`
}

// createGroups creates the user groups.
func createGroups(url, token string) error {
	groups := []group{
		{"admins", []int{1}, true, nil},
		{"editors", []int{}, false, nil},
		{"viewers", []int{}, false, nil},
	}
	for _, g := range groups {
		data, err := json.Marshal(g)
		if err != nil {
			return err
		}
		// check if the group already exists.
		resp, err := getAPIResource(url, "core/groups", token, fmt.Sprintf("name=%s", g.Name))
		if err != nil {
			return err
		}
		results := make(map[string]interface{})
		if err := json.Unmarshal(resp, &results); err != nil {
			return err
		}
		if len(results["results"].([]interface{})) > 0 {
			continue
		}
		_, err = postAPIResource(url, "core/groups", token, data)
		if err != nil {
			return err
		}
	}
	return nil
}
