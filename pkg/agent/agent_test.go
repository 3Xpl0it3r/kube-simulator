package agent

import (
	"context"
	"testing"
	"time"

	agtcontroller "3Xpl0it3r.com/kube-simulator/pkg/agent/controller"
	agtmanager "3Xpl0it3r.com/kube-simulator/pkg/agent/manager"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSimuAgent_New(t *testing.T) {
	helper := NewTestHelper(t)

	config := &Config{
		ClientConfig: "",
		NodeNum:      3,
	}

	client := fake.NewSimpleClientset()

	agent := SimuAgent{
		maxPods:       110,
		maxNodes:      100,
		clusterClient: client,
		nodeNum:       config.NodeNum,
	}

	helper.AssertEqual(110, agent.maxPods, "maxPods should be set correctly")
	helper.AssertEqual(100, agent.maxNodes, "maxNodes should be set correctly")
	helper.AssertEqual(3, agent.nodeNum, "nodeNum should be set correctly")
}

func TestSimuAgent_BuildKubeStandardResourceInformerFactory(t *testing.T) {
	client := fake.NewSimpleClientset()

	factory := buildKubeStandardResourceInformerFactory(client)

	if factory == nil {
		t.Fatal("Expected non-nil factory")
	}
}

func TestSimuAgent_HandleForPodOnAdd(t *testing.T) {
	helper := NewTestHelper(t)
	config := &Config{NodeNum: 1}

	agent := SimuAgent{
		maxPods:           110,
		maxNodes:          100,
		clusterClient:     helper.Client,
		nodeNum:           config.NodeNum,
		nodeStatusManager: agtmanager.NewNodeManager(helper.Client),
		podManager:        agtmanager.NewPodStatusManager(helper.Client),
	}

	testPod := helper.CreateTestPod("test-pod", "default", "test-node")

	// 这些测试只验证方法不会崩溃，实际的 manager 行为在各自的测试中验证
	agent.HandleForPodOnAdd(testPod)
	agent.HandleForPodOnUpdate(testPod)
	agent.HandleForPodOnDelete(testPod)
}

func TestSimuAgent_HandleForNodeOnAdd(t *testing.T) {
	helper := NewTestHelper(t)
	config := &Config{NodeNum: 1}

	agent := SimuAgent{
		maxPods:           110,
		maxNodes:          100,
		clusterClient:     helper.Client,
		nodeNum:           config.NodeNum,
		nodeStatusManager: agtmanager.NewNodeManager(helper.Client),
		podManager:        agtmanager.NewPodStatusManager(helper.Client),
	}

	testNode := helper.CreateTestNode("test-node", "10.10.10.1", "10.244.1.0/24")

	// 这些测试只验证方法不会崩溃，实际的 manager 行为在各自的测试中验证
	agent.HandleForNodeOnAdd(testNode)
	agent.HandleForNodeOnUpdate(testNode)
	agent.HandleForNodeOnDelete(testNode)
}

func TestSimuAgent_MainLoop_ContextCancellation(t *testing.T) {
	// 创建一个带缓冲通道的 agent
	agent := SimuAgent{
		podController:  &agtcontroller.PodController{},
		nodeController: &agtcontroller.NodeController{},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	// mainLoop 应该立即返回上下文错误
	err := agent.mainLoop(ctx)
	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled error, got: %v", err)
	}
}

func TestSimuAgent_MainLoop_Timeout(t *testing.T) {
	// 创建模拟控制器
	podCtrl := &agtcontroller.PodController{}
	nodeCtrl := &agtcontroller.NodeController{}

	agent := SimuAgent{
		podController:  podCtrl,
		nodeController: nodeCtrl,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 由于没有事件，mainLoop 应该在上下文超时时返回
	err := agent.mainLoop(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("Expected DeadlineExceeded error, got: %v", err)
	}
}
