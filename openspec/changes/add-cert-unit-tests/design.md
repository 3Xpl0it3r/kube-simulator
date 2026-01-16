# Design Document: pkg/cert Unit Testing Implementation

## Overview
本文档描述了为 pkg/cert 包实施全面单元测试的设计决策和技术方法。

## Architecture Considerations

### 1. Test Structure Strategy

#### 分层测试方法
```
pkg/cert/
├── cert.go          ────┐
├── config.go        ────┤── Unit Tests (this proposal)
├── util.go          ────┘
└── *_test.go        (new files)
```

#### 测试分类
- **单元测试**: 单独函数测试，使用 mock 外部依赖
- **集成测试**: 函数间协作测试，使用真实文件系统
- **安全测试**: 加密正确性和边界条件验证
- **性能测试**: 大规模操作和并发安全性

### 2. Mock Strategy

#### 文件系统 Mock
由于证书操作涉及大量文件 I/O，我们需要实现文件系统 mock：

```go
type FileSystem interface {
    WriteFile(path string, data []byte, perm fs.FileMode) error
    ReadFile(path string) ([]byte, error)
    Stat(path string) (fs.FileInfo, error)
    Remove(path string) error
}
```

#### 加密操作 Mock
对于某些测试场景，可能需要 mock 加密操作：

```go
type CryptoGenerator interface {
    GenerateRSAKey(bits int) (*rsa.PrivateKey, error)
    CreateCertificate(template, parent *x509.Certificate, 
                     pubKey crypto.PublicKey, privKey crypto.Signer) 
                     (*x509.Certificate, error)
}
```

### 3. Test Data Management

#### 测试证书层次结构
```
Test CA
├── Server Certificate (for API Server)
├── Client Certificate (for ControllerManager)
├── Service Account Key Pair
└── Additional Test Certificates
```

#### 临时目录管理
```go
func createTestTempDir(t *testing.T) string {
    tempDir, err := ioutil.TempDir("", "cert-test-")
    require.NoError(t, err)
    t.Cleanup(func() {
        os.RemoveAll(tempDir)
    })
    return tempDir
}
```

### 4. Security Considerations

#### 密钥安全
- 测试中使用专门的测试密钥对
- 不在生产代码路径中存储真实密钥
- 确保测试后完全清理临时文件

#### 加密算法验证
- 验证 RSA 密钥生成正确性
- 检查证书签名算法符合预期
- 确保证书扩展字段正确设置

### 5. Performance Considerations

#### 基准测试
```go
func BenchmarkCertificateGeneration(b *testing.B) {
    // 设置测试环境
    for i := 0; i < b.N; i++ {
        // 执行证书生成操作
    }
}
```

#### 并发测试
```go
func TestConcurrentCertificateGeneration(t *testing.T) {
    const numGoroutines = 10
    const numCertsPerGoroutine = 5
    
    var wg sync.WaitGroup
    // 并发生成证书并验证无竞态条件
}
```

### 6. Error Handling Strategy

#### 错误类型分类
```go
type TestErrorType int

const (
    ErrorTypeFileSystem TestErrorType = iota
    ErrorTypeCrypto
    ErrorTypeValidation
    ErrorTypeConfiguration
)
```

#### 错误验证
```go
func assertErrorType(t *testing.T, err error, expectedType TestErrorType) {
    // 验证错误类型和消息内容
}
```

## Implementation Decisions

### 1. Test Framework Choice
- **Go testing package**: 标准，内置支持
- **testify/assert**: 提供丰富的断言函数
- **testify/require**: 失败时立即停止测试

### 2. Coverage Targets
- **整体覆盖率**: ≥85%
- **关键函数覆盖率**: 100%
- **错误路径覆盖率**: ≥90%

### 3. Test Organization
```
cert_test.go       - 证书生成核心函数测试
config_test.go     - 配置管理测试  
util_test.go       - 工具函数测试
helpers_test.go    - 测试辅助函数
fixtures_test.go   - 测试数据和常量
```

### 4. Dependency Injection
为了支持测试中的 mock，需要重构部分代码以支持依赖注入：

```go
type CertificateManager struct {
    fs      FileSystem
    crypto  CryptoGenerator
    clock   Clock  // 用于时间相关测试
}
```

## Risk Mitigation

### 1. Test Isolation
- 每个测试使用独立的临时目录
- 确保测试之间不相互影响
- 使用 `t.Cleanup()` 保证资源清理

### 2. Test Reliability
- 避免依赖系统时间进行断言
- 使用确定性测试数据
- 处理不同操作系统下的文件系统差异

### 3. Performance Impact
- 限制基准测试的运行范围
- 使用超时机制防止长时间运行的测试
- 提供快速测试模式用于 CI/CD

## Validation Strategy

### 1. Automated Testing
- CI/CD 集成中运行所有测试
- 覆盖率报告和门禁检查
- 竞态条件检测 (`go test -race`)

### 2. Manual Review Points
- 测试用例覆盖所有业务场景
- 错误消息的准确性和有用性
- 测试数据的完整性和安全性

### 3. Ongoing Maintenance
- 定期更新测试证书的有效期
- 监控测试执行时间趋势
- 根据新功能需求扩展测试

## Success Criteria

### 1. Functional Requirements
- ✅ 所有现有函数都有对应的单元测试
- ✅ 测试覆盖所有错误路径
- ✅ 关键安全功能得到充分验证

### 2. Quality Requirements  
- ✅ 测试覆盖率达到 85%+
- ✅ 所有测试通过且稳定
- ✅ 无竞态条件或内存泄漏

### 3. Maintainability Requirements
- ✅ 测试代码清晰易懂
- ✅ 提供足够的测试文档
- ✅ 易于扩展和维护

## Future Considerations

### 1. Test Evolution
- 随着证书功能扩展相应增加测试
- 考虑添加模糊测试 (fuzz testing)
- 探索属性基测试 (property-based testing)

### 2. Tooling Enhancement
- 开发专门的证书测试工具
- 集成自动化证书验证工具
- 提供测试报告可视化

### 3. Cross-Package Testing
- 与其他包的集成测试
- 端到端证书生命周期测试
- 性能回归测试集成