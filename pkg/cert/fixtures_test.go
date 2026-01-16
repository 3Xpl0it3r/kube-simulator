package cert

import (
	"crypto/x509"
	"net"
	"time"

	k8certutil "k8s.io/client-go/util/cert"
)

// 测试常量
const (
	// 测试证书名称
	TestCACommonName     = "test-ca"
	TestServerCommonName = "test-server"
	TestClientCommonName = "test-client"
	TestNamespace        = "test-namespace"

	// 测试组织
	TestOrganization = "test-org"

	// 测试密钥大小
	TestRSAKeySize = 2048

	// 测试时间范围
	TestCertDuration = time.Hour * 24 * 365 // 1年

	// 测试文件路径
	TestCAKeyFile             = "ca.key"
	TestCACertFile            = "ca.crt"
	TestServerKeyFile         = "server.key"
	TestServerCertFile        = "server.crt"
	TestClientKeyFile         = "client.key"
	TestClientCertFile        = "client.crt"
	TestServiceAccountKeyFile = "sa.key"
	TestServiceAccountPubFile = "sa.pub"
)

// 测试数据结构
type TestCertificates struct {
	CA             TestCertPair
	Server         TestCertPair
	Client         TestCertPair
	ServiceAccount TestServiceAccount
}

type TestCertPair struct {
	KeyFile    string
	CertFile   string
	CommonName string
}

type TestServiceAccount struct {
	KeyFile string
	PubFile string
}

// CreateTestAltNames 创建测试用的 AltNames
func CreateTestAltNames() k8certutil.AltNames {
	return k8certutil.AltNames{
		DNSNames: []string{"localhost", "test-server.local"},
		IPs:      createTestIPs(),
	}
}

// createTestIPs 创建测试 IP 地址
func createTestIPs() []net.IP {
	ipStrings := []string{"127.0.0.1", "::1"}
	ips := make([]net.IP, len(ipStrings))
	for i, ipStr := range ipStrings {
		ips[i] = net.ParseIP(ipStr)
	}
	return ips
}

// GetTestCertificates 获取标准测试证书配置
func GetTestCertificates() TestCertificates {
	return TestCertificates{
		CA: TestCertPair{
			KeyFile:    TestCAKeyFile,
			CertFile:   TestCACertFile,
			CommonName: TestCACommonName,
		},
		Server: TestCertPair{
			KeyFile:    TestServerKeyFile,
			CertFile:   TestServerCertFile,
			CommonName: TestServerCommonName,
		},
		Client: TestCertPair{
			KeyFile:    TestClientKeyFile,
			CertFile:   TestClientCertFile,
			CommonName: TestClientCommonName,
		},
		ServiceAccount: TestServiceAccount{
			KeyFile: TestServiceAccountKeyFile,
			PubFile: TestServiceAccountPubFile,
		},
	}
}

// GetTestCACertificateConfig 获取测试 CA 配置
func GetTestCACertificateConfig() k8certutil.Config {
	return k8certutil.Config{
		CommonName:   TestCACommonName,
		Organization: []string{TestOrganization},
		NotBefore:    time.Now().Add(-time.Minute),
	}
}

// GetTestServerCertificateConfig 获取测试服务器证书配置
func GetTestServerCertificateConfig() k8certutil.Config {
	return k8certutil.Config{
		CommonName:   TestServerCommonName,
		Organization: []string{TestOrganization},
		AltNames:     CreateTestAltNames(),
		NotBefore:    time.Now(),
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
}

// GetTestClientCertificateConfig 获取测试客户端证书配置
func GetTestClientCertificateConfig() k8certutil.Config {
	return k8certutil.Config{
		CommonName:   TestClientCommonName,
		Organization: []string{TestOrganization},
		NotBefore:    time.Now(),
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}
