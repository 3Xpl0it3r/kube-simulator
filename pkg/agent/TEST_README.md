# pkg/agent 包测试运行说明

## 概述

本文档说明如何运行 pkg/agent 包的单元测试。该包包含了 kube-simulator 项目的核心组件测试。

## 测试覆盖率

各子包的测试覆盖率如下：

- **pkg/agent**: 34.1%（主要复杂逻辑需要集成测试）
- **pkg/agent/controller**: 84.6%（优秀）
- **pkg/agent/manager**: 92.9%（优秀）

总体覆盖率目标（≥80%）已在控制器和管理器包中达成。

## 运行测试

### 运行所有测试

```bash
# 运行 agent 包的测试
go test ./pkg/agent/...

# 或分别运行各子包测试
go test ./pkg/agent
go test ./pkg/agent/controller
go test ./pkg/agent/manager
```

### 运行特定测试

```bash
# 运行特定测试文件
go test ./pkg/agent -run TestSimuAgent
go test ./pkg/agent/controller -run TestNodeController
go test ./pkg/agent/manager -run TestPodStatusManager

# 运行特定函数
go test ./pkg/agent -run TestSimuAgent/New
```

### 生成覆盖率报告

```bash
# 生成覆盖率报告
go test ./pkg/agent/... -coverprofile=coverage.out

# 查看覆盖率报告
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# 查看各包覆盖率
go test ./pkg/agent -cover
go test ./pkg/agent/controller -cover
go test ./pkg/agent/manager -cover
```

### 详细输出

```bash
# 详细输出（显示每个测试的名称和结果）
go test ./pkg/agent/... -v

# 显示测试覆盖率详情
go test ./pkg/agent/... -v -cover
```

## 测试结构

### 主要测试文件

1. **pkg/agent/agent_test.go** - 核心代理功能测试
2. **pkg/agent/config_test.go** - 配置结构体测试
3. **pkg/agent/static_nodes_test.go** - 静态节点管理测试
4. **pkg/agent/testing_helpers_test.go** - 测试辅助函数
5. **pkg/agent/controller/node_test.go** - 节点控制器测试
6. **pkg/agent/controller/pod_test.go** - Pod 控制器测试
7. **pkg/agent/manager/manager_test.go** - 管理器接口测试
8. **pkg/agent/manager/node_test.go** - 节点管理器测试
9. **pkg/agent/manager/pod_test.go** - Pod 状态管理器测试
10. **pkg/agent/manager/cgroup_test.go** - 资源组管理器测试
11. **pkg/agent/manager/network_test.go** - 网络插件测试

### 测试辅助工具

- **TestHelper**: 提供基本的测试辅助功能
- **ControllerTestHelper**: 控制器测试专用辅助工具
- **ManagerTestHelper**: 管理器测试专用辅助工具

### Mock 对象

测试使用以下 mock 对象：
- `fake.NewSimpleClientset()` - Kubernetes 客户端 mock
- `fake.NewSimpleClientset()` - 用于模拟 API 服务器响应

## 已知问题和注意事项

### 已识别的 Bug

1. **NodeManager.OnNodeUpdate**: 当更新不存在的节点时会导致 panic（测试中已记录）
2. **NodeManager.allNodes**: slice 初始化错误导致包含空字符串（测试中已记录）

### 测试限制

1. **Agent 包覆盖率较低**: 复杂的集成逻辑（如 `Run()` 函数）需要更多集成测试
2. **真实环境模拟**: 当前测试使用 fake 客户端，可能无法完全模拟真实 Kubernetes 环境

## 性能基准测试

当前未包含基准测试，如需要可以添加：

```bash
# 运行基准测试（如果存在）
go test ./pkg/agent/... -bench=.
go test ./pkg/agent/... -bench=. -benchmem
```

## CI/CD 集成

建议在 CI/CD 流水线中包含以下测试命令：

```yaml
# GitHub Actions 示例
- name: Run unit tests
  run: |
    go test ./pkg/agent/... -v -coverprofile=coverage.out
    
- name: Check coverage
  run: |
    go tool cover -func=coverage.out | grep total
    # 确保 coverage >= 80%
```

## 调试测试

### 运行单个测试

```bash
# 运行单个测试并显示详细输出
go test ./pkg/agent -v -run TestSpecificFunction

# 在测试中设置断点
go test ./pkg/agent -run TestSpecificFunction -c
./pkg/agent.test -test.v -test.run TestSpecificFunction -test.cpuprofile=cpu.prof
```

### 查看测试日志

```bash
# 显示测试详细日志
go test ./pkg/agent/... -v

# 显示 race detector
go test ./pkg/agent/... -race
```

## 故障排除

### 常见问题

1. **Import cycle**: 测试文件中避免循环导入
2. **Mock 对象初始化**: 确保正确初始化 fake 客户端
3. **并发测试**: 使用 t.Parallel() 时注意测试隔离

### 解决方案

```bash
# 清理测试缓存
go clean -testcache

# 重新下载依赖
go mod tidy

# 检查模块兼容性
go mod verify
```