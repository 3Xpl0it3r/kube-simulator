package kuberes

import (
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewLeaseObject(name string) *coordinationv1.Lease {
	var (
		duration  int32 = 40
		renewTime       = metav1.NewMicroTime(time.Now())
	)
	return &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-node-lease",
		},
		Spec: coordinationv1.LeaseSpec{
			HolderIdentity:       &name,
			LeaseDurationSeconds: &duration,
			RenewTime:            &renewTime,
		},
	}
}

func UpdateLease(originLease *coordinationv1.Lease) {
	originLease.Spec.RenewTime = &metav1.MicroTime{Time: time.Now()}
}
