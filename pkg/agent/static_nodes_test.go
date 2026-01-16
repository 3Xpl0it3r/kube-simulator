package agent

import (
	"testing"

	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRegisterBootstrapNode(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("成功创建新节点", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		node, err := registerBootstrapNode(0, client)

		helper.AssertNoError(err, "registerBootstrapNode should not return error")
		if node == nil {
			t.Fatal("Expected non-nil node")
		}

		expectedName := "mock-node-0"
		if node.Name != expectedName {
			t.Fatalf("Expected node name %s, got %s", expectedName, node.Name)
		}
	})

	t.Run("创建多个不同索引的节点", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		testCases := []struct {
			idx             int
			expectedName    string
			expectedPodCIDR string
			expectedHostIP  string
		}{
			{0, "mock-node-0", "10.244.1.0/24", "10.10.10.1"},
			{1, "mock-node-1", "10.244.2.0/24", "10.10.10.2"},
			{5, "mock-node-5", "10.244.6.0/24", "10.10.10.6"},
		}

		for _, tc := range testCases {
			t.Run(tc.expectedName, func(t *testing.T) {
				node, err := registerBootstrapNode(tc.idx, client)

				helper.AssertNoError(err, "registerBootstrapNode should not return error")
				helper.AssertEqual(tc.expectedName, node.Name, "Node name should match")
				helper.AssertEqual(tc.expectedPodCIDR, node.Spec.PodCIDR, "PodCIDR should match")

				// 验证 Host IP
				var hostIP string
				for _, addr := range node.Status.Addresses {
					if addr.Type == coreapi.NodeInternalIP {
						hostIP = addr.Address
						break
					}
				}
				helper.AssertEqual(tc.expectedHostIP, hostIP, "Host IP should match")
			})
		}
	})
}

func TestJoinNewNode(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("创建不存在的节点", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		node := &coreapi.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-new-node",
				UID:  types.UID("test-uid"),
			},
			Spec: coreapi.NodeSpec{
				PodCIDR: "10.244.1.0/24",
			},
			Status: coreapi.NodeStatus{
				Addresses: []coreapi.NodeAddress{
					{
						Type:    coreapi.NodeInternalIP,
						Address: "10.10.10.1",
					},
				},
			},
		}

		err := joinNewNode(client, node)
		helper.AssertNoError(err, "joinNewNode should not return error")
	})

	t.Run("节点已存在", func(t *testing.T) {
		existingNode := &coreapi.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "existing-node",
				UID:  types.UID("existing-uid"),
			},
		}

		client := fake.NewSimpleClientset(existingNode)

		// 尝试加入相同名称的节点
		newNode := &coreapi.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "existing-node",
				UID:  types.UID("new-uid"),
			},
		}

		err := joinNewNode(client, newNode)
		helper.AssertNoError(err, "joinNewNode should handle existing node gracefully")
	})

	t.Run("创建节点租约", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		node := &coreapi.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-with-lease",
				UID:  types.UID("lease-uid"),
			},
			Spec: coreapi.NodeSpec{
				PodCIDR: "10.244.1.0/24",
			},
			Status: coreapi.NodeStatus{
				Addresses: []coreapi.NodeAddress{
					{
						Type:    coreapi.NodeInternalIP,
						Address: "10.10.10.1",
					},
				},
			},
		}

		err := joinNewNode(client, node)
		helper.AssertNoError(err, "joinNewNode should create lease successfully")
	})
}

func TestStaticNodeConstants(t *testing.T) {
	helper := NewTestHelper(t)

	helper.AssertEqual("kube-node-lease", KubeNamespaceNodeLease, "KubeNamespaceNodeLease constant should be correct")

	testNode, err := registerBootstrapNode(0, fake.NewSimpleClientset())
	helper.AssertNoError(err, "registerBootstrapNode should not return error")
	expectedName := "mock-node-0"
	helper.AssertEqual(expectedName, testNode.Name, "StaticNodePrefix should work correctly")
}
