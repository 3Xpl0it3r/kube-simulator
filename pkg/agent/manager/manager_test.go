package manager

import (
	"testing"

	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

// ManagerTestHelper 管理器测试辅助器
type ManagerTestHelper struct {
	T      *testing.T
	Client *fake.Clientset
}

// NewManagerTestHelper 创建管理器测试辅助器
func NewManagerTestHelper(t *testing.T) *ManagerTestHelper {
	return &ManagerTestHelper{
		T:      t,
		Client: fake.NewSimpleClientset(),
	}
}

// CreateTestNode 创建测试节点
func (h *ManagerTestHelper) CreateTestNode(name, ip string, cidr string) *coreapi.Node {
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

// CreateTestPod 创建测试Pod
func (h *ManagerTestHelper) CreateTestPod(name, namespace, nodeName string) *coreapi.Pod {
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

// AssertNoError 断言没有错误
func (h *ManagerTestHelper) AssertNoError(err error, msg string) {
	if err != nil {
		h.T.Fatalf("%s: %v", msg, err)
	}
}

// AssertError 断言有错误
func (h *ManagerTestHelper) AssertError(err error, msg string) {
	if err == nil {
		h.T.Fatalf("Expected error but got none: %s", msg)
	}
}

// AssertEqual 断言相等
func (h *ManagerTestHelper) AssertEqual(expected, actual interface{}, msg string) {
	if expected != actual {
		h.T.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// TestManagerInterface 验证管理器实现了 Manager 接口
func TestManagerInterface(t *testing.T) {
	helper := NewManagerTestHelper(t)

	// 验证 NodeManager 实现了 Manager 接口
	var _ Manager = NewNodeManager(helper.Client)

	// 验证 PodStatusManager 实现了 Manager 接口
	var _ Manager = NewPodStatusManager(helper.Client)
}
