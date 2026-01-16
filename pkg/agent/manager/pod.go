package manager

import (
	"context"
	"time"

	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
)

type PodStatusManager struct {
	ipams         map[string]*CNIPlugin
	removedQueue  chan *coreapi.Pod
	workingQueue  chan *coreapi.Pod
	clusterClient kubeclientset.Interface
}

func NewPodStatusManager(client kubeclientset.Interface) *PodStatusManager {
	pm := &PodStatusManager{
		removedQueue:  make(chan *coreapi.Pod, 1024),
		workingQueue:  make(chan *coreapi.Pod, 1024),
		clusterClient: client,
		ipams:         make(map[string]*CNIPlugin),
	}
	return pm
}

// Run [#TODO](should add some comments)
func (m *PodStatusManager) Run(ctx context.Context) {
	for {
		select {
		case pod := <-m.removedQueue:
			m.stopAllContainers(pod)
		case pod := <-m.workingQueue:
			m.startAllContainers(pod)
		case <-ctx.Done():
			return
		}
	}
}

func (m *PodStatusManager) OnPodAdd(pod *coreapi.Pod) error {
	m.startAllContainers(pod)
	m.workingQueue <- pod
	return nil
}

func (m *PodStatusManager) OnPodUpdate(pod *coreapi.Pod) error {
	m.startAllContainers(pod)
	m.workingQueue <- pod
	return nil
}

func (m *PodStatusManager) OnPodDelete(pod *coreapi.Pod) error {
	m.stopAllContainers(pod)
	m.removedQueue <- pod
	return nil
}

// OnNodeAdd [#TODO](should add some comments)
func (m *PodStatusManager) OnNodeAdd(node *coreapi.Node) error {
	_, ok := m.ipams[node.Name]
	if ok {
		return nil
	}
	if cniPlg, err := NewCNIPlugin(node.Spec.PodCIDR); err != nil {
		return err
	} else {
		m.ipams[node.Name] = cniPlg
	}
	return nil
}

func (m *PodStatusManager) OnNodeUpdate(node *coreapi.Node) error {
	return nil
}

func (m *PodStatusManager) OnNodeDelete(node *coreapi.Node) error {
	delete(m.ipams, node.Name)
	return nil
}

// startAllContainers
func (m *PodStatusManager) startAllContainers(pod *coreapi.Pod) {
	m.setContainersToReadyState(pod)
	m.assignPodIP(pod)
	m.setPodConditionStatuses(pod, true)
	pod.Status.Phase = coreapi.PodRunning
	m.updatePodStatus(pod)

}

// stopAllContainers
func (m *PodStatusManager) stopAllContainers(pod *coreapi.Pod) {
	m.setAllContainersTerminated(pod)
	m.deSetNetwork(pod)
	m.setPodConditionStatuses(pod, false)
	pod.Status.Phase = coreapi.PodSucceeded
	if m.canBeDeleted(pod) {
		m.deletePodImmediatly(pod)
	}
}

func (m *PodStatusManager) assignPodIP(pod *coreapi.Pod) {
	if len(pod.Status.PodIPs) != 0 {
		pod.Status.PodIP = pod.Status.PodIPs[0].IP
		return
	}
	ipam, ok := m.ipams[pod.Spec.NodeName]
	if !ok {
		return
	}

	if ip, err := ipam.AllocatePodIp(); err == nil {
		pod.Status.PodIP = ip
		pod.Status.PodIPs = []coreapi.PodIP{{IP: ip}}
		pod.Status.Phase = coreapi.PodRunning
	}
}

func (m *PodStatusManager) deSetNetwork(pod *coreapi.Pod) {
	pod.Status.PodIP = ""
	pod.Status.PodIPs = []coreapi.PodIP{}
}

// setContainersToReadyState simulates container starting by setting container states to Ready.
// No actual containers are running
func (m *PodStatusManager) setContainersToReadyState(pod *coreapi.Pod) {
	var (
		containersStatus []coreapi.ContainerStatus
		ready            bool = true
	)
	for _, containerSpec := range pod.Spec.InitContainers {
		containersStatus = append(containersStatus, newRunningContainerStatus(&containerSpec))
	}
	for _, containerSpec := range pod.Spec.Containers {
		containersStatus = append(containersStatus, newRunningContainerStatus(&containerSpec))
	}
	for _, containerSpec := range pod.Spec.EphemeralContainers {
		containersStatus = append(containersStatus, coreapi.ContainerStatus{
			Name:    containerSpec.Name,
			Ready:   ready,
			Started: &ready,
			Image:   containerSpec.Image,
			State: coreapi.ContainerState{
				Running: &coreapi.ContainerStateRunning{StartedAt: metav1.Now()},
			},
		})
	}
	pod.Status.ContainerStatuses = containersStatus
}

// setAllContainersTerminated simulates container stopping by setting container states to Terminated.
// No actual containers are stopped since none are running.
func (m *PodStatusManager) setAllContainersTerminated(pod *coreapi.Pod) {
	var finishedAt metav1.Time = metav1.NewTime(time.Now().Add(30 * time.Second))
	m.setContainerStatusTerminated(pod.Status.InitContainerStatuses, finishedAt)
	m.setContainerStatusTerminated(pod.Status.ContainerStatuses, finishedAt)
	m.setContainerStatusTerminated(pod.Status.EphemeralContainerStatuses, finishedAt)

	for idx := range pod.Status.Conditions {
		pod.Status.Conditions[idx].Status = coreapi.ConditionFalse
	}
}

// setContainerStatusTerminated [#TODO](should add some comments)
func (m *PodStatusManager) setContainerStatusTerminated(containerStatuses []coreapi.ContainerStatus, finishedAt metav1.Time) {
	for idx, originContainerStatus := range containerStatuses {
		if originContainerStatus.State.Terminated != nil {
			finishedAt = originContainerStatus.State.Terminated.FinishedAt
		}
		containerStatuses[idx] = newTerminatedContainerStatus(&originContainerStatus, finishedAt)
	}
}

func (m *PodStatusManager) setPodConditionStatuses(pod *coreapi.Pod, toReady bool) {
	expectedStatus := coreapi.ConditionTrue
	if !toReady {
		expectedStatus = coreapi.ConditionFalse
	}
	var now = metav1.Now()
	pod.Status.Conditions = []coreapi.PodCondition{
		{Type: coreapi.PodInitialized, LastProbeTime: now, LastTransitionTime: now, Status: expectedStatus},
		{Type: coreapi.PodReady, LastProbeTime: now, LastTransitionTime: now, Status: expectedStatus},
		{Type: coreapi.ContainersReady, LastProbeTime: now, LastTransitionTime: now, Status: expectedStatus},
		{Type: coreapi.PodScheduled, LastProbeTime: now, LastTransitionTime: now, Status: expectedStatus},
	}

}

//go:inline
func newTerminatedContainerStatus(originStatus *coreapi.ContainerStatus, finishAt metav1.Time) coreapi.ContainerStatus {
	return coreapi.ContainerStatus{
		Name:    originStatus.Name,
		Ready:   false,
		Started: nil,
		Image:   originStatus.Image,
		State:   coreapi.ContainerState{Running: nil, Waiting: nil, Terminated: &coreapi.ContainerStateTerminated{ExitCode: 0, FinishedAt: finishAt}},
	}
}

//go:inline
func newRunningContainerStatus(containerSpec *coreapi.Container) coreapi.ContainerStatus {
	var ready = true
	return coreapi.ContainerStatus{
		Name:    containerSpec.Name,
		Ready:   ready,
		Started: &ready,
		Image:   containerSpec.Image,
		State: coreapi.ContainerState{
			Running: &coreapi.ContainerStateRunning{StartedAt: metav1.Now()},
		},
	}
}

// canBeDeleted [#TODO](should add some comments)
func (m *PodStatusManager) canBeDeleted(pod *coreapi.Pod) bool {
	return true
}

// updatePodStatus [#TODO](should add some comments)
func (m *PodStatusManager) updatePodStatus(pod *coreapi.Pod) error {
	_, err := m.clusterClient.CoreV1().Pods(pod.GetNamespace()).UpdateStatus(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}

// deletePodImmediatly [#TODO](should add some comments)
func (m *PodStatusManager) deletePodImmediatly(pod *coreapi.Pod) error {
	var zero int64 = 0
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &zero,
	}
	return m.clusterClient.CoreV1().Pods(pod.GetNamespace()).Delete(context.TODO(), pod.GetName(), deleteOptions)
}
