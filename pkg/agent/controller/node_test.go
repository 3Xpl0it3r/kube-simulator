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

// NodeControllerTestHelper 节点控制器测试辅助器
type NodeControllerTestHelper struct {
	T      *testing.T
	Client *fake.Clientset
}

// NewNodeControllerTestHelper 创建节点控制器测试辅助器
func NewNodeControllerTestHelper(t *testing.T) *NodeControllerTestHelper {
	return &NodeControllerTestHelper{
		T:      t,
		Client: fake.NewSimpleClientset(),
	}
}

// CreateTestNode 创建测试节点
func (h *NodeControllerTestHelper) CreateTestNode(name, ip string, cidr string) *coreapi.Node {
	node := &coreapi.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:  types.UID("test-node-uid-" + name),
		},
		Status: coreapi.NodeStatus{
			Addresses: []coreapi.NodeAddress{
				{
					Type:    coreapi.NodeInternalIP,
					Address: ip,
				},
			},
		},
		Spec: coreapi.NodeSpec{
			PodCIDR: cidr,
		},
	}
	return node
}

// AssertNoError 断言没有错误
func (h *NodeControllerTestHelper) AssertNoError(err error, msg string) {
	if err != nil {
		h.T.Fatalf("%s: %v", msg, err)
	}
}

// AssertEqual 断言相等
func (h *NodeControllerTestHelper) AssertEqual(expected, actual interface{}, msg string) {
	if expected != actual {
		h.T.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// WaitForChannel 等待通道有数据
func (h *NodeControllerTestHelper) WaitForChannel(ch <-chan NodeEvent, timeout time.Duration, msg string) NodeEvent {
	select {
	case event := <-ch:
		return event
	case <-time.After(timeout):
		h.T.Fatalf("Timeout waiting for channel: %s", msg)
		return NodeEvent{}
	}
}

// WaitForChannelEmpty 等待通道为空
func (h *NodeControllerTestHelper) WaitForChannelEmpty(ch <-chan NodeEvent, timeout time.Duration) {
	select {
	case <-ch:
		h.T.Fatalf("Expected channel to be empty but received data")
	case <-time.After(timeout):
		return
	}
}

func TestNewNodeController(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	if controller == nil {
		t.Fatal("Expected non-nil NodeController")
	}

	if controller.clusterClient != helper.Client {
		t.Error("clusterClient not set correctly")
	}

	if controller.nodeInformer != nodeInformer {
		t.Error("nodeInformer not set correctly")
	}
}

func TestNodeController_Channel(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	ch := controller.Chan()
	if ch == nil {
		t.Fatal("Expected non-nil channel")
	}

	// 验证通道是只读的
	// 我们不能写入，所以只能等待超时来验证通道行为
	helper.WaitForChannelEmpty(ch, 50*time.Millisecond)
}

func TestNodeController_onAdd(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")

	// 模拟 onAdd 方法的行为
	controller.onAdd(testNode)

	// 验证通道收到事件
	event := helper.WaitForChannel(controller.Chan(), 100*time.Millisecond, "onAdd event")

	if event.Op != Added {
		t.Errorf("Expected Added operation, got %v", event.Op)
	}

	if event.Node == nil {
		t.Error("Expected non-nil Node in event")
	}

	if event.Node.Name != "test-node" {
		t.Errorf("Expected node name 'test-node', got %s", event.Node.Name)
	}
}

func TestNodeController_onAdd_InvalidType(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	// 传递非节点类型
	controller.onAdd("not a node")

	// 验证通道没有收到事件
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestNodeController_onUpdate(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	oldNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")
	oldNode.ResourceVersion = "1"

	newNode := helper.CreateTestNode("test-node", "10.10.10.2", "10.244.1.0/24")
	newNode.ResourceVersion = "2"

	controller.onUpdate(oldNode, newNode)

	// 验证通道收到事件
	event := helper.WaitForChannel(controller.Chan(), 100*time.Millisecond, "onUpdate event")

	if event.Op != Update {
		t.Errorf("Expected Update operation, got %v", event.Op)
	}

	if event.Node == nil {
		t.Error("Expected non-nil Node in event")
	}

	if event.Node.Name != "test-node" {
		t.Errorf("Expected node name 'test-node', got %s", event.Node.Name)
	}
}

func TestNodeController_onUpdate_SameResourceVersion(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	oldNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")
	oldNode.ResourceVersion = "1"

	newNode := helper.CreateTestNode("test-node", "10.10.10.2", "10.244.1.0/24")
	newNode.ResourceVersion = "1" // 相同的 ResourceVersion

	controller.onUpdate(oldNode, newNode)

	// 验证通道没有收到事件（相同 ResourceVersion 应该被忽略）
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestNodeController_onUpdate_InvalidType(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	controller.onUpdate("not old node", "not new node")

	// 验证通道没有收到事件
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestNodeController_onDelete(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")

	controller.onDelete(testNode)

	// 验证通道收到事件
	event := helper.WaitForChannel(controller.Chan(), 100*time.Millisecond, "onDelete event")

	if event.Op != Delete {
		t.Errorf("Expected Delete operation, got %v", event.Op)
	}

	if event.Node == nil {
		t.Error("Expected non-nil Node in event")
	}

	if event.Node.Name != "test-node" {
		t.Errorf("Expected node name 'test-node', got %s", event.Node.Name)
	}
}

func TestNodeController_onDelete_InvalidType(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	controller.onDelete("not a node")

	// 验证通道没有收到事件
	helper.WaitForChannelEmpty(controller.Chan(), 50*time.Millisecond)
}

func TestNodeController_Run_ContextCancellation(t *testing.T) {
	helper := NewNodeControllerTestHelper(t)

	factory := informers.NewSharedInformerFactoryWithOptions(helper.Client, 10*time.Millisecond)
	nodeInformer := factory.Core().V1().Nodes()

	controller := NewNodeController(helper.Client, nodeInformer)

	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	err := controller.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled error, got: %v", err)
	}
}
