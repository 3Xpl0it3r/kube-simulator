## 1. 测试环境准备和基础设施
- [x] 1.1 分析 pkg/cert 包的现有代码结构
- [x] 1.2 创建证书测试的通用辅助函数
- [x] 1.3 设置测试数据和临时文件管理
- [x] 1.4 建立 mock 文件系统操作

## 2. pkg/cert/cert.go 测试实现
- [x] 2.1 为 CreateCACertFiles 函数添加单元测试
- [x] 2.2 为 CreateGenericCertFiles 函数添加单元测试
- [x] 2.3 为 CreateServiceAccountKeyAndPublicKeyFiles 函数添加单元测试

## 3. pkg/cert/config.go 测试实现
- [x] 3.1 为 NewCACertificateConfig 函数添加单元测试
- [x] 3.2 为 NewServerCerfiticateConfig 函数添加单元测试
- [x] 3.3 为 NewClientCertificateConfig 函数添加单元测试

## 4. pkg/cert/util.go 测试实现
- [x] 4.1 为 TryWriteCertAndKeyToFile 函数添加单元测试
- [x] 4.2 为 TryLoadCertAndKeyFromFile 函数添加单元测试
- [x] 4.3 为 NewCertAndKey 函数添加单元测试
- [x] 4.4 为 newSignedCert 函数添加单元测试
- [x] 4.5 为 EncodeCertPEM 函数添加单元测试
- [x] 4.6 为 encodePublicKey 函数添加单元测试
  - [ ] 4.6.1 测试公钥编码格式
  - [ ] 4.6.2 测试不同密钥类型支持

## 5. 集成验证和覆盖率
- [x] 5.1 运行 pkg/cert 包完整测试套件
- [x] 5.2 生成测试覆盖率报告
- [x] 5.3 确保覆盖率达到 85%+
- [x] 5.4 运行竞态条件检测

## 6. 边界条件和错误处理测试
- [x] 6.1 测试无效文件路径处理
- [x] 6.2 测试权限不足场景
- [x] 6.3 测试磁盘空间不足情况
- [x] 6.4 测试并发证书生成安全性
- [x] 6.5 测试大证书集合性能

## 7. 文档和维护
- [x] 7.1 创建证书测试运行指南
- [x] 7.2 添加测试数据清理脚本
- [x] 7.3 准备证书测试最佳实践文档
- [x] 7.4 为未来证书功能扩展准备测试模板

## 8. 质量保证
- [x] 8.1 验证所有测试的独立性
- [x] 8.2 确保测试可重复性和稳定性
- [x] 8.3 检查测试性能和执行时间
- [x] 8.4 验证 mock 函数的正确性
- [x] 8.5 确保测试数据安全性（不泄露真实密钥）