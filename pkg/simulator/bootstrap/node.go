package bootstrap

import (
	"fmt"
	"time"

	"3Xpl0it3r.com/kube-simulator/pkg/apiclient"
	"github.com/pkg/errors"
	coordinationv1 "k8s.io/api/coordination/v1"
	coreapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	KubeNamespaceNodeLease = "kube-node-lease"
	StaticNodePrefix       = "k8s-worker-%d"
)

func RegisterPersentNode(nodeIdx int, client kubernetes.Interface) error {
	nodeName := fmt.Sprintf(StaticNodePrefix, nodeIdx)
	podCIDR := fmt.Sprintf("10.244.%d.0/24", nodeIdx+1)
	hostIP := fmt.Sprintf("10.10.10.%d", nodeIdx+1)

	node := newNode(nodeName, hostIP, podCIDR)
	if err := apiclient.CreateOrUpdateNode(client, node); err != nil {
		return errors.Wrap(err, "unabel join new node")
	}

	if err := apiclient.CreateOrUpdateLease(client, newLeaseForNode(node.Name)); err != nil {
		return errors.Wrap(err, "unable renew lease")
	}

	return nil
}

func newLeaseForNode(name string) *coordinationv1.Lease {
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

func newNode(nodeName, nodeIp, podCIDR string) *coreapi.Node {
	node := &coreapi.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Labels: map[string]string{
				"kubernetes.io/hostname": nodeName,
			},
		},
		Spec: coreapi.NodeSpec{
			PodCIDR: podCIDR,
		},
		Status: coreapi.NodeStatus{
			Addresses: []coreapi.NodeAddress{
				{
					Type:    coreapi.NodeInternalIP,
					Address: nodeIp,
				},
				{
					Type:    coreapi.NodeHostName,
					Address: nodeName,
				},
			},
			NodeInfo: coreapi.NodeSystemInfo{
				MachineID:               "machine-id-123",
				SystemUUID:              "system-uuid-456",
				BootID:                  "boot-id-789",
				KernelVersion:           "5.4.0",
				OSImage:                 "Ubuntu 20.04",
				ContainerRuntimeVersion: fmt.Sprintf("docker://%s", nodeIp),
				KubeletVersion:          "v1.20.0",
				KubeProxyVersion:        "v1.20.0",
				OperatingSystem:         "linux",
				Architecture:            "amd64",
			},
			Capacity: coreapi.ResourceList{
				coreapi.ResourceCPU:              resource.MustParse("4"),
				coreapi.ResourceMemory:           resource.MustParse("16Gi"),
				coreapi.ResourceEphemeralStorage: resource.MustParse("100Gi"),
				coreapi.ResourcePods:             resource.MustParse("110"),
			},
			Allocatable: coreapi.ResourceList{
				coreapi.ResourceCPU:              resource.MustParse("3800m"),
				coreapi.ResourceMemory:           resource.MustParse("15.5Gi"),
				coreapi.ResourceEphemeralStorage: resource.MustParse("95Gi"),
				coreapi.ResourcePods:             resource.MustParse("110"),
			},
			Conditions: []coreapi.NodeCondition{
				{
					Type:               coreapi.NodeReady,
					Status:             coreapi.ConditionTrue,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
					Reason:             "KubeletReady",
					Message:            "kubelet is posting ready status",
				},
				{
					Type:               coreapi.NodeMemoryPressure,
					Status:             coreapi.ConditionFalse,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
					Reason:             "KubeletHasSufficientMemory",
					Message:            "kubelet has sufficient memory available",
				},
				{
					Type:               coreapi.NodeDiskPressure,
					Status:             coreapi.ConditionFalse,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
					Reason:             "KubeletHasNoDiskPressure",
					Message:            "kubelet has no disk pressure",
				},
				{
					Type:               coreapi.NodePIDPressure,
					Status:             coreapi.ConditionFalse,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
					Reason:             "KubeletHasSufficientPID",
					Message:            "kubelet has sufficient PID available",
				},
			},
		},
	}
	return node
}
