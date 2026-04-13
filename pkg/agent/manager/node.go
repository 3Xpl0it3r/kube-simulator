package manager

import (
	"context"
	"sync"
	"time"

	apiclient "3Xpl0it3r.com/kube-simulator/pkg/apiclient"
	coordinationv1 "k8s.io/api/coordination/v1"
	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
)

const KubeNamespaceNodeLease = "kube-node-lease"

// nodeStatus represent all formation about a node
type nodeStatus struct {
	sync.Mutex
	hostName string
	hostIp   string
	cgroup   *CGrpupManager
}

// NodeManager 管理节点的状态(节点资源信息)
type NodeManager struct {
	sync.RWMutex
	nodeStorage   map[string]*nodeStatus
	clusterClient kubeclientset.Interface
}

func NewNodeManager(client kubeclientset.Interface) *NodeManager {
	return &NodeManager{nodeStorage: make(map[string]*nodeStatus), clusterClient: client}
}

// Run [#TODO](should add some comments)
func (m *NodeManager) Run(ctx context.Context) {
	nodeSyncTick := time.NewTicker(20 * time.Second)
	defer nodeSyncTick.Stop()
	for {
		select {
		case <-nodeSyncTick.C:
			m.syncAllNodes()
		case <-ctx.Done():
			return
		}
	}
}

func (m *NodeManager) OnPodAdd(pod *coreapi.Pod) error {
	nodeName := pod.Spec.NodeName
	if nodeStatus, ok := m.nodeStorage[nodeName]; ok {
		nodeStatus.Lock()
		nodeStatus.cgroup.OnAdd(pod)
		nodeStatus.Unlock()
	} else {
	}
	return nil
}

func (m *NodeManager) OnPodUpdate(newPod *coreapi.Pod) error {
	nodeName := newPod.Spec.NodeName
	if nodeStatus, ok := m.nodeStorage[nodeName]; ok {
		nodeStatus.Lock()
		nodeStatus.cgroup.OnUpdate(newPod)
		nodeStatus.Unlock()
	}
	return nil
}

func (m *NodeManager) OnPodDelete(pod *coreapi.Pod) error {
	nodeName := pod.Spec.NodeName
	if nodeStatus, ok := m.nodeStorage[nodeName]; ok {
		nodeStatus.Lock()
		nodeStatus.cgroup.OnDelete(pod)
		nodeStatus.Unlock()
	}
	return nil
}

// 添加一个node
func (m *NodeManager) OnNodeAdd(node *coreapi.Node) error {
	m.Lock()
	defer m.Unlock()
	// if node not existed, then added
	if _, ok := m.nodeStorage[node.Name]; !ok {
		m.nodeStorage[node.Name] = nodeStatusFromNodeObj(node)
		tryResyncNodeLease(m.clusterClient, node.Name)
	}
	return nil
}

func (m *NodeManager) OnNodeUpdate(node *coreapi.Node) error {
	m.Lock()
	defer m.Unlock()
	// if node not existed, then added
	originNodeStatus, ok := m.nodeStorage[node.Name]
	if !ok {
		m.nodeStorage[node.Name] = nodeStatusFromNodeObj(node)
	}
	newNodeStatus := nodeStatusFromNodeObj(node)
	originNodeStatus.cgroup.Merge(newNodeStatus.cgroup)
	return nil
}

func (m *NodeManager) OnNodeDelete(node *coreapi.Node) error {
	nodeName := node.GetObjectMeta().GetName()
	m.Lock()
	defer m.Unlock()
	delete(m.nodeStorage, nodeName)
	return nil
}

// allNodes return all nodes name
func (m *NodeManager) allNodes() []string {
	m.RLock()
	defer m.RUnlock()
	nodes := make([]string, len(m.nodeStorage))
	for nodeName := range m.nodeStorage {
		nodes = append(nodes, nodeName)
	}
	return nodes
}

func nodeStatusFromNodeObj(node *coreapi.Node) *nodeStatus {
	var nodeInternalIp string
	for _, address := range node.Status.Addresses {
		if address.Type == coreapi.NodeInternalIP {
			nodeInternalIp = address.Address
		}
	}
	return &nodeStatus{
		cgroup:   NewCGroupManager(node),
		hostName: node.Name,
		hostIp:   nodeInternalIp,
	}
}

// syncAllNodes [#TODO](should add some comments)
func (m *NodeManager) syncAllNodes() {
	allNodes := m.allNodes()
	for _, node := range allNodes {
		tryResyncNodeLease(m.clusterClient, node)
	}
}

func tryResyncNodeLease(client kubeclientset.Interface, nodeName string) error {

	var (
		duration  int32 = 60
		renewTime       = metav1.NewMicroTime(time.Now())
	)
	lease := coordinationv1.Lease{
		Spec: coordinationv1.LeaseSpec{
			HolderIdentity:       &nodeName,
			LeaseDurationSeconds: &duration,
			RenewTime:            &renewTime,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: "kube-node-lease",
		},
	}
	return apiclient.CreateOrUpdateLease(client, &lease)
}
