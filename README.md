## 安装
```bash
git clone https://github.com/3Xpl0it3r/kube-simulator.git
cd kube-simulator
go mod tidy && go mod vendor
bash scripts/build.sh
# 默认启动一个包含4个节点的k8s集群
./kube-simulator --cluster-listen <ip>:6443 
# 启动完会在当前目录下生成一个admin.conf 配置文件,可以通过这个admin.conf配置文件来管理k8s集群
```

