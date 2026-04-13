package apiclient

import (
	"context"

	"github.com/pkg/errors"
	coreapi "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func CreateOrUpdateNode(client clientset.Interface, node *coreapi.Node) error {
	if _, err := client.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{}); err != nil {

		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "unable to create node %s", node.Name)
		}

		if _, err := client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			return errors.Wrapf(err, "unable to update node %s", node.Name)
		}
	}

	return nil

}
