package agent

import (
	"context"
	"time"

	agtcontroller "3Xpl0it3r.com/kube-simulator/pkg/agent/controller"
	agtmanager "3Xpl0it3r.com/kube-simulator/pkg/agent/manager"
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
	podController     *agtcontroller.PodController
	nodeController    *agtcontroller.NodeController
	nodeStatusManager agtmanager.Manager
	podManager        agtmanager.Manager
	maxPods           int
	maxNodes          int
	nodeNum           int
	recorder          record.EventBroadcaster
	clusterClient     kubeclientset.Interface
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
		nodeNum:       config.NodeNum,
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(2).Infof)
	eventBroadcaster.StartRecordingToSink(&coretyped.EventSinkImpl{Interface: client.CoreV1().Events(coreapi.NamespaceAll)})
	agent.recorder = eventBroadcaster

	clusterInformers := buildKubeStandardResourceInformerFactory(client)

	agent.nodeController = agtcontroller.NewNodeController(client, clusterInformers.Core().V1().Nodes())
	agent.podController = agtcontroller.NewPodController(client, clusterInformers.Core().V1().Pods())
	agent.nodeStatusManager = agtmanager.NewNodeManager(client)
	agent.podManager = agtmanager.NewPodStatusManager(client)

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
			a.nodeStatusManager.OnNodeAdd(node)
			if err := a.podManager.OnNodeAdd(node); err != nil {
				return err
			}
		}
	}

	go a.nodeController.Run(ctx)
	go a.podController.Run(ctx)
	go a.nodeStatusManager.Run(ctx)
	go a.podManager.Run(ctx)

	return a.mainLoop(ctx)
}

func (a *SimuAgent) mainLoop(ctx context.Context) error {
	for {
		select {
		case event, ok := <-a.podController.Chan():
			if !ok {
				continue
			}
			switch event.Op {
			case agtcontroller.Added:
				a.HandleForPodOnAdd(event.Pod)
			case agtcontroller.Delete:
				a.HandleForPodOnDelete(event.Pod)
			case agtcontroller.Update:
				a.HandleForPodOnUpdate(event.Pod)
			}
		case event, ok := <-a.nodeController.Chan():
			if !ok {
				continue
			}
			switch event.Op {
			case agtcontroller.Added:
				a.HandleForNodeOnAdd(event.Node)
			case agtcontroller.Delete:
				a.HandleForNodeOnDelete(event.Node)
			case agtcontroller.Update:
				a.HandleForNodeOnUpdate(event.Node)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// for pod add
func (a *SimuAgent) HandleForPodOnAdd(pod *coreapi.Pod) {
	if err := a.podManager.OnPodAdd(pod); err != nil {
		return
	}
	a.nodeStatusManager.OnPodAdd(pod)
}

// for pod update
func (a *SimuAgent) HandleForPodOnUpdate(pod *coreapi.Pod) {
	if err := a.podManager.OnPodUpdate(pod); err != nil {
		return
	}
	a.nodeStatusManager.OnPodUpdate(pod)
}

// for pod delete
func (a *SimuAgent) HandleForPodOnDelete(pod *coreapi.Pod) {
	a.podManager.OnPodDelete(pod)
}

// when node added ,first update nodeManager, then create nodelease or update nodelease if it existed
func (a *SimuAgent) HandleForNodeOnAdd(node *coreapi.Node) {
	if err := a.nodeStatusManager.OnNodeAdd(node); err != nil {
		loggerForAgent.WithError(err).Error("failed update node lease")
	}
	if err := a.podManager.OnNodeAdd(node); err != nil {
		loggerForAgent.WithError(err).Error("podmanager register node failed ")
	}
}

// for node update
func (a *SimuAgent) HandleForNodeOnUpdate(node *coreapi.Node) {
	if err := a.nodeStatusManager.OnNodeAdd(node); err != nil {
		loggerForAgent.WithError(err).Error("failed update node lease")
	}
	if err := a.podManager.OnNodeUpdate(node); err != nil {
		loggerForAgent.WithError(err).Error("podmanager update node failed ")
	}
}

// for node delete
func (a *SimuAgent) HandleForNodeOnDelete(node *coreapi.Node) {
	if err := a.nodeStatusManager.OnNodeDelete(node); err != nil {
		loggerForAgent.WithError(err).Error("failed update node lease")
	}
	if err := a.podManager.OnNodeDelete(node); err != nil {
		loggerForAgent.WithError(err).Error("podmanager delete node failed ")
	}
}

func buildKubeStandardResourceInformerFactory(kubeClient kubernetes.Interface) informers.SharedInformerFactory {
	var factoryOpts []informers.SharedInformerOption
	factoryOpts = append(factoryOpts, informers.WithNamespace(coreapi.NamespaceAll))
	factoryOpts = append(factoryOpts, informers.WithTweakListOptions(func(listOptions *metav1.ListOptions) {}))
	return informers.NewSharedInformerFactoryWithOptions(kubeClient, 5*time.Second, factoryOpts...)
}
