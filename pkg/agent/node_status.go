package agent

import (
	"net"
	"sync"

	coreapi "k8s.io/api/core/v1"
)

// nodeStatus represent nodeinformation
type nodeStatus struct {
	sync.Mutex
	HostName string
	Ip       net.IP
	Cgroup   CGrpupManager
}

// NodeManager represent nodestatusmanager
type NodeManager struct {
	sync.RWMutex
	nodeStorage map[string]*nodeStatus
}

func NewNodeManager() *NodeManager {
	return &NodeManager{nodeStorage: make(map[string]*nodeStatus)}
}

// AllNodes [#TODO](should add some comments)
func (n *NodeManager) AllNodes() []string {
	n.RLock()
	defer n.RUnlock()
	nodes := make([]string, len(n.nodeStorage))
	for node, _ := range n.nodeStorage {
		nodes = append(nodes, node)
	}
	return nodes
}

func (n *NodeManager) onPodAdd(pod *coreapi.Pod) {
	nodeName := pod.Spec.NodeName
	if nodeStatus, ok := n.nodeStorage[nodeName]; ok {
		nodeStatus.Lock()
		nodeStatus.Cgroup.OnAdd(pod)
		nodeStatus.Unlock()
	} else {
	}
}

func (n *NodeManager) onPodUpdate(newPod *coreapi.Pod) {
	nodeName := newPod.Spec.NodeName
	if nodeStatus, ok := n.nodeStorage[nodeName]; ok {
		nodeStatus.Lock()
		nodeStatus.Cgroup.OnUpdate(newPod)
		nodeStatus.Unlock()
	}
}

func (n *NodeManager) onPodDelete(pod *coreapi.Pod) {
	nodeName := pod.Spec.NodeName
	if nodeStatus, ok := n.nodeStorage[nodeName]; ok {
		nodeStatus.Lock()
		nodeStatus.Cgroup.OnDelete(pod)
		nodeStatus.Unlock()
	}
}

// 添加一个node
func (n *NodeManager) onNodeAdd(node *coreapi.Node) {
	n.Lock()
	defer n.Unlock()
	// if node not existed, then added
	if _, ok := n.nodeStorage[node.Name]; !ok {
		n.nodeStorage[node.Name] = nodeStatusFromNodeObj(node)
		return
	}
}

func (n *NodeManager) onNodeUpdate(node *coreapi.Node) {
	n.Lock()
	defer n.Unlock()
	// if node not existed, then added
	originNodeStatus, ok := n.nodeStorage[node.Name]
	if !ok {
		n.nodeStorage[node.Name] = nodeStatusFromNodeObj(node)
		return
	}
	newNodeStatus := nodeStatusFromNodeObj(node)
	originNodeStatus.Cgroup.Merge(&newNodeStatus.Cgroup)
}

func (n *NodeManager) onNodeDelete(node *coreapi.Node) {
	nodeName := node.GetObjectMeta().GetName()
	n.Lock()
	defer n.Unlock()
	delete(n.nodeStorage, nodeName)
}

func nodeStatusFromNodeObj(node *coreapi.Node) *nodeStatus {
	return &nodeStatus{}
}
