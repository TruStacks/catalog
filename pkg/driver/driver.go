package driver

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// addCIDriverServiceAccount adds the service account to the
// application's ci driver rolebinding.
func addCIDriverServiceAccount(toolchain, name, serviceAccountName string, clientset kubernetes.Interface) error {
	applicationNamespace := fmt.Sprintf("trustacks-application-%s-%s", toolchain, name)
	rolebinding, err := clientset.RbacV1().RoleBindings(applicationNamespace).Get(context.TODO(), "application-ci-driver", metav1.GetOptions{})
	if err != nil {
		return err
	}
	rolebinding.Subjects = append(rolebinding.Subjects, v1.Subject{
		Kind:      "ServiceAccount",
		Name:      serviceAccountName,
		Namespace: fmt.Sprintf("trustacks-toolchain-%s", toolchain),
	})
	data, err := json.Marshal(rolebinding)
	if err != nil {
		return err
	}
	if _, err := clientset.RbacV1().RoleBindings(applicationNamespace).Patch(context.TODO(), "application-ci-driver", types.StrategicMergePatchType, data, metav1.PatchOptions{}); err != nil {
		return err
	}
	return nil
}
