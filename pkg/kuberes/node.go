package kuberes

import (
	coreapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNodeObject(nodeName, nodeIp, podCIDR string) *coreapi.Node {
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
				ContainerRuntimeVersion: "docker://19.3.12",
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
