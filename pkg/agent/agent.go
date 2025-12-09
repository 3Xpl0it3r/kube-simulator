package agent

import (
	"context"
	"time"

	kuberesource "3Xpl0it3r.com/kube-simulator/pkg/kuberes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kubeclientset "k8s.io/client-go/kubernetes"
	coretyped "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

var loggerForAgent = logrus.WithField("component", "agent")

// SimuAgent represent SimuAgent
type SimuAgent struct {
	podController  *PodController
	nodeController *NodeController
	nodeManager    *NodeManager
	maxPods        int
	maxNodes       int
	nodeNum        int
	recorder       record.EventBroadcaster
	clusterClient  kubeclientset.Interface
}

func Run(config *Config) error {
	client, err := kuberesource.NewClusterClient("", config.ClientConfig)
	if err != nil {
		return errors.Wrap(err, "build clientconfig for agent failed")
	}
	agent := SimuAgent{
		maxPods:       110,
		maxNodes:      100,
		clusterClient: client,
		nodeManager:   NewNodeManager(),
		nodeNum:       config.NodeNum,
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(2).Infof)
	eventBroadcaster.StartRecordingToSink(&coretyped.EventSinkImpl{Interface: client.CoreV1().Events(coreapi.NamespaceAll)})
	agent.recorder = eventBroadcaster

	clusterInformers := buildKubeStandardResourceInformerFactory(client)

	agent.nodeController = NewNodeController(client, clusterInformers.Core().V1().Nodes())
	agent.podController = NewPodController(client, clusterInformers.Core().V1().Pods())

	go func() {
		loggerForAgent.Info("begin run simu-agent")
		loggerForAgent.Fatalf("run agent faild %s", agent.run())
	}()
	return nil
}

func (a *SimuAgent) run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for idx := 0; idx < a.nodeNum; idx++ {
		if node, err := registerBootstrapNode(idx, a.clusterClient); err != nil {
			return err
		} else {
			a.nodeManager.onNodeAdd(node)
		}
	}

	go a.nodeController.Run(ctx)
	go a.podController.Run(ctx)

	return a.syncLoop(ctx)
}

func (a *SimuAgent) syncLoop(ctx context.Context) error {
	nodeSyncTicker := time.NewTicker(5 * time.Second)
	defer nodeSyncTicker.Stop()
	for {
		select {
		case event, ok := <-a.podController.Chan():
			if !ok {
				continue
			}
			switch event.Op {
			case Added:
				a.HandleForPodOnAdd(event.Pod)
			case Delete:
				a.HandleForPodOnDelete(event.Pod)
			case Update:
				a.HandleForPodOnUpdate(event.Pod)
			}
		case event, ok := <-a.nodeController.Chan():
			if !ok {
				continue
			}
			switch event.Op {
			case Added:
				a.HandleForNodeOnAdd(event.Node)
			case Delete:
				a.HandleForNodeOnDelete(event.Node)
			case Update:
				a.HandleForNodeOnUpdate(event.Node)
			}
		case <-nodeSyncTicker.C:
			a.SyncNodeLease()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// MethodName [#TODO](should add some comments)
func (a *SimuAgent) SyncNodeLease() {
	allNodes := a.nodeManager.AllNodes()
	for _, node := range allNodes {
		tryResyncNodeLease(a.clusterClient, node)
	}
}

func (a *SimuAgent) HandleForPodOnAdd(pod *coreapi.Pod) {
	a.nodeManager.onPodAdd(pod)
}

func (a *SimuAgent) HandleForPodOnUpdate(newPod *coreapi.Pod) {
	a.nodeManager.onPodUpdate(newPod)
}

func (a *SimuAgent) HandleForPodOnDelete(pod *coreapi.Pod) {
	a.nodeManager.onPodDelete(pod)
}

func (a *SimuAgent) HandleForNodeOnAdd(node *coreapi.Node) {
	a.nodeManager.onNodeAdd(node)
	if err := tryResyncNodeLease(a.clusterClient, node.Name); err != nil {
		loggerForAgent.WithError(err).Error("failed update node lease")
	}
}

func (a *SimuAgent) HandleForNodeOnUpdate(node *coreapi.Node) {
	a.nodeManager.onNodeUpdate(node)
	if err := tryResyncNodeLease(a.clusterClient, node.Name); err != nil {
		loggerForAgent.WithError(err).Error("failed update node lease")
	}
}

func (a *SimuAgent) HandleForNodeOnDelete(node *coreapi.Node) {
	a.nodeManager.onNodeDelete(node)
	if err := tryResyncNodeLease(a.clusterClient, node.Name); err != nil {
		loggerForAgent.WithError(err).Error("failed update node lease")
	}
}

func buildKubeStandardResourceInformerFactory(kubeClient kubernetes.Interface) informers.SharedInformerFactory {
	var factoryOpts []informers.SharedInformerOption
	factoryOpts = append(factoryOpts, informers.WithNamespace(coreapi.NamespaceAll))
	factoryOpts = append(factoryOpts, informers.WithTweakListOptions(func(listOptions *metav1.ListOptions) {}))
	return informers.NewSharedInformerFactoryWithOptions(kubeClient, 5*time.Second, factoryOpts...)
}

func tryResyncNodeLease(client kubeclientset.Interface, nodeName string) error {
	if originlease, err := client.CoordinationV1().Leases(KubeNamespaceNodeLease).Get(context.TODO(), nodeName, metav1.GetOptions{}); err == nil {
		kuberesource.UpdateLease(originlease)
		_, err = client.CoordinationV1().Leases(KubeNamespaceNodeLease).Update(context.TODO(), originlease, metav1.UpdateOptions{})
		return err
	}
	newLease := kuberesource.NewLeaseObject(nodeName)
	_, err := client.CoordinationV1().Leases(KubeNamespaceNodeLease).Create(context.TODO(), newLease, metav1.CreateOptions{})
	return err
}
