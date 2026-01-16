package cert

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8certutil "k8s.io/client-go/util/cert"
)

// TestCertData 包含测试所需的证书数据
type TestCertData struct {
	CAKey      *rsa.PrivateKey
	CACert     *x509.Certificate
	ServerKey  *rsa.PrivateKey
	ServerCert *x509.Certificate
	ClientKey  *rsa.PrivateKey
	ClientCert *x509.Certificate
	TempDir    string
}

// createTestTempDir 创建临时测试目录并注册清理函数
func createTestTempDir(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "cert-test-")
	require.NoError(t, err, "Failed to create temp directory")

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}

// generateTestCA 生成测试用的 CA 证书和密钥
func generateTestCA(t *testing.T) (*rsa.PrivateKey, *x509.Certificate) {
	// 确保生成真正的2048位密钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate test CA key")

	// 验证密钥长度
	assert.Equal(t, 2048, key.N.BitLen(), "CA密钥长度应该为 2048 位")

	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(1000))
	require.NoError(t, err, "Failed to generate serial number")

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "test-ca",
			Organization: []string{"test-org"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err, "Failed to create test CA certificate")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "Failed to parse test CA certificate")

	return key, cert
}

// generateTestCertPair 生成测试用的证书密钥对
func generateTestCertPair(t *testing.T, caKey *rsa.PrivateKey, caCert *x509.Certificate, commonName string) (*rsa.PrivateKey, *x509.Certificate) {
	// 确保生成真正的2048位密钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate test key")

	// 验证密钥长度
	assert.Equal(t, 2048, key.N.BitLen(), "服务器密钥长度应该为 2048 位")

	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(1000))
	require.NoError(t, err, "Failed to generate serial number")

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"test-org"},
		},
		NotBefore:             time.Now().Add(-30 * time.Minute),
		NotAfter:              time.Now().Add(30 * time.Minute),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &key.PublicKey, caKey)
	require.NoError(t, err, "Failed to create test certificate")

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err, "Failed to parse test certificate")

	return key, cert
}

// createTestCertData 创建完整的测试证书数据集
func createTestCertData(t *testing.T) *TestCertData {
	tempDir := createTestTempDir(t)

	// 生成 CA
	caKey, caCert := generateTestCA(t)

	// 生成服务器证书
	serverKey, serverCert := generateTestCertPair(t, caKey, caCert, "test-server")

	// 生成客户端证书
	clientKey, clientCert := generateTestCertPair(t, caKey, caCert, "test-client")

	return &TestCertData{
		CAKey:      caKey,
		CACert:     caCert,
		ServerKey:  serverKey,
		ServerCert: serverCert,
		ClientKey:  clientKey,
		ClientCert: clientCert,
		TempDir:    tempDir,
	}
}

// writeTestCertPair 将测试证书和密钥写入文件
func writeTestCertPair(t *testing.T, key crypto.Signer, cert *x509.Certificate, keyFile, certFile string) {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err, "Failed to marshal test key")

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	err = ioutil.WriteFile(keyFile, keyPEM, 0600)
	require.NoError(t, err, "Failed to write test key file")

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	err = ioutil.WriteFile(certFile, certPEM, 0644)
	require.NoError(t, err, "Failed to write test cert file")
}

// assertCertExists 验证证书文件存在且格式正确
func assertCertExists(t *testing.T, certFile string, expectedCommonName string) {
	require.FileExists(t, certFile, "Certificate file should exist")

	data, err := ioutil.ReadFile(certFile)
	require.NoError(t, err, "Failed to read certificate file")

	block, _ := pem.Decode(data)
	require.NotNil(t, block, "Should decode PEM block")
	require.Equal(t, "CERTIFICATE", block.Type, "Should be certificate PEM block")

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err, "Should parse certificate")
	require.Equal(t, expectedCommonName, cert.Subject.CommonName, "Certificate should have expected common name")
}

// assertKeyExists 验证密钥文件存在且格式正确
func assertKeyExists(t *testing.T, keyFile string) {
	require.FileExists(t, keyFile, "Key file should exist")

	data, err := ioutil.ReadFile(keyFile)
	require.NoError(t, err, "Failed to read key file")

	block, _ := pem.Decode(data)
	require.NotNil(t, block, "Should decode PEM block")
	require.Contains(t, block.Type, "PRIVATE KEY", "Should be private key PEM block")
}

// createTestConfig 创建测试用的证书配置
func createTestConfig(commonName string, altNames k8certutil.AltNames) k8certutil.Config {
	return k8certutil.Config{
		CommonName:   commonName,
		Organization: []string{"test-org"},
		AltNames:     altNames,
		NotBefore:    time.Now().Add(-1 * time.Minute),
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
}

// generateTestSerial 生成测试用的序列号
func generateTestSerial(t *testing.T) *big.Int {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(1000000))
	require.NoError(t, err, "Failed to generate serial number")
	return serial
}

// createEmptyFiles 创建空文件用于测试文件存在检查
func createEmptyFiles(t *testing.T, dir string, filenames ...string) {
	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		err := ioutil.WriteFile(path, []byte{}, 0644)
		require.NoError(t, err, "Failed to create empty file: %s", filename)
	}
}

// assertFileContent 验证文件内容不为空
func assertFileContent(t *testing.T, filename string) {
	data, err := ioutil.ReadFile(filename)
	require.NoError(t, err, "Failed to read file: %s", filename)
	assert.Greater(t, len(data), 0, "File should have content: %s", filename)
}

// cleanupTempFiles 清理临时文件
func cleanupTempFiles(t *testing.T, files ...string) {
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			os.Remove(file)
		}
	}
}

// TestFileSystem 实现用于测试的文件系统接口
type TestFileSystem struct {
	files map[string][]byte
}

// NewTestFileSystem 创建新的测试文件系统
func NewTestFileSystem() *TestFileSystem {
	return &TestFileSystem{
		files: make(map[string][]byte),
	}
}

// WriteFile 模拟文件写入
func (fs *TestFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	fs.files[path] = data
	return nil
}

// ReadFile 模拟文件读取
func (fs *TestFileSystem) ReadFile(path string) ([]byte, error) {
	data, exists := fs.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

// Stat 模拟文件状态检查
func (fs *TestFileSystem) Stat(path string) (os.FileInfo, error) {
	data, exists := fs.files[path]
	if !exists {
		return nil, os.ErrNotExist
	}
	return &mockFileInfo{path: path, size: int64(len(data))}, nil
}

// Remove 模拟文件删除
func (fs *TestFileSystem) Remove(path string) error {
	delete(fs.files, path)
	return nil
}

// HasFile 检查文件是否存在
func (fs *TestFileSystem) HasFile(path string) bool {
	_, exists := fs.files[path]
	return exists
}

// GetContent 获取文件内容
func (fs *TestFileSystem) GetContent(path string) []byte {
	return fs.files[path]
}

// mockFileInfo 实现测试用的文件信息
type mockFileInfo struct {
	path string
	size int64
}

func (fi *mockFileInfo) Name() string       { return filepath.Base(fi.path) }
func (fi *mockFileInfo) Size() int64        { return fi.size }
func (fi *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (fi *mockFileInfo) ModTime() time.Time { return time.Now() }
func (fi *mockFileInfo) IsDir() bool        { return false }
func (fi *mockFileInfo) Sys() interface{}   { return nil }
