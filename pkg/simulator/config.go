package simulator

import (
	"path/filepath"

	"3Xpl0it3r.com/kube-simulator/pkg/agent"
	mycertutil "3Xpl0it3r.com/kube-simulator/pkg/cert"
	"3Xpl0it3r.com/kube-simulator/pkg/cluster"
)

const (
	DefaultCertNameCA         = "etcd"
	DefaultCertNameRootCA     = "ca"
	DefaultCertNameEtcdServer = "etcd-server"
	DefaultCertNameApiServer  = "apiserver"
	DefaultCertNameEtcdClient = "apiserver-etcd"
	DefaultServiceAccountName = "service-account"

	DefaultConfKubeControllerManager = "kube-controller-manager.yml"
	DefaultConfKubeScheduler         = "kube-scheduler.yml"
	DefaultConfKubeAdmin             = "admin.conf"
)

// EtcdConfig represent etcdconfig
type EtcdConfig struct {
	DataDir    string
	Listener   string
	CACert     mycertutil.CertKeyPair
	ServerCert mycertutil.CertKeyPair
}

// Config represent config
type Config struct {
	DataDir        string
	CertificateDir string
	Etcd           EtcdConfig
	Cluster        cluster.Config
	Agent          agent.Config
}

// Complete [#TODO](should add some comments)
func (c *Config) Complete() error {
	if c.Etcd.CACert.Name == "" {
		c.Etcd.CACert.Name = DefaultCertNameCA
		c.Etcd.CACert.KeyFile = pathForKey(c.CertificateDir, DefaultCertNameCA)
		c.Etcd.CACert.CertFile = pathForCert(c.CertificateDir, DefaultCertNameCA)
		c.Cluster.TLS.EtcdCA = c.Etcd.CACert.CertFile
	}
	if c.Etcd.ServerCert.Name == "" {
		c.Etcd.ServerCert.Name = DefaultCertNameEtcdServer
		c.Etcd.ServerCert.KeyFile = pathForKey(c.CertificateDir, DefaultCertNameEtcdServer)
		c.Etcd.ServerCert.CertFile = pathForCert(c.CertificateDir, DefaultCertNameEtcdServer)
	}

	if c.Cluster.TLS.Server.Name == "" {
		c.Cluster.TLS.Server.Name = DefaultCertNameApiServer
		c.Cluster.TLS.Server.KeyFile = pathForKey(c.CertificateDir, DefaultCertNameApiServer)
		c.Cluster.TLS.Server.CertFile = pathForCert(c.CertificateDir, DefaultCertNameApiServer)
	}
	if c.Cluster.TLS.CA.Name == "" {
		c.Cluster.TLS.CA.Name = DefaultCertNameRootCA
		c.Cluster.TLS.CA.KeyFile = pathForKey(c.CertificateDir, DefaultCertNameRootCA)
		c.Cluster.TLS.CA.CertFile = pathForCert(c.CertificateDir, DefaultCertNameRootCA)
	}

	if c.Cluster.TLS.EtcdClient.Name == "" {
		c.Cluster.TLS.EtcdClient.Name = DefaultCertNameEtcdClient
		c.Cluster.TLS.EtcdClient.KeyFile = pathForKey(c.CertificateDir, DefaultCertNameEtcdClient)
		c.Cluster.TLS.EtcdClient.CertFile = pathForCert(c.CertificateDir, DefaultCertNameEtcdClient)
	}

	if c.Cluster.TLS.ServiceAccountSigningKeyFile == "" {
		c.Cluster.TLS.ServiceAccountSigningKeyFile = pathForKey(c.CertificateDir, DefaultServiceAccountName)
	}
	if c.Cluster.TLS.ServiceAccountKeyFile == "" {
		c.Cluster.TLS.ServiceAccountKeyFile = pathForPublicKey(c.CertificateDir, DefaultServiceAccountName)
	}

	if c.Cluster.ClientConfigFile.ControllerManager == "" {
		c.Cluster.ClientConfigFile.ControllerManager = filepath.Join(c.DataDir, DefaultConfKubeControllerManager)
	}
	if c.Cluster.ClientConfigFile.Scheduler == "" {
		c.Cluster.ClientConfigFile.Scheduler = filepath.Join(c.DataDir, DefaultConfKubeScheduler)
	}
	if c.Cluster.ClientConfigFile.Administrator == "" {
		c.Cluster.ClientConfigFile.Administrator = filepath.Join(DefaultConfKubeAdmin)
	}

	// for agent
	if c.Agent.ClientConfig == "" {
		c.Agent.ClientConfig = c.Cluster.ClientConfigFile.Administrator
	}

	return nil
}

func pathForKey(certDir, name string) string {
	return filepath.Join(certDir, name+".key")
}

func pathForCert(certDir, name string) string {
	return filepath.Join(certDir, name+".crt")
}

func pathForPublicKey(certDir, name string) string {
	return filepath.Join(certDir, name+".pub")
}
