---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 304402207d31a971b6c1ccc7349cb72ea741fa7a9e745fada68018f5cd09be0dcf04772b0220196902913aede7d190ef2f0fd453bbc7f68468dbedddaa47a18f4e2580814cca
    ReservedCode2: 3046022100fac4af432b614ae683afcbe3504fcc66aa32d65390a4343358d4a1f691c276da022100fb843fda301d48f0ec93740a419dcec3a063259c10ec65cd28a9edf4a1a681d7
---

# K8s Manager

一个简洁的 Kubernetes 集群管理 CLI 工具。

## 安装

```bash
# 编译
cd k8s-manager
go build -o k8s-manager .

# 或安装到系统
go install .
```

## 配置

设置 `KUBECONFIG` 环境变量指向你的 kubeconfig 文件：

```bash
export KUBECONFIG=~/.kube/config
```

默认路径: `~/.kube/config`

## 使用方法

### 查看资源

```bash
# 查看所有节点
./k8s-manager nodes

# 查看命名空间
./k8s-manager ns

# 查看 Pod（默认 default 命名空间）
./k8s-manager pods
./k8s-manager pods kube-system

# 查看 Deployment
./k8s-manager deploy
./k8s-manager deploy kube-system

# 查看 Service
./k8s-manager svc
./k8s-manager svc kube-system
```

### 管理操作

```bash
# 扩缩容 Deployment
./k8s-manager scale my-deployment 3
./k8s-manager scale my-deployment 3 default

# 查看 Pod 日志
./k8s-manager logs my-pod
./k8s-manager logs my-pod kube-system
```

## 命令列表

| 命令 | 说明 |
|------|------|
| `nodes` | 列出所有节点 |
| `ns` | 列出所有命名空间 |
| `pods [ns]` | 列出 Pod |
| `deploy [ns]` | 列出 Deployment |
| `svc [ns]` | 列出 Service |
| `scale <name> <replicas> [ns]` | 扩缩容 |
| `logs <pod> [ns]` | 查看日志 |

## 示例输出

```
$ ./k8s-manager nodes
NAME               STATUS           ROLES               AGE
master             Ready            control-plane,master   30d
worker-1           Ready            worker              30d
worker-2           Ready            worker              30d
```

```
$ ./k8s-manager pods kube-system
NAME                               READY         STATUS      AGE
coredns-6d4d8598f-k7x9x           1/1           Running    30d
coredns-6d4d8598f-xq2p4           1/1           Running    30d
etcd-master                       1/1           Running    30d
kube-apiserver-master             1/1           Running    30d
```
