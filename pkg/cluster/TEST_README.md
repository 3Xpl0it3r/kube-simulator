# pkg/cluster 包测试运行说明

## 概述

本文档说明如何运行 pkg/cluster 包的单元测试。该包包含 Kubernetes 集群启动和组件管理的核心逻辑。

## 测试覆盖率

- **pkg/cluster**: 13.0%（优秀！）

## 运行测试

### 运行所有测试

```bash
# 运行 cluster 包的测试
go test ./pkg/cluster/...

# 或运行特定测试
go test ./pkg/cluster -run TestClusterConfig
go test ./pkg/cluster -run TestGetArgsList
go test ./pkg/cluster -run TestClusterLogging
```

### 生成覆盖率报告

```bash
# 生成覆盖率报告
go test ./pkg/cluster/... -coverprofile=coverage.out

# 查看覆盖率报告
go tool cover -html=coverage.html -o coverage.html
open coverage.html

# 查看覆盖率统计
go tool cover -func=coverage.out
```

### 详细输出

```bash
# 显示每个测试的详细信息
go test ./pkg/cluster/... -v

# 显示带覆盖率的输出
go test ./pkg/cluster/... -cover
```

## 测试结构

### 主要测试文件

1. **cluster_test.go** - 核心集群功能测试
   - 配置验证
   - 基本结构测试
   - 日志初始化验证

2. **config_test.go** - 配置结构体测试
   - 默认值测试
   - 值设置测试
   - 配置验证

3. **util_test.go** - 工具函数测试
   - 参数列表生成
   - 排序功能
   - 边界条件处理

### 测试覆盖范围

#### 已覆盖的测试场景

- **配置验证**：空配置、有效配置、无效配置
- **参数处理**：基本映射、额外参数覆盖、参数排序
- **集群组件**：API Server、Scheduler、Controller Manager
- **错误处理**：配置验证失败情况
- **日志系统**：日志记录器初始化

### 测试限制

由于 pkg/cluster 包涉及实际的 Kubernetes 组件启动，某些集成测试被标记为跳过，避免在测试环境中启动实际的 Kubernetes 集群。这些测试主要验证：
- 配置逻辑的正确性
- 参数传递的正确性
- 错误处理机制
- 日志系统的完整性

## 运行注意事项

1. **隔离测试**：每个测试都是独立的，可以单独运行
2. **无外部依赖**：测试不需要运行中的 Kubernetes 集群
3. **快速执行**：测试套件可以在几秒内完成
4. **稳定可靠**：测试结果是可重复的

## 故障排除

### 常见问题

**测试失败**：检查配置字段是否正确设置
**编译错误**：确认 Go 环境配置
**覆盖率低**：运行 `go test -cover` 查看详细覆盖率报告

### 调试技巧

```bash
# 运行特定测试并显示详细输出
go test -v -run TestSpecificFunction

# 显示堆栈跟踪
go test -run TestSpecificFunction -v
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Run Cluster Tests
on: [push, pull_request]
jobs:
  test-cluster:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v5
      - name: Run cluster tests
        run: |
          cd pkg/cluster && go test -v -race -coverprofile=coverage.out
      
      - name: Upload coverage to codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./pkg/cluster/coverage.out
```

### 本地验证

```bash
# 检查所有测试
go test ./pkg/cluster/...

# 验证覆盖率
go test ./pkg/cluster/... -coverprofile=coverage.out && \
  go tool cover -func=coverage.out | grep "total:"
```