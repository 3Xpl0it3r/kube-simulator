# Change: 为 pkg/agent 包添加单元测试

## Why
pkg/agent 包是 kube-simulator 项目的核心组件，负责模拟节点和 Pod 的生命周期管理。目前该包缺乏单元测试，可能导致代码质量下降和潜在的回归问题。添加单元测试可以提高代码可靠性，确保新功能不会破坏现有行为，并提升开发团队的信心。

## What Changes
- 为 pkg/agent 包下的所有 Go 文件创建对应的单元测试文件
- 测试覆盖核心功能：SimuAgent 结构体、控制器（NodeController、PodController）、管理器（NodeManager、PodStatusManager）以及辅助功能
- 添加 mock 对象用于 Kubernetes 客户端接口测试
- 确保测试覆盖率至少达到 80%
- 添加测试数据和测试辅助函数

## Impact
- 受影响的代码：pkg/agent/ 目录下所有文件
- 新增文件：*_test.go 测试文件
- 测试框架：使用 Go 标准库 testing 包，可能需要 testify/mock 用于复杂 mock 场景
- 构建影响：增加测试执行时间，但不会影响生产构建
- 开发流程：开发者需要在提交代码前运行相关测试