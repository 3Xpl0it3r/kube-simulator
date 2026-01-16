# kube-simulator

## 项目简介
kube-simulator 是一个轻量级的 Kubernetes 集群模拟器，用于本地开发、测试和学习 Kubernetes。



## 功能特性

- 🚀 快速启动本地 Kubernetes 集群
- 💾 使用 SQLite 作为轻量级存储后端
- 🔄 支持集群重置功能
- 📝 自动生成必要的证书和配置文件
- 🎯 模拟可配置数量的工作节点（默认4个）
- 🔧 支持自定义网络配置（CIDR）

## 系统要求

- Go 1.25.4 或更高版本
- 支持的操作系统：Linux、macOS、Windows

## 安装和使用

### 从源码构建

```bash
# 克隆项目
git clone https://github.com/3Xpl0it3r/kube-simulator.git
cd kube-simulator

# 构建
go build -o kube-simulator cmd/kube-simulator/kube-simulator.go
```

### 基本使用

```bash
# 启动集群（使用默认配置）
./kube-simulator

# 指定监听地址
./kube-simulator --cluster-listen=0.0.0.0:6443

# 重置集群
./kube-simulator --reset

# 自定义节点数量
./kube-simulator --node-num=8

# 指定数据目录
./kube-simulator --data-dir=/path/to/data
```

### 启动后访问

集群启动后，你可以使用生成的 kubeconfig 文件来访问集群：

```bash
# 使用 admin kubeconfig
export KUBECONFIG=./admin.conf
kubectl get nodes
kubectl get pods --all-namespaces
```

## 配置选项

| 参数 | 默认值 | 描述 |
|------|--------|------|
| `--cluster-listen` | `127.0.0.1:6443` | kube-apiserver 监听地址 |
| `--data-dir` | `.data` | 数据存储目录 |
| `--certificate-dir` | `.data/pki` | 证书存储目录 |
| `--etcd-listen` | `127.0.0.1:2379` | etcd 监听地址 |
| `--db-dir` | `.data/db` | 数据库文件目录 |
| `--cluster-cidr` | `10.244.0.0/16` | Pod 网络 CIDR |
| `--service-cidr` | `10.96.0.0/12` | Service 网络 CIDR |
| `--node-num` | `4` | 模拟节点数量 |
| `--reset` | `false` | 重置现有集群 |

## 目录结构

启动后，会在指定目录下生成以下结构：

```
.
├── .data/
│   ├── pki/           # 证书文件
│   ├── db/            # SQLite 数据库
│   ├── controller-manager.yml  # controller-manager kubeconfig
│   ├── scheduler.yml   # scheduler kubeconfig
│   └── admin.conf     # admin kubeconfig
└── kube-simulator     # 可执行文件
```


## 运行测试

```bash
# 运行所有测试
go test ./...
```

## 常见问题

### Q: 如何连接到运行中的集群？
A: 使用生成的 `admin.conf` 文件作为 kubeconfig，或者设置 `KUBECONFIG` 环境变量。

### Q: 如何重置集群？
A: 使用 `--reset` 参数启动程序，或者手动删除 `.data` 目录。

### Q: 支持哪些平台？
A: 支持 Go 语言支持的所有平台，主要测试在 Linux 和 macOS 上进行。

## 许可证

本项目采用 Apache License 2.0 许可证。详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目！

## 相关链接

- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [kine 项目](https://github.com/k3s-io/kine)
