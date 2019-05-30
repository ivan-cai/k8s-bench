# k8s-bench
benchmark for kubernetes

### 实现
使用client-go发送请求

### 打包
cd ivan-cai/k8s-bench
go build

### 命令
#### 创建pod的命令
```
 ./k8s-bench -n 10000 -c 200 -H  "Authorization: Bearer ${token}" -p pod.yaml -K /root/.kube/config https://172.30.123.10:60002/api/v1/namespaces/default/pods
```

### 版本变化
- 0.0.0
  - 仅仅支持创建pod的压测
