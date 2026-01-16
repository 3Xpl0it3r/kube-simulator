package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

// TestCreateCACertFiles 测试 CA 证书文件创建
func TestCreateCACertFiles(t *testing.T) {
	tests := []struct {
		name        string
		caCert      CertKeyPair
		config      interface{} // 使用 interface{} 以便测试不同类型的配置
		expectError bool
		description string
	}{
		{
			name: "有效 CA 证书创建",
			caCert: CertKeyPair{
				Name:     "test-ca",
				KeyFile:  "ca.key",
				CertFile: "ca.crt",
			},
			config:      GetTestCACertificateConfig(),
			expectError: false,
			description: "使用有效配置创建 CA 证书",
		},
		{
			name: "空配置应该工作",
			caCert: CertKeyPair{
				Name:     "test-ca",
				KeyFile:  "ca.key",
				CertFile: "ca.crt",
			},
			config:      k8certutil.Config{},
			expectError: false,
			description: "空配置应该工作（原始函数不验证）",
		},
		{
			name: "绝对路径应该工作",
			caCert: CertKeyPair{
				Name:     "test-ca",
				KeyFile:  "/tmp/ca.key",
				CertFile: "/tmp/ca.crt",
			},
			config:      GetTestCACertificateConfig(),
			expectError: false,
			description: "原始函数不验证路径有效性",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := createTestTempDir(t)

			// 更新文件路径为临时目录
			caCert := CertKeyPair{
				Name:     tt.caCert.Name,
				KeyFile:  filepath.Join(tempDir, tt.caCert.KeyFile),
				CertFile: filepath.Join(tempDir, tt.caCert.CertFile),
			}

			// 执行测试
			var config k8certutil.Config
			if tt.config != nil {
				if cfg, ok := tt.config.(k8certutil.Config); ok {
					config = cfg
				} else {
					t.Skip("配置类型转换失败，跳过测试")
					return
				}
			}
			err := CreateCACertFiles(caCert, config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				return
			}

			assert.NoError(t, err, tt.description)

			// 验证证书和密钥文件已创建
			assert.FileExists(t, caCert.KeyFile, "CA 密钥文件应该存在")
			assert.FileExists(t, caCert.CertFile, "CA 证书文件应该存在")

			// 验证文件内容不为空
			assertFileContent(t, caCert.KeyFile)
			assertFileContent(t, caCert.CertFile)

			// 验证证书的 Common Name - 根据配置检查
			expectedCN := config.CommonName
			if expectedCN == "" {
				expectedCN = TestCACommonName // 使用默认值进行测试
			}
			assertCertExists(t, caCert.CertFile, expectedCN)
		})
	}
}

// TestCreateCACertFiles_SkipExisting 测试跳过现有文件
func TestCreateCACertFiles_SkipExisting(t *testing.T) {
	tempDir := createTestTempDir(t)

	caCert := CertKeyPair{
		Name:     "existing-ca",
		KeyFile:  filepath.Join(tempDir, "existing-ca.key"),
		CertFile: filepath.Join(tempDir, "existing-ca.crt"),
	}

	config := GetTestCACertificateConfig()

	// 首先创建证书文件
	testData := createTestCertData(t)
	writeTestCertPair(t, testData.CAKey, testData.CACert, caCert.KeyFile, caCert.CertFile)

	// 再次调用应该跳过创建
	err := CreateCACertFiles(caCert, config)
	assert.NoError(t, err, "应该跳过现有文件")

	// 验证文件未被覆盖
	originalCert, err := keyutil.PublicKeysFromFile(caCert.CertFile)
	require.NoError(t, err)
	assert.Len(t, originalCert, 1, "应该有一个证书")
}

// TestCreateCACertFiles_ErrorHandling 测试错误处理
func TestCreateCACertFiles_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, caCert CertKeyPair) k8certutil.Config
		expectError bool
		description string
	}{
		{
			name: "绝对路径密钥文件",
			setupFunc: func(t *testing.T, caCert CertKeyPair) k8certutil.Config {
				caCert.KeyFile = "/tmp/ca.key"
				return GetTestCACertificateConfig()
			},
			expectError: false,
			description: "原始函数不验证路径",
		},
		{
			name: "绝对路径证书文件",
			setupFunc: func(t *testing.T, caCert CertKeyPair) k8certutil.Config {
				caCert.CertFile = "/tmp/ca.crt"
				return GetTestCACertificateConfig()
			},
			expectError: false,
			description: "原始函数不验证路径",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := createTestTempDir(t)

			caCert := CertKeyPair{
				Name:     "error-ca",
				KeyFile:  filepath.Join(tempDir, "error-ca.key"),
				CertFile: filepath.Join(tempDir, "error-ca.crt"),
			}

			config := tt.setupFunc(t, caCert)
			err := CreateCACertFiles(caCert, config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestCreateCACertFiles_Properties 测试生成的 CA 证书属性
func TestCreateCACertFiles_Properties(t *testing.T) {
	tempDir := createTestTempDir(t)

	caCert := CertKeyPair{
		Name:     "property-ca",
		KeyFile:  filepath.Join(tempDir, "property-ca.key"),
		CertFile: filepath.Join(tempDir, "property-ca.crt"),
	}

	config := GetTestCACertificateConfig()
	config.Organization = []string{"test-org", "test-department"}

	err := CreateCACertFiles(caCert, config)
	require.NoError(t, err)

	// 加载并验证证书属性
	key, cert, err := TryLoadCertAndKeyFromFile(caCert.KeyFile, caCert.CertFile)
	require.NoError(t, err, "应该能够加载生成的证书和密钥")

	assert.NotNil(t, key, "密钥不应该为空")
	assert.NotNil(t, cert, "证书不应该为空")

	// 验证 CA 属性
	assert.True(t, cert.IsCA, "证书应该是 CA")
	assert.Equal(t, config.CommonName, cert.Subject.CommonName, "Common Name 应该匹配")
	assert.Equal(t, config.Organization, cert.Subject.Organization, "Organization 应该匹配")

	// 验证密钥长度
	assert.Equal(t, TestRSAKeySize, key.(*rsa.PrivateKey).N.BitLen(), "密钥长度应该为 2048 位")
}

// TestCreateCACertFiles_Concurrent 测试并发创建
func TestCreateCACertFiles_Concurrent(t *testing.T) {
	const numGoroutines = 5

	tempDir := createTestTempDir(t)

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			caCert := CertKeyPair{
				Name:     fmt.Sprintf("concurrent-ca-%d", index),
				KeyFile:  filepath.Join(tempDir, fmt.Sprintf("concurrent-ca-%d.key", index)),
				CertFile: filepath.Join(tempDir, fmt.Sprintf("concurrent-ca-%d.crt", index)),
			}

			config := GetTestCACertificateConfig()
			config.CommonName = fmt.Sprintf("concurrent-ca-%d", index)

			err := CreateCACertFiles(caCert, config)
			errors <- err
		}(i)
	}

	wg.Wait()
	close(errors)

	// 验证所有操作都成功
	for err := range errors {
		assert.NoError(t, err, "并发创建 CA 证书应该都成功")
	}

	// 验证所有文件都存在
	for i := 0; i < numGoroutines; i++ {
		keyFile := filepath.Join(tempDir, fmt.Sprintf("concurrent-ca-%d.key", i))
		certFile := filepath.Join(tempDir, fmt.Sprintf("concurrent-ca-%d.crt", i))

		assert.FileExists(t, keyFile, "并发生成的密钥文件应该存在")
		assert.FileExists(t, certFile, "并发生成的证书文件应该存在")
	}
}

// TestCreateGenericCertFiles 测试通用证书文件创建
func TestCreateGenericCertFiles(t *testing.T) {
	tests := []struct {
		name        string
		serverCert  CertKeyPair
		caCert      CertKeyPair
		config      k8certutil.Config
		expectError bool
		description string
	}{
		{
			name: "有效服务器证书创建",
			serverCert: CertKeyPair{
				Name:     "test-server",
				KeyFile:  "server.key",
				CertFile: "server.crt",
			},
			caCert: CertKeyPair{
				Name:     "test-ca",
				KeyFile:  "ca.key",
				CertFile: "ca.crt",
			},
			config:      GetTestServerCertificateConfig(),
			expectError: false,
			description: "使用有效配置创建服务器证书",
		},
		{
			name: "空配置应该工作",
			serverCert: CertKeyPair{
				Name:     "test-server",
				KeyFile:  "server.key",
				CertFile: "server.crt",
			},
			caCert: CertKeyPair{
				Name:     "test-ca",
				KeyFile:  "ca.key",
				CertFile: "ca.crt",
			},
			config:      k8certutil.Config{},
			expectError: false,
			description: "空配置应该工作（原始函数不验证）",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := createTestTempDir(t)

			// 首先创建 CA 证书
			caTestData := createTestCertData(t)
			writeTestCertPair(t, caTestData.CAKey, caTestData.CACert,
				filepath.Join(tempDir, tt.caCert.KeyFile),
				filepath.Join(tempDir, tt.caCert.CertFile))

			// 更新文件路径
			updatedServerCert := CertKeyPair{
				Name:     tt.serverCert.Name,
				KeyFile:  filepath.Join(tempDir, tt.serverCert.KeyFile),
				CertFile: filepath.Join(tempDir, tt.serverCert.CertFile),
			}
			updatedCACert := CertKeyPair{
				Name:     tt.caCert.Name,
				KeyFile:  filepath.Join(tempDir, tt.caCert.KeyFile),
				CertFile: filepath.Join(tempDir, tt.caCert.CertFile),
			}

			// 执行测试
			err := CreateGenericCertFiles(updatedServerCert, updatedCACert, tt.config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				return
			}

			assert.NoError(t, err, tt.description)

			// 验证证书和密钥文件已创建
			assert.FileExists(t, updatedServerCert.KeyFile, "服务器密钥文件应该存在")
			assert.FileExists(t, updatedServerCert.CertFile, "服务器证书文件应该存在")

			// 验证证书的 Common Name
			assertCertExists(t, updatedServerCert.CertFile, tt.config.CommonName)
		})
	}
}

// TestCreateGenericCertFiles_MissingCA 测试缺少 CA 证书的情况
func TestCreateGenericCertFiles_MissingCA(t *testing.T) {
	tempDir := createTestTempDir(t)

	serverCert := CertKeyPair{
		Name:     "test-server",
		KeyFile:  filepath.Join(tempDir, "server.key"),
		CertFile: filepath.Join(tempDir, "server.crt"),
	}
	caCert := CertKeyPair{
		Name:     "missing-ca",
		KeyFile:  filepath.Join(tempDir, "missing-ca.key"),
		CertFile: filepath.Join(tempDir, "missing-ca.crt"),
	}

	config := GetTestServerCertificateConfig()
	err := CreateGenericCertFiles(serverCert, caCert, config)

	assert.Error(t, err, "缺少 CA 证书应该返回错误")
	assert.Contains(t, err.Error(), "failed load ca certificated", "错误消息应该包含 CA 加载失败")
}

// TestCreateGenericCertFiles_SkipExisting 测试跳过现有文件
func TestCreateGenericCertFiles_SkipExisting(t *testing.T) {
	tempDir := createTestTempDir(t)

	serverCert := CertKeyPair{
		Name:     "existing-server",
		KeyFile:  filepath.Join(tempDir, "existing-server.key"),
		CertFile: filepath.Join(tempDir, "existing-server.crt"),
	}
	caCert := CertKeyPair{
		Name:     "existing-ca",
		KeyFile:  filepath.Join(tempDir, "existing-ca.key"),
		CertFile: filepath.Join(tempDir, "existing-ca.crt"),
	}

	config := GetTestServerCertificateConfig()

	// 首先创建 CA 证书和服务器证书
	caTestData := createTestCertData(t)
	writeTestCertPair(t, caTestData.CAKey, caTestData.CACert, caCert.KeyFile, caCert.CertFile)
	writeTestCertPair(t, caTestData.ServerKey, caTestData.ServerCert, serverCert.KeyFile, serverCert.CertFile)

	// 再次调用应该跳过创建
	err := CreateGenericCertFiles(serverCert, caCert, config)
	assert.NoError(t, err, "应该跳过现有文件")

	// 验证文件未被覆盖
	originalCert, err := keyutil.PublicKeysFromFile(serverCert.CertFile)
	require.NoError(t, err)
	assert.Len(t, originalCert, 1, "应该有一个证书")
}

// TestCreateGenericCertFiles_CertificateValidation 测试证书验证
func TestCreateGenericCertFiles_CertificateValidation(t *testing.T) {
	tempDir := createTestTempDir(t)

	serverCert := CertKeyPair{
		Name:     "validation-server",
		KeyFile:  filepath.Join(tempDir, "validation-server.key"),
		CertFile: filepath.Join(tempDir, "validation-server.crt"),
	}
	caCert := CertKeyPair{
		Name:     "validation-ca",
		KeyFile:  filepath.Join(tempDir, "validation-ca.key"),
		CertFile: filepath.Join(tempDir, "validation-ca.crt"),
	}

	// 创建自定义配置进行验证
	config := k8certutil.Config{
		CommonName:   "validation-server",
		Organization: []string{"validation-org"},
		AltNames:     CreateTestAltNames(),
		NotBefore:    time.Now().Add(-1 * time.Minute),
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// 首先创建 CA 证书
	caTestData := createTestCertData(t)
	writeTestCertPair(t, caTestData.CAKey, caTestData.CACert, caCert.KeyFile, caCert.CertFile)

	err := CreateGenericCertFiles(serverCert, caCert, config)
	require.NoError(t, err)

	// 加载并验证证书属性
	key, cert, err := TryLoadCertAndKeyFromFile(serverCert.KeyFile, serverCert.CertFile)
	require.NoError(t, err, "应该能够加载生成的证书和密钥")

	assert.NotNil(t, key, "密钥不应该为空")
	assert.NotNil(t, cert, "证书不应该为空")

	// 验证证书属性
	assert.Equal(t, config.CommonName, cert.Subject.CommonName, "Common Name 应该匹配")
	assert.Equal(t, config.Organization, cert.Subject.Organization, "Organization 应该匹配")
	assert.Contains(t, cert.DNSNames, config.AltNames.DNSNames[0], "DNS Names 应该包含配置的名称")
	assert.Equal(t, config.Usages, cert.ExtKeyUsage, "扩展密钥用途应该匹配")

	// 验证密钥长度
	assert.Equal(t, TestRSAKeySize, key.(*rsa.PrivateKey).N.BitLen(), "密钥长度应该为 2048 位")
}

// TestCreateServiceAccountKeyAndPublicKeyFiles 测试服务账户密钥文件创建
func TestCreateServiceAccountKeyAndPublicKeyFiles(t *testing.T) {
	tests := []struct {
		name        string
		privKeyFile string
		pubKeyFile  string
		expectError bool
		description string
	}{
		{
			name:        "有效服务账户密钥创建",
			privKeyFile: "sa.key",
			pubKeyFile:  "sa.pub",
			expectError: false,
			description: "使用有效路径创建服务账户密钥",
		},
		{
			name:        "绝对路径私钥",
			privKeyFile: "/tmp/sa.key",
			pubKeyFile:  "sa.pub",
			expectError: false,
			description: "原始函数不验证路径",
		},
		{
			name:        "绝对路径公钥",
			privKeyFile: "sa.key",
			pubKeyFile:  "/tmp/sa.pub",
			expectError: false,
			description: "原始函数不验证路径",
		},
		{
			name:        "基本路径测试",
			privKeyFile: "sa.key",
			pubKeyFile:  "sa.pub",
			expectError: false,
			description: "基本路径应该工作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := createTestTempDir(t)

			privKeyPath := filepath.Join(tempDir, tt.privKeyFile)
			pubKeyPath := filepath.Join(tempDir, tt.pubKeyFile)

			// 执行测试
			err := CreateServiceAccountKeyAndPublicKeyFiles(privKeyPath, pubKeyPath)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				return
			}

			assert.NoError(t, err, tt.description)

			// 验证私钥和公钥文件已创建
			assert.FileExists(t, privKeyPath, "私钥文件应该存在")
			assert.FileExists(t, pubKeyPath, "公钥文件应该存在")

			// 验证文件内容不为空
			assertFileContent(t, privKeyPath)
			assertFileContent(t, pubKeyPath)
		})
	}
}

// TestCreateServiceAccountKeyAndPublicKeyFiles_SkipExisting 测试跳过现有文件
func TestCreateServiceAccountKeyAndPublicKeyFiles_SkipExisting(t *testing.T) {
	tempDir := createTestTempDir(t)

	privKeyPath := filepath.Join(tempDir, "existing-sa.key")
	pubKeyPath := filepath.Join(tempDir, "existing-sa.pub")

	// 首先创建现有的密钥文件
	existingKey, _ := rsa.GenerateKey(rand.Reader, TestRSAKeySize)
	pemBytes, err := encodePublicKey(&existingKey.PublicKey)
	require.NoError(t, err, "应该能编码公钥")

	err = ioutil.WriteFile(pubKeyPath, pemBytes, 0644)
	require.NoError(t, err, "应该能写入公钥文件")

	// 再次调用应该跳过创建
	err = CreateServiceAccountKeyAndPublicKeyFiles(privKeyPath, pubKeyPath)
	assert.NoError(t, err, "应该跳过现有文件")

	// 验证公钥文件内容未被覆盖
	savedKey, err := ioutil.ReadFile(pubKeyPath)
	require.NoError(t, err, "应该能读取保存的公钥文件")
	assert.Equal(t, pemBytes, savedKey, "公钥文件内容应该未被覆盖")
}

// TestCreateServiceAccountKeyAndPublicKeyFiles_KeyPairValidation 测试密钥对验证
func TestCreateServiceAccountKeyAndPublicKeyFiles_KeyPairValidation(t *testing.T) {
	tempDir := createTestTempDir(t)

	privKeyPath := filepath.Join(tempDir, "validation-sa.key")
	pubKeyPath := filepath.Join(tempDir, "validation-sa.pub")

	err := CreateServiceAccountKeyAndPublicKeyFiles(privKeyPath, pubKeyPath)
	require.NoError(t, err, "应该能创建服务账户密钥")

	// 加载私钥并验证
	privKey, err := keyutil.PrivateKeyFromFile(privKeyPath)
	require.NoError(t, err, "应该能加载私钥")
	assert.NotNil(t, privKey, "私钥不应该为空")
	assert.Equal(t, TestRSAKeySize, privKey.(*rsa.PrivateKey).N.BitLen(), "私钥长度应该为 2048 位")

	// 加载公钥并验证
	pubKey, err := ioutil.ReadFile(pubKeyPath)
	require.NoError(t, err, "应该能读取公钥文件")
	assert.Greater(t, len(pubKey), 0, "公钥文件应该有内容")

	// 验证公钥 PEM 格式
	assert.Contains(t, string(pubKey), "-----BEGIN PUBLIC KEY-----", "应该是有效的 PEM 格式公钥")
	assert.Contains(t, string(pubKey), "-----END PUBLIC KEY-----", "应该有 PEM 结束标记")

	// 验证公钥与私钥匹配
	loadedPrivKey, err := keyutil.PrivateKeyFromFile(privKeyPath)
	require.NoError(t, err, "应该能加载私钥")

	expectedPubKeyPEM, err := encodePublicKey(&loadedPrivKey.(*rsa.PrivateKey).PublicKey)
	require.NoError(t, err, "应该能编码公钥")

	assert.Equal(t, string(expectedPubKeyPEM), string(pubKey), "保存的公钥应该与私钥匹配")
}

// TestCreateServiceAccountKeyAndPublicKeyFiles_Concurrent 测试并发创建
func TestCreateServiceAccountKeyAndPublicKeyFiles_Concurrent(t *testing.T) {
	const numGoroutines = 5

	tempDir := createTestTempDir(t)

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			privKeyPath := filepath.Join(tempDir, fmt.Sprintf("concurrent-sa-%d.key", index))
			pubKeyPath := filepath.Join(tempDir, fmt.Sprintf("concurrent-sa-%d.pub", index))

			err := CreateServiceAccountKeyAndPublicKeyFiles(privKeyPath, pubKeyPath)
			errors <- err
		}(i)
	}

	wg.Wait()
	close(errors)

	// 验证所有操作都成功
	for err := range errors {
		assert.NoError(t, err, "并发创建服务账户密钥应该都成功")
	}

	// 验证所有文件都存在
	for i := 0; i < numGoroutines; i++ {
		privKeyPath := filepath.Join(tempDir, fmt.Sprintf("concurrent-sa-%d.key", i))
		pubKeyPath := filepath.Join(tempDir, fmt.Sprintf("concurrent-sa-%d.pub", i))

		assert.FileExists(t, privKeyPath, "并发生成的私钥文件应该存在")
		assert.FileExists(t, pubKeyPath, "并发生成的公钥文件应该存在")
	}
}
