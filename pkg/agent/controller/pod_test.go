package controller

import (
	"context"
	"testing"
	"time"

	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// PodControllerTestHelper Pod控制器测试辅助器
type PodControllerTestHelper struct {
	T      *testing.T
	Client *fake.Clientset
}

// NewPodControllerTestHelper 创建Pod控制器测试辅助器
func NewPodControllerTestHelper(t *testing.T) *PodControllerTestHelper {
	return &PodControllerTestHelper{
		T:      t,
		Client: fake.NewSimpleClientset(),
	}
}

// CreateTestPod 创建测试Pod
func (h *PodControllerTestHelper) CreateTestPod(name, namespace, nodeName string) *coreapi.Pod {
	pod := &coreapi.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID("test-pod-uid-" + name),
		},
		Spec: coreapi.PodSpec{
			NodeName: nodeName,
			Containers: []coreapi.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
		Status: coreapi.PodStatus{
			Phase: coreapi.PodPending,
		},
	}
	return pod
}

// CreateTestPodWithoutNode 创建没有节点名称的测试Pod
func (h *PodControllerTestHelper) CreateTestPodWithoutNode(name, namespace string) *coreapi.Pod {
	pod := &coreapi.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID("test-pod-uid-" + name),
		},
		Spec: coreapi.PodSpec{
			Containers: []coreapi.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
		Status: coreapi.PodStatus{
			Phase: coreapi.PodPending,
		},
	}
	return pod
}

// AssertNoError 断言没有错误
func (h *PodControllerTestHelper) AssertNoError(err error, msg string) {
	if err != nil {
		h.T.Fatalf("%s: %v", msg, err)
	}
}

// AssertEqual 断言相等
func (h *PodControllerTestHelper) AssertEqual(expected, actual interface{}, msg string) {
	if expected != actual {
		h.T.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// WaitForChannel 等待通道有数据
func (h *PodControllerTestHelper) WaitForChannel(ch <-chan PodEvent, timeout time.Duration, msg string) PodEvent {
	select {
	case event := <-ch:
		return event
	case <-time.After(timeout):
		h.T.Fatalf("Timeout waiting for channel: %s", msg)
		return PodEvent{}
	}
}

// WaitForChannelEmpty 等待通道为空
func (h *PodControllerTestHelper) WaitForChannelEmpty(ch <-chan PodEvent, timeout time.Duration) {
	select {
	case <-ch:
		h.T.Fatalf("Expected channel to be empty but received data")
	case <-time.After(timeout):
		return
	}
}

func TestNewPodController(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	if controller == nil {
		t.Fatal("Expected non-nil PodController")
	}

	if controller.clusterClient != helper.Client {
		t.Error("clusterClient not set correctly")
	}

	if controller.podInformer != podInformer {
		t.Error("podInformer not set correctly")
	}
}

func TestPodController_Channel(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	ch := controller.Chan()
	if ch == nil {
		t.Fatal("Expected non-nil channel")
	}

	// 验证通道是只读的
	helper.WaitForChannelEmpty(ch, 50*time.Millisecond)
}

func TestPodController_onAdd(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 模拟 onAdd 方法的行为
	controller.onAdd(testPod)

	// 验证通道收到事件
	event := helper.WaitForChannel(controller.Chan(), 100*time.Millisecond, "onAdd event")

	if event.Op != Added {
		t.Errorf("Expected Added operation, got %v", event.Op)
	}

	if event.Pod == nil {
		t.Error("Expected non-nil Pod in event")
	}

	if event.Pod.Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got %s", event.Pod.Name)
	}
}

func TestPodController_onAdd_WithoutNodeName(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	// 创建没有 NodeName 的 Pod
	testPod := helper.CreateTestPodWithoutNode("test-pod", "default")

	controller.onAdd(testPod)

	// 验证通道没有收到事件（没有 NodeName 应该被过滤）
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestPodController_onAdd_InvalidType(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	// 传递非Pod类型
	controller.onAdd("not a pod")

	// 验证通道没有收到事件
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestPodController_onUpdate(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	oldPod := helper.CreateTestPod("test-pod", "default", "test-node")
	oldPod.ResourceVersion = "1"
	oldPod.Status.Phase = coreapi.PodPending

	newPod := helper.CreateTestPod("test-pod", "default", "test-node")
	newPod.ResourceVersion = "2"
	newPod.Status.Phase = coreapi.PodRunning

	controller.onUpdate(oldPod, newPod)

	// 验证通道收到事件
	event := helper.WaitForChannel(controller.Chan(), 100*time.Millisecond, "onUpdate event")

	if event.Op != Update {
		t.Errorf("Expected Update operation, got %v", event.Op)
	}

	if event.Pod == nil {
		t.Error("Expected non-nil Pod in event")
	}

	if event.Pod.Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got %s", event.Pod.Name)
	}
}

func TestPodController_onUpdate_SameResourceVersion(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	oldPod := helper.CreateTestPod("test-pod", "default", "test-node")
	oldPod.ResourceVersion = "1"
	oldPod.Status.Phase = coreapi.PodPending

	newPod := helper.CreateTestPod("test-pod", "default", "test-node")
	newPod.ResourceVersion = "1" // 相同的 ResourceVersion
	newPod.Status.Phase = coreapi.PodRunning

	controller.onUpdate(oldPod, newPod)

	// 验证通道没有收到事件（相同 ResourceVersion 应该被忽略）
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestPodController_onUpdate_PodWithDeletionTimestamp(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	now := metav1.Now()

	oldPod := helper.CreateTestPod("test-pod", "default", "test-node")
	oldPod.ResourceVersion = "1"
	oldPod.Status.Phase = coreapi.PodRunning

	newPod := helper.CreateTestPod("test-pod", "default", "test-node")
	newPod.ResourceVersion = "2"
	newPod.DeletionTimestamp = &now
	newPod.Status.Phase = coreapi.PodRunning

	controller.onUpdate(oldPod, newPod)

	// 验证通道收到删除事件（有 DeletionTimestamp 应该被视为删除）
	event := helper.WaitForChannel(controller.Chan(), 100*time.Millisecond, "onUpdate deletion event")

	if event.Op != Delete {
		t.Errorf("Expected Delete operation for pod with deletion timestamp, got %v", event.Op)
	}
}

func TestPodController_onUpdate_InvalidType(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	controller.onUpdate("not old pod", "not new pod")

	// 验证通道没有收到事件
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestPodController_onDelete(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	controller.onDelete(testPod)

	// 验证通道收到事件
	event := helper.WaitForChannel(controller.Chan(), 100*time.Millisecond, "onDelete event")

	if event.Op != Delete {
		t.Errorf("Expected Delete operation, got %v", event.Op)
	}

	if event.Pod == nil {
		t.Error("Expected non-nil Pod in event")
	}

	if event.Pod.Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got %s", event.Pod.Name)
	}
}

func TestPodController_onDelete_WithoutNodeName(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	// 创建没有 NodeName 的 Pod
	testPod := helper.CreateTestPodWithoutNode("test-pod", "default")

	controller.onDelete(testPod)

	// 验证通道没有收到事件（没有 NodeName 应该被过滤）
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestPodController_onDelete_InvalidType(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	controller.onDelete("not a pod")

	// 验证通道没有收到事件
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestPodController_Run_ContextCancellation(t *testing.T) {
	helper := NewPodControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	podInformer := factory.Core().V1().Pods()

	controller := NewPodController(helper.Client, podInformer)

	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	err := controller.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled error, got: %v", err)
	}
}
