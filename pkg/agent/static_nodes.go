package agent

import (
	"context"
	"fmt"

	"3Xpl0it3r.com/kube-simulator/pkg/kuberes"
	"github.com/sirupsen/logrus"
	coreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var loggerForCli = logrus.WithField("component", "simulator")

const KubeNamespaceNodeLease = "kube-node-lease"
const (
	StaticNodePrefix = "mock-node-%d"
)

func registerBootstrapNode(nodeIdx int, client kubernetes.Interface) (*coreapi.Node, error) {
	nodeName := fmt.Sprintf(StaticNodePrefix, nodeIdx)
	podCIDR := fmt.Sprintf("10.244.%d.0/24", nodeIdx+1)
	hostIP := fmt.Sprintf("10.10.10.%d", nodeIdx+1)

	node := kuberes.NewNodeObject(nodeName, hostIP, podCIDR)
	if err := joinNewNode(client, node); err != nil {
		return nil, err
	}
	return node, nil
}

// create new node, if node existed in cluster, return , else create new node
func joinNewNode(client kubernetes.Interface, node *coreapi.Node) error {
	var nodeName = node.Name
	// check node existed , if node already then return with error
	if _, err := client.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{}); err != nil {
		// create new node
		if _, err := client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	// lease existed, then renew it, elase create new one
	if lease, err := client.CoordinationV1().Leases(KubeNamespaceNodeLease).Get(context.TODO(), nodeName, metav1.GetOptions{}); err == nil {
		kuberes.UpdateLease(lease)
		if _, err := client.CoordinationV1().Leases(KubeNamespaceNodeLease).Update(context.TODO(), lease, metav1.UpdateOptions{}); err != nil {
			loggerForCli.WithError(err).Warningf("update lease for node %s failed", nodeName)
		}
		return nil
	}

	// create new lease for this ndde
	if _, err := client.CoordinationV1().Leases(KubeNamespaceNodeLease).Create(context.TODO(), kuberes.NewLeaseObject(nodeName), metav1.CreateOptions{}); err != nil {
		// if create faild ,return error directory, don't need delete node that create before for reback
		// for we will resync node with lease in health check loop
		loggerForCli.WithError(err).Warningf("create lease for node %s failed", nodeName)
		return nil
	}

	return nil
}
