package apiclient

import (
	"context"

	"github.com/pkg/errors"
	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func CreateOrUpdateLease(client clientset.Interface, lease *coordinationv1.Lease) error {
	existLease, err := client.CoordinationV1().Leases(lease.Namespace).Get(context.TODO(), lease.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			if _, err := client.CoordinationV1().Leases(lease.Namespace).Create(context.TODO(), lease, metav1.CreateOptions{}); err != nil {
				return errors.Wrapf(err, "unable to create lease %s:%s", lease.Namespace, lease.Name)
			}
			return nil
		}
		return errors.Wrapf(err, "unable to get lease %s:%s", lease.Namespace, lease.Name)
	}

	lease.ResourceVersion = existLease.ResourceVersion
	if _, err := client.CoordinationV1().Leases(lease.Namespace).Update(context.TODO(), lease, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "unable to update lease %s:%s", lease.Namespace, lease.Name)
	}
	return nil

}
