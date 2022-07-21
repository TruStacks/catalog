package inputs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	// systemVarsConfigMapName is the name of the variables config
	// map.
	systemVarsConfigMapName = "system-vars"
	// systemSecretsSecretName is the name of the secrets secret.
	systemSecretsSecretName = "system-secrets"
)

// AddSystemVars adds the component variables to the system vars
// config map.
func AddSystemVars(component, namespace string, vars map[string]string, clientset kubernetes.Interface) error {
	client := clientset.CoreV1().ConfigMaps(namespace)
	// Add the component prefix to the variables.
	data := map[string]string{}
	for k, v := range vars {
		data[fmt.Sprintf("%s.%s", component, k)] = v
	}
	// Check if the config map exists and create if not.
	if _, err := client.Get(context.TODO(), systemVarsConfigMapName, metav1.GetOptions{}); err != nil {
		if strings.Contains(err.Error(), "not found") {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: systemVarsConfigMapName,
				},
				Data: data,
			}
			if _, err := client.Create(context.TODO(), configMap, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
		return nil
	}
	// patch the existing config map.
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	patch := []byte(fmt.Sprintf(`{"data": %s}`, dataJSON))
	_, err = client.Patch(context.TODO(), systemVarsConfigMapName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

// AddSystemSecrets adds the component secrets to the system secrets
// secret.
func AddSystemSecrets(component, namespace string, secrets map[string][]byte, clientset kubernetes.Interface) error {
	client := clientset.CoreV1().Secrets(namespace)
	// Add the component prefix to the variables.
	data := map[string][]byte{}
	for k, v := range secrets {
		data[fmt.Sprintf("%s.%s", component, k)] = v
	}
	// Check if the secret exists and create if not.
	if _, err := client.Get(context.TODO(), systemSecretsSecretName, metav1.GetOptions{}); err != nil {
		if strings.Contains(err.Error(), "not found") {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: systemSecretsSecretName,
				},
				Data: data,
			}
			if _, err := client.Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
		return nil
	}
	// patch the existing secret.
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	patch := []byte(fmt.Sprintf(`{"data": %s}`, dataJSON))
	_, err = client.Patch(context.TODO(), systemSecretsSecretName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
