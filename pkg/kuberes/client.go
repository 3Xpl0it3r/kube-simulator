package kuberes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClusterClient(masterUrl, kubeConfig string) (*kubernetes.Clientset, error) {
	restConfig, err := buildClientConfig(masterUrl, kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}

func buildClientConfig(masterUrl, kubeConfig string) (*rest.Config, error) {
	cfgLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfgLoadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	cfgLoadingRules.ExplicitPath = kubeConfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(cfgLoadingRules, &clientcmd.ConfigOverrides{})

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	if err = rest.SetKubernetesDefaults(config); err != nil {
		return nil, err
	}
	return config, nil
}
