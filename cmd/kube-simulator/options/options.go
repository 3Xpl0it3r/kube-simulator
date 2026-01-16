package options

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"3Xpl0it3r.com/kube-simulator/pkg/simulator"
	"3Xpl0it3r.com/kube-simulator/pkg/util"
	"github.com/spf13/pflag"
)

var (
	DefaultSimulatorDir   = ".data"
	DefaultCertificateDir = filepath.Join(DefaultSimulatorDir, "pki")
	DefaultEtcdDataDir    = filepath.Join(DefaultSimulatorDir, "db")
)

type Options struct {
	ResetCluster bool
	// template workspace dir
	DataDir        string
	CertificateDir string
	ClusterListen  string

	// options for kube-apiserver
	Simulator simulator.Config
}

// NewOptions create an instance option and return
func NewOptions() *Options {
	return &Options{}
}

// Validate [#TODO](should add some comments)
func (o *Options) Validate() error {
	if o.Simulator.Etcd.CACert.CertFile == "" && o.Simulator.Etcd.CACert.KeyFile != "" {
		return errors.New("etcd ca invalid")
	}
	if o.Simulator.Etcd.CACert.CertFile != "" && o.Simulator.Etcd.CACert.KeyFile == "" {
		return errors.New("etcd ca invalid")
	}
	// if cluster cidr provided, then validate cluster cidr

	return nil
}

// Complete fill some default value to options
func (o *Options) Complete() error {
	if o.ClusterListen == "" {
		o.ClusterListen = fmt.Sprintf("%s:%d", getLocalIp(), 6443)
	}
	if host, port, err := net.SplitHostPort(o.ClusterListen); err != nil {
		return err
	} else {
		o.Simulator.Cluster.ListenHost = host
		o.Simulator.Cluster.ListenPort = port
	}
	if o.Simulator.Etcd.CACert.KeyFile != "" && o.Simulator.Etcd.CACert.CertFile != "" {
		o.Simulator.Etcd.CACert.Name = certName(o.Simulator.Etcd.CACert.KeyFile, o.Simulator.Etcd.CACert.CertFile)
		o.Simulator.Cluster.TLS.EtcdCA = o.Simulator.Etcd.CACert.CertFile
	}
	if o.Simulator.Etcd.ServerCert.KeyFile != "" && o.Simulator.Etcd.CACert.CertFile != "" {
		o.Simulator.Etcd.ServerCert.Name = certName(o.Simulator.Etcd.ServerCert.KeyFile, o.Simulator.Etcd.ServerCert.CertFile)
	}
	if o.Simulator.Cluster.TLS.CA.CertFile != "" && o.Simulator.Cluster.TLS.CA.KeyFile != "" {
		o.Simulator.Cluster.TLS.CA.Name = certName(o.Simulator.Cluster.TLS.CA.CertFile, o.Simulator.Cluster.TLS.CA.KeyFile)
	}
	if o.Simulator.Cluster.TLS.Server.KeyFile != "" && o.Simulator.Cluster.TLS.Server.CertFile != "" {
		o.Simulator.Cluster.TLS.Server.Name = certName(o.Simulator.Cluster.TLS.Server.KeyFile, o.Simulator.Cluster.TLS.Server.CertFile)
	}
	if o.Simulator.Cluster.TLS.EtcdClient.KeyFile != "" && o.Simulator.Cluster.TLS.EtcdClient.CertFile != "" {
		o.Simulator.Cluster.TLS.EtcdClient.Name = certName(o.Simulator.Cluster.TLS.EtcdClient.KeyFile, o.Simulator.Cluster.TLS.EtcdClient.CertFile)
	}

	return nil
}

func (o *Options) FlagsSets() *pflag.FlagSet {
	fs := pflag.NewFlagSet("kube-simulator", pflag.ContinueOnError)
	// global options
	fs.BoolVar(&o.ResetCluster, "reset", false, "reset cluster if cluster is already inited")

	fs.StringVar(&o.ClusterListen, "cluster-listen", "", "the address that kube-apiserver listen")
	fs.StringVar(&o.DataDir, "data-dir", DefaultSimulatorDir, "data dir")
	fs.StringVar(&o.CertificateDir, "certificate-dir", DefaultCertificateDir, "certificated dir")

	// etcd options
	fs.StringVar(&o.Simulator.Etcd.Listener, "etcd-listen", "127.0.0.1:2379", "etcd-bind")
	fs.StringVar(&o.Simulator.Etcd.CACert.KeyFile, "etcd-ca-key", "", "etcd cakey")
	fs.StringVar(&o.Simulator.Etcd.CACert.CertFile, "etcd-ca-cert", "", "etcd ca cert")
	fs.StringVar(&o.Simulator.Etcd.DataDir, "db-dir", DefaultEtcdDataDir, "the dir of db ")

	// apiserver
	fs.StringVar(&o.Simulator.Cluster.ClusterCIDR, "cluster-cidr", "10.244.0.0/16", "pod cidr")
	fs.StringVar(&o.Simulator.Cluster.ServiceCIDR, "service-cidr", "10.96.0.0/12", "service cidr")
	fs.StringVar(&o.Simulator.Cluster.TLS.CA.KeyFile, "ca-key", "", "ca key file for cluster")
	fs.StringVar(&o.Simulator.Cluster.TLS.CA.CertFile, "ca-cert", "", "ca cert file for cluster")
	fs.StringVar(&o.Simulator.Cluster.TLS.EtcdClient.KeyFile, "etcd-client-key", "", "ca key file for etcd")
	fs.StringVar(&o.Simulator.Cluster.TLS.EtcdClient.CertFile, "etcd-client-cert", "", "ca cert file for etcd")
	fs.StringVar(&o.Simulator.Cluster.TLS.Server.KeyFile, "server-cert-key", "", "apiserver cert keyfile")
	fs.StringVar(&o.Simulator.Cluster.TLS.Server.CertFile, "server-cert", "", "apiserver cert file")
	fs.StringVar(&o.Simulator.Cluster.TLS.ServiceAccountKeyFile, "service-account-priv-key", "", "")
	fs.StringVar(&o.Simulator.Cluster.TLS.ServiceAccountSigningKeyFile, "service-accont-pub-key", "", "")

	// agent
	fs.IntVar(&o.Simulator.Agent.NodeNum, "node-num", 4, "the numebr of node")

	return fs
}

// Config [#TODO](should add some comments)
func (o *Options) Config() simulator.Config {
	config := o.Simulator
	config.DataDir = o.DataDir
	config.CertificateDir = o.CertificateDir
	return config
}

func certName(keyFile, certFile string) string {
	return keyFile
}

func getLocalIp() string {
	ip, err := util.GetLocalIP()
	if err != nil {
		return "127.0.0.1"
	}
	return ip
}
