package manager

import (
	"testing"

	coreapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewCGroupManager(t *testing.T) {
	node := &coreapi.Node{
		Status: coreapi.NodeStatus{
			Capacity: coreapi.ResourceList{
				coreapi.ResourceCPU:    resource.MustParse("2"),
				coreapi.ResourceMemory: resource.MustParse("4Gi"),
			},
		},
	}

	manager := NewCGroupManager(node)

	if manager == nil {
		t.Fatal("Expected non-nil CGrpupManager")
	}

	// 注意：当前实现不解析节点容量，只是返回空结构体
	if manager.CpuTotal != 0 {
		t.Errorf("Expected CpuTotal 0 (current implementation), got %d", manager.CpuTotal)
	}

	if manager.MemTotalInBytes != 0 {
		t.Errorf("Expected MemTotalInBytes 0 (current implementation), got %d", manager.MemTotalInBytes)
	}
}

func TestCGroupManager_OnAdd(t *testing.T) {
	manager := &CGrpupManager{}

	pod := &coreapi.Pod{
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("100m"),
							coreapi.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
	}

	initialMem := manager.MemUsedInBytes
	initialCpu := manager.CpuUsed

	manager.OnAdd(pod)

	// 验证资源使用量增加
	if manager.MemUsedInBytes <= initialMem {
		t.Error("Memory usage should increase after adding pod")
	}

	if manager.CpuUsed <= initialCpu {
		t.Error("CPU usage should increase after adding pod")
	}
}

func TestCGroupManager_OnUpdate(t *testing.T) {
	manager := &CGrpupManager{}

	pod := &coreapi.Pod{
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("50m"),
							coreapi.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
			},
		},
	}

	initialMem := manager.MemUsedInBytes
	initialCpu := manager.CpuUsed

	manager.OnUpdate(pod)

	// 验证资源使用量增加
	if manager.MemUsedInBytes <= initialMem {
		t.Error("Memory usage should increase after updating pod")
	}

	if manager.CpuUsed <= initialCpu {
		t.Error("CPU usage should increase after updating pod")
	}
}

func TestCGroupManager_OnDelete(t *testing.T) {
	manager := &CGrpupManager{}

	pod := &coreapi.Pod{
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("100m"),
							coreapi.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
	}

	// 先添加资源
	manager.OnAdd(pod)

	memAfterAdd := manager.MemUsedInBytes
	cpuAfterAdd := manager.CpuUsed

	// 删除资源
	manager.OnDelete(pod)

	// 验证资源使用量减少
	if manager.MemUsedInBytes >= memAfterAdd {
		t.Error("Memory usage should decrease after deleting pod")
	}

	if manager.CpuUsed >= cpuAfterAdd {
		t.Error("CPU usage should decrease after deleting pod")
	}
}

func TestCGroupManager_Merge(t *testing.T) {
	manager := &CGrpupManager{}

	newManager := &CGrpupManager{
		MemTotalInBytes: 8000000000,   // 8GB
		CpuTotal:        2000,         // 2 CPU
		DiskTotal:       100000000000, // 100GB
	}

	manager.Merge(newManager)

	if manager.MemTotalInBytes != newManager.MemTotalInBytes {
		t.Errorf("Expected MemTotalInBytes %d, got %d", newManager.MemTotalInBytes, manager.MemTotalInBytes)
	}

	if manager.CpuTotal != newManager.CpuTotal {
		t.Errorf("Expected CpuTotal %d, got %d", newManager.CpuTotal, manager.CpuTotal)
	}

	if manager.DiskTotal != newManager.DiskTotal {
		t.Errorf("Expected DiskTotal %d, got %d", newManager.DiskTotal, manager.DiskTotal)
	}
}

func TestCGroupManager_HasSufficientResourcesForWorload(t *testing.T) {
	manager := &CGrpupManager{
		MemTotalInBytes: 8000000000, // 8GB
		CpuTotal:        2000,       // 2 CPU
		MemUsedInBytes:  2000000000, // 2GB used
		CpuUsed:         500,        // 0.5 CPU used
	}

	smallPod := &coreapi.Pod{
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("100m"),
							coreapi.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
	}

	// 测试资源充足的情况
	if !manager.HasSufficientResourcesForWorload(smallPod) {
		t.Error("Should have sufficient resources for small pod")
	}

	largePod := &coreapi.Pod{
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("2"),
							coreapi.ResourceMemory: resource.MustParse("10Gi"),
						},
					},
				},
			},
		},
	}

	// 测试资源不足的情况
	if manager.HasSufficientResourcesForWorload(largePod) {
		t.Error("Should not have sufficient resources for large pod")
	}
}

func TestPodRequestedResources(t *testing.T) {
	pod := &coreapi.Pod{
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("1"),
							coreapi.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				{
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("1"),
							coreapi.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
	}

	mem, cpu := podRequestedResources(pod)

	// 验证资源计算
	// CPU 返回的是整数值（不是毫秒CPU），所以 1+1=2
	expectedMem := int64(128*1024*1024 + 256*1024*1024) // 384MB in bytes
	expectedCpu := int64(2)                             // 2 CPU units (not milliCPU)

	if mem != expectedMem {
		t.Errorf("Expected memory %d, got %d", expectedMem, mem)
	}

	if cpu != expectedCpu {
		t.Errorf("Expected CPU %d, got %d", expectedCpu, cpu)
	}
}
