package agent

import (
	"context"

	coreapi "k8s.io/api/core/v1"
	coreinformer "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type EventOp int

const (
	Added EventOp = iota
	Update
	Delete
	Unknown
)

// PodEvent represent podevent
type PodEvent struct {
	Pod *coreapi.Pod
	Op  EventOp
}

// PodController represent podmanager
type PodController struct {
	podInformer   coreinformer.PodInformer
	clusterClient kubeclientset.Interface
	recorder      record.EventRecorder
	podEventCh    chan PodEvent
}

func NewPodController(client kubeclientset.Interface, podInformer coreinformer.PodInformer) *PodController {
	manager := &PodController{
		clusterClient: client,
		podInformer:   podInformer,
		podEventCh:    make(chan PodEvent),
	}

	return manager
}

func (p *PodController) Run(ctx context.Context) error {
	p.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			p.onAdd(obj)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			p.onUpdate(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			p.onDelete(obj)
		},
	})
	p.podInformer.Informer().Run(ctx.Done())

	<-ctx.Done()
	return ctx.Err()
}

func (p *PodController) Chan() <-chan PodEvent {
	return p.podEventCh
}

func (p *PodController) onAdd(obj interface{}) {
	if pod, ok := obj.(*coreapi.Pod); ok {
		p.podEventCh <- PodEvent{Op: Added, Pod: pod}
	}
}
func (p *PodController) onDelete(obj interface{}) {
	if pod, ok := obj.(*coreapi.Pod); ok {
		p.podEventCh <- PodEvent{Op: Added, Pod: pod}
	}
}
func (p *PodController) onUpdate(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*coreapi.Pod)
	if !ok {
		return
	}
	newPod, ok := newObj.(*coreapi.Pod)
	if !ok {
		return
	}
	if oldPod.GetResourceVersion() == newPod.GetResourceVersion() {
		return
	}
	p.podEventCh <- PodEvent{Op: Update, Pod: newPod}
}
