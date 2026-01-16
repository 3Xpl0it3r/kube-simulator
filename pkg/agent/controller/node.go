package controller

import (
	"context"

	coreapi "k8s.io/api/core/v1"
	coreinformer "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)


// NodeEvent represent nodeevent
type NodeEvent struct {
	Node *coreapi.Node
	Op   EventOp
}

// NodeController represent nodemanager
type NodeController struct {
	clusterClient kubeclientset.Interface
	recorder      record.EventRecorder
	nodeInformer  coreinformer.NodeInformer
	nodeCh        chan NodeEvent
}

func NewNodeController(client kubeclientset.Interface, nodeInformer coreinformer.NodeInformer) *NodeController {
	manager := &NodeController{
		clusterClient: client,
		nodeInformer:  nodeInformer,
		nodeCh:        make(chan NodeEvent, DefaultEventBufferSize),
	}
	return manager
}

func (n *NodeController) Run(ctx context.Context) error {
	n.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			n.onAdd(obj)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			n.onUpdate(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			n.onDelete(obj)
		},
	})
	n.nodeInformer.Informer().Run(ctx.Done())

	<-ctx.Done()
	return ctx.Err()
}

func (n *NodeController) Chan() <-chan NodeEvent {
	return n.nodeCh
}

func (n *NodeController) onAdd(obj interface{}) {
	if node, ok := obj.(*coreapi.Node); ok {
		n.nodeCh <- NodeEvent{Op: Added, Node: node}
	}
}

func (n *NodeController) onUpdate(oldObj, newObj interface{}) {
	oldNode, ok := oldObj.(*coreapi.Node)
	if !ok {
		return
	}
	newNode, ok := newObj.(*coreapi.Node)
	if !ok {
		return
	}
	if oldNode.GetResourceVersion() == newNode.GetResourceVersion() {
		return
	}
	n.nodeCh <- NodeEvent{Op: Update, Node: newNode}
}

func (n *NodeController) onDelete(obj interface{}) {
	if node, ok := obj.(*coreapi.Node); ok {
		n.nodeCh <- NodeEvent{Op: Delete, Node: node}
	}
}
