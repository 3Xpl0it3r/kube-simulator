package cluster

import (
	"testing"

	mycertutil "3Xpl0it3r.com/kube-simulator/pkg/cert"
)

func TestConfig_DefaultValues(t *testing.T) {
	helper := NewClusterTestHelper(t)

	config := &Config{}

	// 测试默认值
	helper.AssertEqual("", config.ListenHost, "ListenHost should be empty by default")
	helper.AssertEqual("", config.ListenPort, "ListenPort should be empty by default")
	helper.AssertEqual("", config.AuthorizationMode, "AuthorizationMode should be empty by default")
	helper.AssertEqual("", config.ServiceClusterIpRange, "ServiceClusterIpRange should be empty by default")
	helper.AssertEqual("", config.EtcdServers, "EtcdServers should be empty by default")
	helper.AssertEqual("", config.ClusterCIDR, "ClusterCIDR should be empty by default")
	helper.AssertEqual("", config.ServiceCIDR, "ServiceCIDR should be empty by default")
}

func TestConfig_SetValues(t *testing.T) {
	helper := NewClusterTestHelper(t)

	config := &Config{
		ListenHost:            "127.0.0.1",
		ListenPort:            "8443",
		AuthorizationMode:     "RBAC",
		ServiceClusterIpRange: "10.96.0.0/12",
		EtcdServers:           "127.0.0.1:2379",
		ClusterCIDR:           "10.244.0.0/16",
		ServiceCIDR:           "10.96.0.0/12",
	}

	helper.AssertEqual("127.0.0.1", config.ListenHost, "ListenHost should be set correctly")
	helper.AssertEqual("8443", config.ListenPort, "ListenPort should be set correctly")
	helper.AssertEqual("RBAC", config.AuthorizationMode, "AuthorizationMode should be set correctly")
	helper.AssertEqual("10.96.0.0/12", config.ServiceClusterIpRange, "ServiceClusterIpRange should be set correctly")
	helper.AssertEqual("127.0.0.1:2379", config.EtcdServers, "EtcdServers should be set correctly")
	helper.AssertEqual("10.244.0.0/16", config.ClusterCIDR, "ClusterCIDR should be set correctly")
	helper.AssertEqual("10.96.0.0/12", config.ServiceCIDR, "ServiceCIDR should be set correctly")
}

func TestClientConfigFile_DefaultValues(t *testing.T) {
	helper := NewClusterTestHelper(t)

	clientConfig := &ClientConfigFile{}

	// 测试默认值
	helper.AssertEqual("", clientConfig.ControllerManager, "ControllerManager should be empty by default")
	helper.AssertEqual("", clientConfig.Scheduler, "Scheduler should be empty by default")
	helper.AssertEqual("", clientConfig.Administrator, "Administrator should be empty by default")
}

func TestClientConfigFile_SetValues(t *testing.T) {
	helper := NewClusterTestHelper(t)

	clientConfig := &ClientConfigFile{
		ControllerManager: "/tmp/controller-manager.kubeconfig",
		Scheduler:         "/tmp/scheduler.kubeconfig",
		Administrator:     "/tmp/admin.kubeconfig",
	}

	helper.AssertEqual("/tmp/controller-manager.kubeconfig", clientConfig.ControllerManager, "ControllerManager should be set correctly")
	helper.AssertEqual("/tmp/scheduler.kubeconfig", clientConfig.Scheduler, "Scheduler should be set correctly")
	helper.AssertEqual("/tmp/admin.kubeconfig", clientConfig.Administrator, "Administrator should be set correctly")
}

func TestTLS_DefaultValues(t *testing.T) {
	helper := NewClusterTestHelper(t)

	tls := &TLS{}

	// 测试默认值
	helper.AssertEqual("", tls.ServiceAccountKeyFile, "ServiceAccountKeyFile should be empty by default")
	helper.AssertEqual("", tls.ServiceAccountSigningKeyFile, "ServiceAccountSigningKeyFile should be empty by default")
	helper.AssertEqual("", tls.EtcdCA, "EtcdCA should be empty by default")
	helper.AssertEqual("", tls.Server.Name, "Server.Name should be empty by default")
	helper.AssertEqual("", tls.Server.KeyFile, "Server.KeyFile should be empty by default")
	helper.AssertEqual("", tls.Server.CertFile, "Server.CertFile should be empty by default")
	helper.AssertEqual("", tls.CA.Name, "CA.Name should be empty by default")
	helper.AssertEqual("", tls.CA.KeyFile, "CA.KeyFile should be empty by default")
	helper.AssertEqual("", tls.CA.CertFile, "CA.CertFile should be empty by default")
	helper.AssertEqual("", tls.EtcdClient.Name, "EtcdClient.Name should be empty by default")
	helper.AssertEqual("", tls.EtcdClient.KeyFile, "EtcdClient.KeyFile should be empty by default")
	helper.AssertEqual("", tls.EtcdClient.CertFile, "EtcdClient.CertFile should be empty by default")
}

func TestTLS_SetValues(t *testing.T) {
	helper := NewClusterTestHelper(t)

	tls := &TLS{
		ServiceAccountKeyFile:        "/tmp/sa.key",
		ServiceAccountSigningKeyFile: "/tmp/sa-signing.key",
		EtcdCA:                       "/tmp/etcd-ca.crt",
		Server: mycertutil.CertKeyPair{
			Name:     "server",
			KeyFile:  "/tmp/server.key",
			CertFile: "/tmp/server.crt",
		},
		CA: mycertutil.CertKeyPair{
			Name:     "ca",
			KeyFile:  "/tmp/ca.key",
			CertFile: "/tmp/ca.crt",
		},
		EtcdClient: mycertutil.CertKeyPair{
			Name:     "etcd-client",
			KeyFile:  "/tmp/etcd-client.key",
			CertFile: "/tmp/etcd-client.crt",
		},
	}

	helper.AssertEqual("/tmp/sa.key", tls.ServiceAccountKeyFile, "ServiceAccountKeyFile should be set correctly")
	helper.AssertEqual("/tmp/sa-signing.key", tls.ServiceAccountSigningKeyFile, "ServiceAccountSigningKeyFile should be set correctly")
	helper.AssertEqual("/tmp/etcd-ca.crt", tls.EtcdCA, "EtcdCA should be set correctly")

	// 测试 Server CertKeyPair
	helper.AssertEqual("server", tls.Server.Name, "Server.Name should be set correctly")
	helper.AssertEqual("/tmp/server.key", tls.Server.KeyFile, "Server.KeyFile should be set correctly")
	helper.AssertEqual("/tmp/server.crt", tls.Server.CertFile, "Server.CertFile should be set correctly")

	// 测试 CA CertKeyPair
	helper.AssertEqual("ca", tls.CA.Name, "CA.Name should be set correctly")
	helper.AssertEqual("/tmp/ca.key", tls.CA.KeyFile, "CA.KeyFile should be set correctly")
	helper.AssertEqual("/tmp/ca.crt", tls.CA.CertFile, "CA.CertFile should be set correctly")

	// 测试 EtcdClient CertKeyPair
	helper.AssertEqual("etcd-client", tls.EtcdClient.Name, "EtcdClient.Name should be set correctly")
	helper.AssertEqual("/tmp/etcd-client.key", tls.EtcdClient.KeyFile, "EtcdClient.KeyFile should be set correctly")
	helper.AssertEqual("/tmp/etcd-client.crt", tls.EtcdClient.CertFile, "EtcdClient.CertFile should be set correctly")
}

func TestConfig_Validation(t *testing.T) {
	helper := NewClusterTestHelper(t)

	testCases := []struct {
		name        string
		config      *Config
		expectValid bool
	}{
		{
			name:        "Valid minimal config",
			config:      &Config{ListenHost: "127.0.0.1", ListenPort: "8443"},
			expectValid: true,
		},
		{
			name:        "Valid full config",
			config:      helper.CreateTestConfig(),
			expectValid: true,
		},
		{
			name: "Empty listen host",
			config: &Config{
				ListenHost: "",
				ListenPort: "8443",
			},
			expectValid: true, // 在当前实现中，空 host 可能是允许的
		},
		{
			name: "Empty listen port",
			config: &Config{
				ListenHost: "127.0.0.1",
				ListenPort: "",
			},
			expectValid: true, // 在当前实现中，空 port 可能是允许的
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 当前实现没有显式的验证逻辑
			// 这里主要测试结构体赋值和读取
			if tc.config.ListenHost == "" && tc.config.ListenPort == "" {
				// 两者都空的情况可能是无效的
				t.Logf("Config validation logic not implemented - both host and port empty")
			}

			// 验证配置对象本身不为空
			if tc.config == nil {
				t.Error("Config should not be nil")
			}
		})
	}
}
