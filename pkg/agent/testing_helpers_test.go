package agent

import (
	"context"
	"testing"
	"time"

	coreapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

// TestHelper 提供测试辅助函数
type TestHelper struct {
	T      *testing.T
	Client kubeclientset.Interface
}

// NewTestHelper 创建测试辅助器
func NewTestHelper(t *testing.T) *TestHelper {
	client := fake.NewSimpleClientset()
	return &TestHelper{
		T:      t,
		Client: client,
	}
}

// CreateTestNode 创建测试节点
func (h *TestHelper) CreateTestNode(name, ip string, cidr string) *coreapi.Node {
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

// CreateTestPod 创建测试 Pod
func (h *TestHelper) CreateTestPod(name, namespace, nodeName string) *coreapi.Pod {
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
					Resources: coreapi.ResourceRequirements{
						Requests: coreapi.ResourceList{
							coreapi.ResourceCPU:    resource.MustParse("100m"),
							coreapi.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
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
func (h *TestHelper) AssertNoError(err error, msg string) {
	if err != nil {
		h.T.Fatalf("%s: %v", msg, err)
	}
}

// AssertError 断言有错误
func (h *TestHelper) AssertError(err error, msg string) {
	if err == nil {
		h.T.Fatalf("Expected error but got none: %s", msg)
	}
}

// AssertEqual 断言相等
func (h *TestHelper) AssertEqual(expected, actual interface{}, msg string) {
	if expected != actual {
		h.T.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// WaitForCondition 等待条件满足
func (h *TestHelper) WaitForCondition(condition func() bool, timeout time.Duration, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.T.Fatalf("Timeout waiting for condition: %s", msg)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// MockEventRecorder 创建模拟的事件记录器
func MockEventRecorder() record.EventRecorder {
	return record.NewFakeRecorder(100)
}

// AssertChannelReceived 断言通道收到数据
func (h *TestHelper) AssertChannelReceived(ch <-chan interface{}, timeout time.Duration) {
	select {
	case <-ch:
		return
	case <-time.After(timeout):
		h.T.Fatalf("Expected to receive from channel within %v", timeout)
	}
}

// AssertChannelEmpty 断言通道为空
func (h *TestHelper) AssertChannelEmpty(ch <-chan interface{}, timeout time.Duration) {
	select {
	case <-ch:
		h.T.Fatalf("Expected channel to be empty")
	case <-time.After(timeout):
		return
	}
}

// SetupTestInformer 设置测试 informer
func SetupTestInformer() cache.Store {
	return cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
}
