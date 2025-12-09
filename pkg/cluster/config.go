package cluster

import (
	mycertutil "3Xpl0it3r.com/kube-simulator/pkg/cert"
)

// Config represent config
type Config struct {
	ListenHost            string
	ListenPort            string
	AuthorizationMode     string
	ServiceClusterIpRange string
	EtcdServers           string
	ClusterCIDR           string
	ServiceCIDR           string
	ClientConfigFile      ClientConfigFile
	TLS                   TLS
}

type ClientConfigFile struct {
	ControllerManager string
	Scheduler         string
	Administrator     string
}

type TLS struct {
	ServiceAccountKeyFile        string
	ServiceAccountSigningKeyFile string
	Server                       mycertutil.CertKeyPair
	CA                           mycertutil.CertKeyPair
	EtcdCA                       string
	EtcdClient                   mycertutil.CertKeyPair
}
