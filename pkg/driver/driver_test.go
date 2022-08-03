package driver

import (
	"context"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAddCIDriverRolebindingSubject(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-ci-driver",
			Namespace: "trustacks-application-test-test",
		},
	}
	if _, err := clientset.RbacV1().RoleBindings("trustacks-application-test-test").Create(context.TODO(), rb, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := addCIDriverServiceAccount("test", "test", "test-service-account", clientset); err != nil {
		t.Fatal(err)
	}
}
