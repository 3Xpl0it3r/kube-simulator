package simulator

import (
	"3Xpl0it3r.com/kube-simulator/pkg/simulator/bootstrap"
	clientset "k8s.io/client-go/kubernetes"
)

// kubeadm创建集群的时候会创建一些rbach和cluster-info
// `kubeadm token create`..会用到其中一些权限
// `kubeadm join` 会用到`cluster-info`配置信息
// 因此为了兼容/测试/调试kubeadm 代码，增加了这些权限
func bootstrapKubeadmCompatibleRBAC(client clientset.Interface) error {
	if err := bootstrap.AllowBoostrapTokensToGetNodes(client); err != nil {
		return err
	}
	if err := bootstrap.AllowBootstrapTokensToPostCSRs(client); err != nil {
		return err
	}
	return nil
}

// 预内置一些节点用来测试
func bootstrapOptionalPresetNodes(nodeCount int, client clientset.Interface) error {
	for idx := range nodeCount {
		if err := bootstrap.RegisterPersentNode(idx, client); err != nil {
			bootstraplogger.Errorf("unable register node %d;Err: %v", idx, err)
		} else {
			bootstraplogger.Infof("register node %d", idx)
		}
	}
	return nil

}
