package simulator

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"

	mycertutil "3Xpl0it3r.com/kube-simulator/pkg/cert"
	"3Xpl0it3r.com/kube-simulator/pkg/cluster"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	k8certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

// bootstrapAllNecessaryClusterCertificates is response for prepare some necessary ceitificetes that needed by k8s components
// if certificated existed on disk, then  return
// if cetigicated file not existed on disk, then create new one
func bootstrapAllNecessaryClusterCertificates(config *Config) error {
	// create ca certificated for kubernetes
	if err := mycertutil.CreateCACertFiles(config.Cluster.TLS.CA, mycertutil.NewCACertificateConfig("kubernetes")); err != nil {
		return errors.Wrap(err, "create certificate file for rootca failed")
	}

	// create ca certificated for etcd
	if err := mycertutil.CreateCACertFiles(config.Etcd.CACert, mycertutil.NewCACertificateConfig("etcd-ca")); err != nil {
		return errors.Wrap(err, "create certificate file for etcd-ca failed")
	}

	kubeExtAlt := k8certutil.AltNames{
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
		},
		IPs: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP(config.Cluster.ListenHost),
		},
	}
	if err := mycertutil.CreateGenericCertFiles(config.Cluster.TLS.Server, config.Cluster.TLS.CA, mycertutil.NewServerCerfiticateConfig("kube-apiserver", kubeExtAlt, "kubernetes")); err != nil {
		return errors.Wrap(err, "create certificate file for kube-apiserver failed")
	}

	etcdExtAlt := k8certutil.AltNames{
		IPs: []net.IP{
			net.ParseIP("127.0.0.1"),
		},
	}
	if err := mycertutil.CreateGenericCertFiles(config.Etcd.ServerCert, config.Etcd.CACert, mycertutil.NewServerCerfiticateConfig("etcd-server", etcdExtAlt)); err != nil {
		return errors.Wrap(err, "create certificate file for etcd-server failed")
	}

	if err := mycertutil.CreateGenericCertFiles(config.Cluster.TLS.EtcdClient, config.Etcd.CACert, mycertutil.NewClientCertificateConfig("etcd-client")); err != nil {
		return errors.Wrap(err, "create certificate file for etcd-client failed")
	}

	if err := mycertutil.CreateServiceAccountKeyAndPublicKeyFiles(config.Cluster.TLS.ServiceAccountSigningKeyFile, config.Cluster.TLS.ServiceAccountKeyFile); err != nil {
		return errors.Wrap(err, "create ServiceAccountSigningKey failed")
	}

	return nil
}

func bootstrapComponentClusterConfigs(clusterCfg *cluster.Config) error {
	caKey, caCert, err := mycertutil.TryLoadCertAndKeyFromFile(clusterCfg.TLS.CA.KeyFile, clusterCfg.TLS.CA.CertFile)
	if err != nil {
		return errors.Wrap(err, "load ca from file")
	}
	controlPlaneURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(clusterCfg.ListenHost, clusterCfg.ListenPort),
	}
	controlPlane := controlPlaneURL.String()
	adminCertConf := mycertutil.NewClientCertificateConfig(KubeUserAdmin, KubeGroupWithAdmin)
	if err := generateClusterClientsConfig(caKey, caCert, controlPlane, adminCertConf, clusterCfg.ClientConfigFile.Administrator); err != nil {
		return errors.Wrap(err, "generate kubeconfig for admin failed")
	}

	controlManagerCertConf := mycertutil.NewClientCertificateConfig(KubeGroupWithControllerManager, KubeGroupWithDefault)
	if err := generateClusterClientsConfig(caKey, caCert, controlPlane, controlManagerCertConf, clusterCfg.ClientConfigFile.ControllerManager); err != nil {
		return errors.Wrap(err, "generate kubeconfig for controller-manager failed")
	}

	scheduleCertConf := mycertutil.NewClientCertificateConfig(KubeGroupWithScheduler, KubeGroupWithDefault)
	if err := generateClusterClientsConfig(caKey, caCert, controlPlane, scheduleCertConf, clusterCfg.ClientConfigFile.Scheduler); err != nil {
		return errors.Wrap(err, "generate kubeconfig for scheduler failed")
	}
	return nil

}

func generateClusterClientsConfig(caKey crypto.Signer, caCert *x509.Certificate, controlPlane string, clientCertConfig k8certutil.Config, fileName string) error {
	clientKey, clientCrt, err := mycertutil.NewCertAndKey(caKey, caCert, clientCertConfig)
	if err != nil {
		return errors.Wrap(err, "create new cert and key failed")
	}
	encodedClientKey, err := keyutil.MarshalPrivateKeyToPEM(clientKey)
	if err != nil {
		return err
	}
	contextName := fmt.Sprintf("%s@%s", clientCertConfig.CommonName, ClusterDefaultName)
	kubeCfg := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			ClusterDefaultName: {
				Server:                   controlPlane,
				CertificateAuthorityData: mycertutil.EncodeCertPEM(caCert),
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  ClusterDefaultName,
				AuthInfo: clientCertConfig.CommonName,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			clientCertConfig.CommonName: {
				ClientKeyData:         encodedClientKey,
				ClientCertificateData: mycertutil.EncodeCertPEM(clientCrt),
			},
		},
		CurrentContext: contextName,
	}

	return clientcmd.WriteToFile(kubeCfg, fileName)
}
