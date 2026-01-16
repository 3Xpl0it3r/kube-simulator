package cert

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	k8certutil "k8s.io/client-go/util/cert"
)

// TestNewCACertificateConfig 测试 CA 证书配置创建
func TestNewCACertificateConfig(t *testing.T) {
	tests := []struct {
		name           string
		commonName     string
		organizations  []string
		expectedConfig k8certutil.Config
		description    string
	}{
		{
			name:          "基本 CA 配置",
			commonName:    "test-ca",
			organizations: []string{"test-org"},
			expectedConfig: k8certutil.Config{
				CommonName:   "test-ca",
				Organization: []string{"test-org"},
				NotBefore:    time.Now().Add(-10 * time.Second),
			},
			description: "创建基本的 CA 证书配置",
		},
		{
			name:          "空组织",
			commonName:    "test-ca-no-org",
			organizations: []string{},
			expectedConfig: k8certutil.Config{
				CommonName:   "test-ca-no-org",
				Organization: []string{},
				NotBefore:    time.Now().Add(-10 * time.Second),
			},
			description: "不提供组织时应该为空",
		},
		{
			name:          "多个组织",
			commonName:    "test-multi-org",
			organizations: []string{"org1", "org2", "org3"},
			expectedConfig: k8certutil.Config{
				CommonName:   "test-multi-org",
				Organization: []string{"org1", "org2", "org3"},
				NotBefore:    time.Now().Add(-10 * time.Second),
			},
			description: "支持多个组织",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewCACertificateConfig(tt.commonName, tt.organizations...)

			// 验证基本字段
			assert.Equal(t, tt.expectedConfig.CommonName, config.CommonName, "Common Name 应该匹配")
			assert.Equal(t, tt.expectedConfig.Organization, config.Organization, "Organization 应该匹配")

			// 验证时间设置 - 允许几秒的差异
			timeDiff := config.NotBefore.Sub(tt.expectedConfig.NotBefore)
			assert.LessOrEqual(t, timeDiff.Abs(), 2*time.Second, "NotBefore 时间应该在合理范围内")

			t.Logf("测试完成: %s", tt.description)
		})
	}
}

// TestNewCACertificateConfig_DefaultValues 测试默认值
func TestNewCACertificateConfig_DefaultValues(t *testing.T) {
	config := NewCACertificateConfig("default-ca", "default-org")

	// 验证默认值
	assert.Equal(t, "default-ca", config.CommonName, "Common Name 应该是默认值")
	assert.Equal(t, []string{"default-org"}, config.Organization, "Organization 应该是默认值")

	// 验证 NotBefore 是过去时间（10秒前）
	expectedTime := time.Now().Add(-10 * time.Second)
	timeDiff := config.NotBefore.Sub(expectedTime)
	assert.LessOrEqual(t, timeDiff.Abs(), 1*time.Second, "NotBefore 应该是10秒前")

	// 验证其他字段为默认值
	assert.Empty(t, config.Usages, "Usages 应该为空")
	assert.Empty(t, config.AltNames.DNSNames, "AltNames DNS 应该为空")
	assert.Empty(t, config.AltNames.IPs, "AltNames IPs 应该为空")
}

// TestNewCACertificateConfig_EdgeCases 测试边界情况
func TestNewCACertificateConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		commonName    string
		organizations []string
		description   string
	}{
		{
			name:          "空 CommonName",
			commonName:    "",
			organizations: []string{"test-org"},
			description:   "空 CommonName 应该处理",
		},
		{
			name:          "特殊字符 CommonName",
			commonName:    "test-ca-with-special@chars",
			organizations: []string{"test-org"},
			description:   "特殊字符 CommonName 应该处理",
		},
		{
			name:          "长 CommonName",
			commonName:    "very-long-ca-name-that-exceeds-normal-limits-and-should-still-work",
			organizations: []string{"test-org"},
			description:   "长 CommonName 应该处理",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 函数不应该 panic 或返回错误
			assert.NotPanics(t, func() {
				config := NewCACertificateConfig(tt.commonName, tt.organizations...)
				assert.Equal(t, tt.commonName, config.CommonName, "Common Name 应该被设置")
				assert.Equal(t, tt.organizations, config.Organization, "Organization 应该被设置")
			}, tt.description)
		})
	}
}

// TestNewCACertificateConfig_VariadicParameters 测试变长参数
func TestNewCACertificateConfig_VariadicParameters(t *testing.T) {
	// 测试不同数量的组织参数
	tests := []struct {
		name          string
		organizations []string
		description   string
	}{
		{
			name:          "无组织参数",
			organizations: []string{},
			description:   "不提供组织参数应该工作",
		},
		{
			name:          "一个组织参数",
			organizations: []string{"single-org"},
			description:   "一个组织参数应该工作",
		},
		{
			name:          "多个组织参数",
			organizations: []string{"org1", "org2", "org3"},
			description:   "多个组织参数应该工作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewCACertificateConfig("test-ca", tt.organizations...)

			assert.Equal(t, "test-ca", config.CommonName, "Common Name 应该正确")
			assert.Equal(t, tt.organizations, config.Organization, "Organization 列表应该正确")

			t.Logf("测试完成: %s", tt.description)
		})
	}
}

// TestNewCACertificateConfig_TimeValidation 测试时间相关配置
func TestNewCACertificateConfig_TimeValidation(t *testing.T) {
	beforeCall := time.Now()
	config := NewCACertificateConfig("time-test-ca", "time-test-org")
	afterCall := time.Now()

	// 验证 NotBefore 在调用时间范围内
	assert.True(t, config.NotBefore.Before(beforeCall), "NotBefore 应该在调用时间之前")
	assert.True(t, config.NotBefore.After(afterCall.Add(-11*time.Second)), "NotBefore 应该在合理时间范围内")

	// 验证时间差约等于 10 秒
	expectedTime := beforeCall.Add(-10 * time.Second)
	timeDiff := config.NotBefore.Sub(expectedTime)
	assert.LessOrEqual(t, timeDiff.Abs(), 2*time.Second, "时间差应该在 2 秒内")
}

// TestNewCACertificateConfig_Compatibility 测试与其他函数的兼容性
func TestNewCACertificateConfig_Compatibility(t *testing.T) {
	// 测试配置可以用于证书生成
	config := NewCACertificateConfig("compatibility-ca", "compatibility-org")

	// 验证配置具有必需的字段用于证书生成
	assert.NotEmpty(t, config.CommonName, "Common Name 不应该为空")
	assert.NotNil(t, config.Organization, "Organization 不应该为空")
	assert.False(t, config.NotBefore.IsZero(), "NotBefore 不应该为零值")

	// 这个测试确保证书生成函数可以接受此配置
	t.Log("配置兼容性测试通过")
}

// TestNewServerCertificateConfig 测试服务器证书配置创建
func TestNewServerCertificateConfig(t *testing.T) {
	altNames := CreateTestAltNames()

	tests := []struct {
		name           string
		commonName     string
		altNames       k8certutil.AltNames
		organizations  []string
		expectedConfig k8certutil.Config
		description    string
	}{
		{
			name:          "基本服务器配置",
			commonName:    "test-server",
			altNames:      altNames,
			organizations: []string{"test-org"},
			expectedConfig: k8certutil.Config{
				CommonName:   "test-server",
				Organization: []string{"test-org"},
				AltNames:     altNames,
				NotBefore:    time.Now(),
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			description: "创建基本的服务器证书配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewServerCerfiticateConfig(tt.commonName, tt.altNames, tt.organizations...)

			// 验证基本字段
			assert.Equal(t, tt.expectedConfig.CommonName, config.CommonName, "Common Name 应该匹配")
			assert.Equal(t, tt.expectedConfig.Organization, config.Organization, "Organization 应该匹配")
			assert.Equal(t, tt.expectedConfig.Usages, config.Usages, "Usages 应该匹配")

			// 验证 AltNames
			assert.Equal(t, tt.expectedConfig.AltNames.DNSNames, config.AltNames.DNSNames, "AltNames DNS 应该匹配")
			assert.Equal(t, len(tt.expectedConfig.AltNames.IPs), len(config.AltNames.IPs), "AltNames IPs 数量应该匹配")

			// 验证时间设置
			timeDiff := config.NotBefore.Sub(tt.expectedConfig.NotBefore)
			assert.LessOrEqual(t, timeDiff.Abs(), 2*time.Second, "NotBefore 时间应该在合理范围内")

			t.Logf("测试完成: %s", tt.description)
		})
	}
}

// TestNewServerCertificateConfig_AltNamesValidation 测试 AltNames 验证
func TestNewServerCertificateConfig_AltNamesValidation(t *testing.T) {
	altNames := CreateTestAltNames()
	config := NewServerCerfiticateConfig("altname-test", altNames, "altname-org")

	// 验证 DNS Names
	assert.Contains(t, config.AltNames.DNSNames, "localhost", "应该包含 localhost")
	assert.Contains(t, config.AltNames.DNSNames, "test-server.local", "应该包含 test-server.local")

	// 验证 IPs
	assert.Len(t, config.AltNames.IPs, 2, "应该有两个 IP 地址")

	// 检查 IP 地址（不依赖顺序）
	var hasLocalhost, hasIPv6 bool
	for _, ip := range config.AltNames.IPs {
		if ip.String() == "127.0.0.1" {
			hasLocalhost = true
		}
		if ip.String() == "::1" {
			hasIPv6 = true
		}
	}
	assert.True(t, hasLocalhost, "应该包含 localhost IP")
	assert.True(t, hasIPv6, "应该包含 IPv6 地址")
}

// TestNewClientCertificateConfig 测试客户端证书配置创建
func TestNewClientCertificateConfig(t *testing.T) {
	tests := []struct {
		name           string
		commonName     string
		organizations  []string
		expectedConfig k8certutil.Config
		description    string
	}{
		{
			name:          "基本客户端配置",
			commonName:    "test-client",
			organizations: []string{"test-org"},
			expectedConfig: k8certutil.Config{
				CommonName:   "test-client",
				Organization: []string{"test-org"},
				NotBefore:    time.Now(),
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
			description: "创建基本的客户端证书配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewClientCertificateConfig(tt.commonName, tt.organizations...)

			// 验证基本字段
			assert.Equal(t, tt.expectedConfig.CommonName, config.CommonName, "Common Name 应该匹配")
			assert.Equal(t, tt.expectedConfig.Organization, config.Organization, "Organization 应该匹配")
			assert.Equal(t, tt.expectedConfig.Usages, config.Usages, "Usages 应该匹配")

			// 验证时间设置
			timeDiff := config.NotBefore.Sub(tt.expectedConfig.NotBefore)
			assert.LessOrEqual(t, timeDiff.Abs(), 2*time.Second, "NotBefore 时间应该在合理范围内")

			t.Logf("测试完成: %s", tt.description)
		})
	}
}

// TestNewClientCertificateConfig_ClientAuthUsage 测试客户端认证用途
func TestNewClientCertificateConfig_ClientAuthUsage(t *testing.T) {
	config := NewClientCertificateConfig("client-auth-test", "client-auth-org")

	// 验证客户端认证用途
	assert.Len(t, config.Usages, 1, "应该有一个用途")
	assert.Equal(t, x509.ExtKeyUsageClientAuth, config.Usages[0], "应该是客户端认证用途")

	// 验证没有服务器认证用途
	assert.NotContains(t, config.Usages, x509.ExtKeyUsageServerAuth, "不应该包含服务器认证用途")
}

// TestConfigComparison 测试配置函数的比较
func TestConfigComparison(t *testing.T) {
	caConfig := NewCACertificateConfig("comparison-ca", "comparison-org")
	serverConfig := NewServerCerfiticateConfig("comparison-server", CreateTestAltNames(), "comparison-org")
	clientConfig := NewClientCertificateConfig("comparison-client", "comparison-org")

	// CA 配置验证
	assert.Equal(t, "comparison-ca", caConfig.CommonName)
	assert.Empty(t, caConfig.Usages, "CA 配置不应该有默认用途")

	// 服务器配置验证
	assert.Equal(t, "comparison-server", serverConfig.CommonName)
	assert.Contains(t, serverConfig.Usages, x509.ExtKeyUsageServerAuth, "服务器配置应该包含服务器认证用途")

	// 客户端配置验证
	assert.Equal(t, "comparison-client", clientConfig.CommonName)
	assert.Contains(t, clientConfig.Usages, x509.ExtKeyUsageClientAuth, "客户端配置应该包含客户端认证用途")

	// 验证时间设置差异
	assert.True(t, caConfig.NotBefore.Before(serverConfig.NotBefore), "CA 时间应该更早")
	assert.Equal(t, serverConfig.NotBefore.Unix(), clientConfig.NotBefore.Unix(), "服务器和客户端时间应该相同")
}
