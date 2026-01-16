package manager

import (
	"context"
	"testing"
	"time"

	coreapi "k8s.io/api/core/v1"
)

func TestNewPodStatusManager(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)

	if manager == nil {
		t.Fatal("Expected non-nil PodStatusManager")
	}

	if manager.clusterClient != helper.Client {
		t.Error("clusterClient not set correctly")
	}

	if manager.ipams == nil {
		t.Error("ipams not initialized")
	}

	if manager.removedQueue == nil {
		t.Error("removedQueue not initialized")
	}

	if manager.workingQueue == nil {
		t.Error("workingQueue not initialized")
	}
}

func TestPodStatusManager_OnNodeAdd(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")

	// 添加节点
	err := manager.OnNodeAdd(testNode)
	helper.AssertNoError(err, "OnNodeAdd should not return error")

	// 验证 IPAM 被创建
	if _, exists := manager.ipams[testNode.Name]; !exists {
		t.Errorf("IPAM for node %s not found", testNode.Name)
	}
}

func TestPodStatusManager_OnNodeAdd_Duplicate(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")

	// 第一次添加
	err := manager.OnNodeAdd(testNode)
	helper.AssertNoError(err, "First OnNodeAdd should not return error")

	// 第二次添加相同节点
	err = manager.OnNodeAdd(testNode)
	helper.AssertNoError(err, "Second OnNodeAdd should not return error")

	// 验证只有一个 IPAM
	if len(manager.ipams) != 1 {
		t.Errorf("Expected 1 IPAM, got %d", len(manager.ipams))
	}
}

func TestPodStatusManager_OnNodeUpdate(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")

	// 更新节点（当前实现为空操作）
	err := manager.OnNodeUpdate(testNode)
	helper.AssertNoError(err, "OnNodeUpdate should not return error")
}

func TestPodStatusManager_OnNodeDelete(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")

	// 先添加节点
	err := manager.OnNodeAdd(testNode)
	helper.AssertNoError(err, "OnNodeAdd should not return error")

	// 删除节点
	err = manager.OnNodeDelete(testNode)
	helper.AssertNoError(err, "OnNodeDelete should not return error")

	// 验证 IPAM 被删除
	if _, exists := manager.ipams[testNode.Name]; exists {
		t.Errorf("IPAM for node %s should be deleted", testNode.Name)
	}
}

func TestPodStatusManager_OnPodAdd(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")
	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 先添加节点
	err := manager.OnNodeAdd(testNode)
	helper.AssertNoError(err, "OnNodeAdd should not return error")

	// 添加 Pod
	err = manager.OnPodAdd(testPod)
	helper.AssertNoError(err, "OnPodAdd should not return error")

	// 验证工作队列收到数据
	select {
	case podFromQueue := <-manager.workingQueue:
		if podFromQueue.Name != testPod.Name {
			t.Errorf("Expected pod name %s, got %s", testPod.Name, podFromQueue.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive pod from working queue")
	}
}

func TestPodStatusManager_OnPodAdd_WithoutNode(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testPod := helper.CreateTestPod("test-pod", "default", "non-existent-node")

	// 添加 Pod 到不存在的节点
	err := manager.OnPodAdd(testPod)
	helper.AssertNoError(err, "OnPodAdd should not return error for non-existent node")

	// 验证工作队列仍然收到数据
	select {
	case podFromQueue := <-manager.workingQueue:
		if podFromQueue.Name != testPod.Name {
			t.Errorf("Expected pod name %s, got %s", testPod.Name, podFromQueue.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive pod from working queue")
	}
}

func TestPodStatusManager_OnPodUpdate(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 更新 Pod
	err := manager.OnPodUpdate(testPod)
	helper.AssertNoError(err, "OnPodUpdate should not return error")

	// 验证工作队列收到数据
	select {
	case podFromQueue := <-manager.workingQueue:
		if podFromQueue.Name != testPod.Name {
			t.Errorf("Expected pod name %s, got %s", testPod.Name, podFromQueue.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive pod from working queue")
	}
}

func TestPodStatusManager_OnPodDelete(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 删除 Pod
	err := manager.OnPodDelete(testPod)
	helper.AssertNoError(err, "OnPodDelete should not return error")

	// 验证移除队列收到数据
	select {
	case podFromQueue := <-manager.removedQueue:
		if podFromQueue.Name != testPod.Name {
			t.Errorf("Expected pod name %s, got %s", testPod.Name, podFromQueue.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive pod from removed queue")
	}
}

func TestPodStatusManager_AssignPodIP(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")
	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 先添加节点
	err := manager.OnNodeAdd(testNode)
	helper.AssertNoError(err, "OnNodeAdd should not return error")

	// 分配 IP
	manager.assignPodIP(testPod)

	// 验证 Pod 有 IP 地址
	if testPod.Status.PodIP == "" {
		t.Error("Pod should have IP assigned")
	}

	if len(testPod.Status.PodIPs) == 0 {
		t.Error("Pod should have PodIPs populated")
	}

	if testPod.Status.Phase != coreapi.PodRunning {
		t.Errorf("Expected Pod phase %s, got %s", coreapi.PodRunning, testPod.Status.Phase)
	}
}

func TestPodStatusManager_AssignPodIP_ExistingIP(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 设置现有 IP
	existingIP := "192.168.1.100"
	testPod.Status.PodIP = existingIP
	testPod.Status.PodIPs = []coreapi.PodIP{{IP: existingIP}}

	// 分配 IP
	manager.assignPodIP(testPod)

	// 验证 IP 没有改变
	if testPod.Status.PodIP != existingIP {
		t.Errorf("Expected IP %s, got %s", existingIP, testPod.Status.PodIP)
	}
}

func TestPodStatusManager_AssignPodIP_NoNodeIPAM(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testPod := helper.CreateTestPod("test-pod", "default", "non-existent-node")

	// 分配 IP（没有节点的 IPAM）
	manager.assignPodIP(testPod)

	// 验证 Pod 没有 IP 地址（因为没有节点的 IPAM）
	if testPod.Status.PodIP != "" {
		t.Errorf("Pod should not have IP when no node IPAM exists, got %s", testPod.Status.PodIP)
	}
}

func TestPodStatusManager_DeSetNetwork(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 设置网络
	testPod.Status.PodIP = "192.168.1.100"
	testPod.Status.PodIPs = []coreapi.PodIP{{IP: "192.168.1.100"}}

	// 移除网络设置
	manager.deSetNetwork(testPod)

	// 验证网络设置被移除
	if testPod.Status.PodIP != "" {
		t.Errorf("Pod IP should be empty, got %s", testPod.Status.PodIP)
	}

	if len(testPod.Status.PodIPs) != 0 {
		t.Errorf("PodIPs should be empty, got %d items", len(testPod.Status.PodIPs))
	}
}

func TestPodStatusManager_SetContainersToReadyState(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)
	testPod := &coreapi.Pod{
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{Name: "container1", Image: "nginx"},
				{Name: "container2", Image: "redis"},
			},
			InitContainers: []coreapi.Container{
				{Name: "init-container", Image: "busybox"},
			},
		},
	}

	// 设置容器为就绪状态
	manager.setContainersToReadyState(testPod)

	// 验证容器状态
	expectedContainers := len(testPod.Spec.Containers) + len(testPod.Spec.InitContainers)
	if len(testPod.Status.ContainerStatuses) != expectedContainers {
		t.Errorf("Expected %d container statuses, got %d", expectedContainers, len(testPod.Status.ContainerStatuses))
	}

	// 验证所有容器都处于运行状态
	for _, status := range testPod.Status.ContainerStatuses {
		if !status.Ready {
			t.Errorf("Container %s should be ready", status.Name)
		}

		if status.State.Running == nil {
			t.Errorf("Container %s should be in running state", status.Name)
		}
	}
}

func TestPodStatusManager_SetAllContainersTerminated(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)

	testPod := &coreapi.Pod{
		Status: coreapi.PodStatus{
			ContainerStatuses: []coreapi.ContainerStatus{
				{Name: "container1", Ready: true},
				{Name: "container2", Ready: true},
			},
			InitContainerStatuses: []coreapi.ContainerStatus{
				{Name: "init-container", Ready: true},
			},
			EphemeralContainerStatuses: []coreapi.ContainerStatus{
				{Name: "ephemeral-container", Ready: true},
			},
			Conditions: []coreapi.PodCondition{
				{Type: coreapi.PodReady, Status: coreapi.ConditionTrue},
			},
		},
	}

	// 设置所有容器为终止状态
	manager.setAllContainersTerminated(testPod)

	// 验证所有容器都处于终止状态
	allStatuses := append(testPod.Status.ContainerStatuses, testPod.Status.InitContainerStatuses...)
	allStatuses = append(allStatuses, testPod.Status.EphemeralContainerStatuses...)

	for _, status := range allStatuses {
		if status.State.Running != nil || status.State.Terminated == nil {
			t.Errorf("Container %s should be terminated", status.Name)
		}
	}

	// 验证条件状态为 False
	for _, condition := range testPod.Status.Conditions {
		if condition.Status != coreapi.ConditionFalse {
			t.Errorf("Condition %s should be False", condition.Type)
		}
	}
}

func TestPodStatusManager_Run_ContextCancellation(t *testing.T) {
	helper := NewManagerTestHelper(t)

	manager := NewPodStatusManager(helper.Client)

	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	// Run 应该在上下文取消时返回
	manager.Run(ctx)

	// 如果运行到这里而没有超时，说明正确处理了上下文取消
}
