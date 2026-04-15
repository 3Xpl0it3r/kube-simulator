# kube-simulator

## 项目简介
kube-simulator 是一个轻量级的 Kubernetes 集群模拟器，用于本地开发、测试和学习 Kubernetes。

- Go 1.25.4 或更高版本
- 支持的操作系统：Linux、macOS、Windows

## 安装和使用

### 从源码构建

```bash
# 克隆项目
git clone https://github.com/3Xpl0it3r/kube-simulator.git
cd kube-simulator

# 执行完build.sh 在当前目录下会生成 kubectl 和 kube-simulator 两个二进制文件
bash script/build.sh

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

