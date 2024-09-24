## 安装

对于victoria-metrics-k8s-stack这个大的堆栈所依赖的Chart，请读者阅读官方文档

![VM-stack-1](./VM-stack-1.png)

```
helm install -n monitoring  vm  --version 0.23.3   vm/victoria-metrics-k8s-stack   -f ./victoria.yaml
```

