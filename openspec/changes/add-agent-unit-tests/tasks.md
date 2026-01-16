## 1. 测试环境准备
- [x] 1.1 检查现有测试框架和依赖
- [x] 1.2 创建测试辅助函数和 mock 对象
- [x] 1.3 设置测试数据结构

## 2. 核心组件测试
- [x] 2.1 为 pkg/agent/agent.go 添加单元测试
  - [x] 2.1.1 测试 Run() 函数的初始化逻辑
  - [x] 2.1.2 测试 SimuAgent 结构体的方法
  - [x] 2.1.3 测试 mainLoop() 事件处理
- [x] 2.2 为 pkg/agent/config.go 添加单元测试
  - [x] 2.2.1 测试 Config 结构体验证
- [x] 2.3 为 pkg/agent/static_nodes.go 添加单元测试
  - [x] 2.3.1 测试 registerBootstrapNode() 函数
  - [x] 2.3.2 测试 joinNewNode() 函数

## 3. 控制器测试
- [x] 3.1 为 pkg/agent/controller/node.go 添加单元测试
  - [x] 3.1.1 测试 NewNodeController() 构造函数
  - [x] 3.1.2 测试节点事件处理（Add、Update、Delete）
  - [x] 3.1.3 测试 Run() 方法的启动逻辑
- [x] 3.2 为 pkg/agent/controller/pod.go 添加单元测试
  - [x] 3.2.1 测试 NewPodController() 构造函数
  - [x] 3.2.2 测试 Pod 事件处理（Add、Update、Delete）
  - [x] 3.2.3 测试 Pod 事件过滤逻辑

## 4. 管理器测试
- [x] 4.1 为 pkg/agent/manager/manager.go 添加单元测试
  - [x] 4.1.1 测试 Manager 接口实现
- [x] 4.2 为 pkg/agent/manager/node.go 添加单元测试
  - [x] 4.2.1 测试 NewNodeManager() 构造函数
  - [x] 4.2.2 测试节点状态管理方法
  - [x] 4.2.3 测试 Pod 事件对节点状态的影响
  - [x] 4.2.4 测试节点租约同步逻辑
- [x] 4.3 为 pkg/agent/manager/pod.go 添加单元测试
  - [x] 4.3.1 测试 NewPodStatusManager() 构造函数
  - [x] 4.3.2 测试 Pod 状态更新逻辑
  - [x] 4.3.3 测试 IP 地址分配和释放
  - [x] 4.3.4 测试容器状态模拟
- [x] 4.4 为 pkg/agent/manager/cgroup.go 添加单元测试
  - [x] 4.4.1 测试资源计算和管理
  - [x] 4.4.2 测试资源充足性检查
- [x] 4.5 为 pkg/agent/manager/network.go 添加单元测试
  - [x] 4.5.1 测试 CNIPlugin IP 地址分配
  - [x] 4.5.2 测试 IP 地址释放逻辑
  - [x] 4.5.3 测试网络 CIDR 解析

## 5. 集成测试和覆盖率
- [x] 5.1 添加端到端测试场景
- [x] 5.2 验证测试覆盖率目标（≥80%）
- [x] 5.3 添加基准测试（如需要）
- [x] 5.4 确保所有测试通过 CI/CD 流水线

## 6. 文档和维护
- [x] 6.1 添加测试运行说明
- [x] 6.2 更新项目文档中的测试策略
- [x] 6.3 验证测试可重复性和稳定性