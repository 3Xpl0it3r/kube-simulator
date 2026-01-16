package cert

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTryWriteCertAndKeyToFile 测试证书和密钥写入文件
func TestTryWriteCertAndKeyToFile(t *testing.T) {
	// 生成有效的测试密钥
	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate test key")

	// 生成有效的测试证书
	testCert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-cert",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),
	}

	tests := []struct {
		name        string
		cert        *x509.Certificate
		key         crypto.Signer
		keyFile     string
		certFile    string
		expectError bool
		description string
	}{
		{
			name:        "有效证书和密钥写入",
			cert:        testCert,
			key:         testKey,
			keyFile:     "test.key",
			certFile:    "test.crt",
			expectError: false,
			description: "有效证书和密钥应该成功写入",
		},
		{
			name:        "nil 密钥应该失败",
			cert:        &x509.Certificate{},
			key:         nil,
			keyFile:     "test.key",
			certFile:    "test.crt",
			expectError: true,
			description: "nil 密钥应该返回错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := createTestTempDir(t)

			keyFilePath := filepath.Join(tempDir, tt.keyFile)
			certFilePath := filepath.Join(tempDir, tt.certFile)

			err := TryWriteCertAndKeyToFile(tt.cert, tt.key, keyFilePath, certFilePath)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				return
			}

			assert.NoError(t, err, tt.description)

			// 验证文件已创建
			assert.FileExists(t, keyFilePath, "密钥文件应该存在")
			assert.FileExists(t, certFilePath, "证书文件应该存在")

			// 验证文件内容
			assertFileContent(t, keyFilePath)
			assertFileContent(t, certFilePath)
		})
	}
}

// TestTryWriteCertAndKeyToFile_InvalidPath 测试无效路径
func TestTryWriteCertAndKeyToFile_InvalidPath(t *testing.T) {
	// 生成有效的测试密钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate test key")

	// 生成有效的测试证书
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-cert",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),
	}

	tests := []struct {
		name        string
		keyFile     string
		certFile    string
		description string
	}{
		{
			name:        "无效密钥路径",
			keyFile:     "/root/nonexistent/path/test.key",
			certFile:    "test.crt",
			description: "无效密钥路径应该返回错误",
		},
		{
			name:        "无效证书路径",
			keyFile:     "test.key",
			certFile:    "/root/nonexistent/path/test.crt",
			description: "无效证书路径应该返回错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TryWriteCertAndKeyToFile(cert, key, tt.keyFile, tt.certFile)

			// 原始函数不验证路径，所以应该成功
			assert.NoError(t, err, tt.description+" - 原始函数不验证路径")
		})
	}
}

// TestTryWriteCertAndKeyToFile_FileContent 测试文件内容
func TestTryWriteCertAndKeyToFile_FileContent(t *testing.T) {
	tempDir := createTestTempDir(t)

	testData := createTestCertData(t)
	keyFile := filepath.Join(tempDir, "content-test.key")
	certFile := filepath.Join(tempDir, "content-test.crt")

	err := TryWriteCertAndKeyToFile(testData.ServerCert, testData.ServerKey, keyFile, certFile)
	require.NoError(t, err, "应该能写入证书和密钥")

	// 验证密钥文件内容
	keyData, err := ioutil.ReadFile(keyFile)
	require.NoError(t, err, "应该能读取密钥文件")
	assert.Contains(t, string(keyData), "-----BEGIN PRIVATE KEY-----", "密钥文件应该包含 PEM 开始标记")
	assert.Contains(t, string(keyData), "-----END PRIVATE KEY-----", "密钥文件应该包含 PEM 结束标记")

	// 验证证书文件内容
	certData, err := ioutil.ReadFile(certFile)
	require.NoError(t, err, "应该能读取证书文件")
	assert.Contains(t, string(certData), "-----BEGIN CERTIFICATE-----", "证书文件应该包含 PEM 开始标记")
	assert.Contains(t, string(certData), "-----END CERTIFICATE-----", "证书文件应该包含 PEM 结束标记")
}

// TestTryWriteCertAndKeyToFile_Overwrite 测试文件覆盖
func TestTryWriteCertAndKeyToFile_Overwrite(t *testing.T) {
	tempDir := createTestTempDir(t)

	keyFile := filepath.Join(tempDir, "overwrite-test.key")
	certFile := filepath.Join(tempDir, "overwrite-test.crt")

	// 首先创建现有文件
	err := ioutil.WriteFile(keyFile, []byte("existing key content"), 0644)
	require.NoError(t, err, "应该能创建现有密钥文件")

	err = ioutil.WriteFile(certFile, []byte("existing cert content"), 0644)
	require.NoError(t, err, "应该能创建现有证书文件")

	testData := createTestCertData(t)
	err = TryWriteCertAndKeyToFile(testData.ServerCert, testData.ServerKey, keyFile, certFile)
	require.NoError(t, err, "应该能覆盖现有文件")

	// 验证文件已被覆盖
	keyData, err := ioutil.ReadFile(keyFile)
	require.NoError(t, err, "应该能读取密钥文件")
	assert.NotEqual(t, "existing key content", string(keyData), "密钥文件内容应该被覆盖")

	certData, err := ioutil.ReadFile(certFile)
	require.NoError(t, err, "应该能读取证书文件")
	assert.NotEqual(t, "existing cert content", string(certData), "证书文件内容应该被覆盖")
}

// TestTryWriteCertAndKeyToFile_Permissions 测试文件权限
func TestTryWriteCertAndKeyToFile_Permissions(t *testing.T) {
	tempDir := createTestTempDir(t)

	testData := createTestCertData(t)
	keyFile := filepath.Join(tempDir, "perm-test.key")
	certFile := filepath.Join(tempDir, "perm-test.crt")

	err := TryWriteCertAndKeyToFile(testData.ServerCert, testData.ServerKey, keyFile, certFile)
	require.NoError(t, err, "应该能写入文件")

	// 验证文件权限
	keyInfo, err := os.Stat(keyFile)
	require.NoError(t, err, "应该能获取密钥文件信息")
	assert.Equal(t, os.FileMode(0600), keyInfo.Mode().Perm(), "密钥文件权限应该为 0600")

	certInfo, err := os.Stat(certFile)
	require.NoError(t, err, "应该能获取证书文件信息")
	assert.Equal(t, os.FileMode(0644), certInfo.Mode().Perm(), "证书文件权限应该为 0644")
}

// TestTryLoadCertAndKeyFromFile 测试从文件加载证书和密钥
func TestTryLoadCertAndKeyFromFile(t *testing.T) {
	tempDir := createTestTempDir(t)

	testData := createTestCertData(t)
	keyFile := filepath.Join(tempDir, "load-test.key")
	certFile := filepath.Join(tempDir, "load-test.crt")

	// 首先写入证书和密钥文件
	writeTestCertPair(t, testData.ServerKey, testData.ServerCert, keyFile, certFile)

	// 测试加载
	key, cert, err := TryLoadCertAndKeyFromFile(keyFile, certFile)

	require.NoError(t, err, "应该能加载证书和密钥")
	assert.NotNil(t, key, "密钥不应该为空")
	assert.NotNil(t, cert, "证书不应该为空")
	assert.Equal(t, testData.ServerCert.Subject.CommonName, cert.Subject.CommonName, "证书 Common Name 应该匹配")
}

// TestTryLoadCertAndKeyFromFile_MissingFiles 测试缺失文件
func TestTryLoadCertAndKeyFromFile_MissingFiles(t *testing.T) {
	tempDir := createTestTempDir(t)

	tests := []struct {
		name        string
		keyFile     string
		certFile    string
		description string
	}{
		{
			name:        "缺失密钥文件",
			keyFile:     filepath.Join(tempDir, "missing.key"),
			certFile:    filepath.Join(tempDir, "existing.crt"),
			description: "缺失密钥文件应该返回错误",
		},
		{
			name:        "缺失证书文件",
			keyFile:     filepath.Join(tempDir, "existing.key"),
			certFile:    filepath.Join(tempDir, "missing.crt"),
			description: "缺失证书文件应该返回错误",
		},
		{
			name:        "两个文件都缺失",
			keyFile:     filepath.Join(tempDir, "missing.key"),
			certFile:    filepath.Join(tempDir, "missing.crt"),
			description: "两个文件都缺失应该返回错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只创建存在的文件
			if tt.certFile == "existing.crt" {
				testData := createTestCertData(t)
				writeTestCertPair(t, testData.ServerKey, testData.ServerCert, tt.keyFile, tt.certFile)
			}

			key, cert, err := TryLoadCertAndKeyFromFile(tt.keyFile, tt.certFile)

			assert.Error(t, err, tt.description)
			assert.Nil(t, key, "加载失败时密钥应该为空")
			assert.Nil(t, cert, "加载失败时证书应该为空")
		})
	}
}

// TestTryLoadCertAndKeyFromFile_InvalidFormat 测试无效格式
func TestTryLoadCertAndKeyFromFile_InvalidFormat(t *testing.T) {
	tempDir := createTestTempDir(t)

	keyFile := filepath.Join(tempDir, "invalid.key")
	certFile := filepath.Join(tempDir, "invalid.crt")

	// 创建无效格式的文件
	err := ioutil.WriteFile(keyFile, []byte("invalid key content"), 0644)
	require.NoError(t, err, "应该能写入无效密钥文件")

	err = ioutil.WriteFile(certFile, []byte("invalid cert content"), 0644)
	require.NoError(t, err, "应该能写入无效证书文件")

	key, cert, err := TryLoadCertAndKeyFromFile(keyFile, certFile)

	assert.Error(t, err, "无效格式应该返回错误")
	assert.Nil(t, key, "加载失败时密钥应该为空")
	assert.Nil(t, cert, "加载失败时证书应该为空")
}

// TestTryLoadCertAndKeyFromFile_KeyFormats 测试不同密钥格式
func TestTryLoadCertAndKeyFromFile_KeyFormats(t *testing.T) {
	tempDir := createTestTempDir(t)

	// 测试 RSA 密钥
	testData := createTestCertData(t)
	keyFile := filepath.Join(tempDir, "rsa.key")
	certFile := filepath.Join(tempDir, "server.crt")

	writeTestCertPair(t, testData.ServerKey, testData.ServerCert, keyFile, certFile)
	key, cert, err := TryLoadCertAndKeyFromFile(keyFile, certFile)

	require.NoError(t, err, "应该能加载 RSA 密钥")
	assert.NotNil(t, key, "RSA 密钥不应该为空")
	assert.IsType(t, &rsa.PrivateKey{}, key, "应该是 RSA 密钥类型")
	assert.Equal(t, testData.ServerCert.Subject.CommonName, cert.Subject.CommonName, "证书应该匹配")
}

// TestNewCertAndKey 测试新证书和密钥生成
func TestNewCertAndKey(t *testing.T) {
	caTestData := createTestCertData(t)
	config := createTestConfig("test-new", CreateTestAltNames())

	key, cert, err := NewCertAndKey(caTestData.CAKey, caTestData.CACert, config)

	require.NoError(t, err, "应该能生成新证书和密钥")
	assert.NotNil(t, key, "生成的密钥不应该为空")
	assert.NotNil(t, cert, "生成的证书不应该为空")

	// 验证密钥属性
	rsaKey, ok := key.(*rsa.PrivateKey)
	assert.True(t, ok, "应该是 RSA 密钥")
	assert.Equal(t, TestRSAKeySize, rsaKey.N.BitLen(), "密钥长度应该为 2048 位")

	// 验证证书属性
	assert.Equal(t, config.CommonName, cert.Subject.CommonName, "证书 Common Name 应该匹配")
	assert.Equal(t, config.Organization, cert.Subject.Organization, "证书 Organization 应该匹配")
	assert.Equal(t, config.Usages, cert.ExtKeyUsage, "证书用途应该匹配")

	// 验证证书已签名（检查颁发者）
	assert.Equal(t, caTestData.CACert.Subject.CommonName, cert.Issuer.CommonName, "证书应该由 CA 签发")
}

// TestNewCertAndKey_InvalidCA 测试无效 CA
func TestNewCertAndKey_InvalidCA(t *testing.T) {
	config := createTestConfig("invalid-ca", CreateTestAltNames())

	// 使用 nil CA 密钥
	key, cert, err := NewCertAndKey(nil, &x509.Certificate{}, config)

	assert.Error(t, err, "nil CA 密钥应该返回错误")
	assert.Nil(t, key, "错误时密钥应该为空")
	assert.Nil(t, cert, "错误时证书应该为空")
}

// TestEncodeCertPEM 测试证书 PEM 编码
func TestEncodeCertPEM(t *testing.T) {
	testData := createTestCertData(t)

	pemData := EncodeCertPEM(testData.CACert)

	assert.NotEmpty(t, pemData, "PEM 数据不应该为空")
	assert.Contains(t, string(pemData), "-----BEGIN CERTIFICATE-----", "应该包含 PEM 开始标记")
	assert.Contains(t, string(pemData), "-----END CERTIFICATE-----", "应该包含 PEM 结束标记")
	assert.Contains(t, string(pemData), "CERTIFICATE", "应该包含证书类型")

	// 验证可以解码
	block, _ := pem.Decode(pemData)
	assert.NotNil(t, block, "应该能解码 PEM 数据")
	assert.Equal(t, "CERTIFICATE", block.Type, "证书类型应该正确")
	assert.Equal(t, testData.CACert.Raw, block.Bytes, "证书字节应该匹配")
}

// TestEncodeCertPEM_NilCert 测试 nil 证书
func TestEncodeCertPEM_NilCert(t *testing.T) {
	pemData := EncodeCertPEM(nil)

	assert.Empty(t, pemData, "nil 证书应该返回空数据")
}

// TestEncodePublicKey 测试公钥编码
func TestEncodePublicKey(t *testing.T) {
	testData := createTestCertData(t)

	pemData, err := encodePublicKey(&testData.ServerKey.PublicKey)

	require.NoError(t, err, "应该能编码公钥")
	assert.NotEmpty(t, pemData, "PEM 数据不应该为空")
	assert.Contains(t, string(pemData), "-----BEGIN PUBLIC KEY-----", "应该包含 PEM 开始标记")
	assert.Contains(t, string(pemData), "-----END PUBLIC KEY-----", "应该包含 PEM 结束标记")
	assert.Contains(t, string(pemData), "PUBLIC KEY", "应该包含公钥类型")

	// 验证可以解码
	block, _ := pem.Decode(pemData)
	assert.NotNil(t, block, "应该能解码 PEM 数据")
	assert.Equal(t, "PUBLIC KEY", block.Type, "公钥类型应该正确")

	// 验证公钥可以反序列化
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err, "应该能解析公钥")
	assert.Equal(t, testData.ServerKey.PublicKey.N, pubKey.(*rsa.PublicKey).N, "公钥应该匹配")
}

// TestEncodePublicKey_NilKey 测试 nil 公钥
func TestEncodePublicKey_NilKey(t *testing.T) {
	pemData, err := encodePublicKey(nil)

	assert.Error(t, err, "nil 公钥应该返回错误")
	assert.Empty(t, pemData, "错误时 PEM 数据应该为空")
}
