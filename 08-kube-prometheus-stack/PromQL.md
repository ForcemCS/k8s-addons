## Pod-RAM

1. 显示每个 Pod 的当前内存使用情况（以字节为单位）,转换为Mi

   ```
   sum(container_memory_usage_bytes{container!="",namespace = "roh5", pod =~ "game-roh5-server.*"}) by (pod, namespace) / 1048576
   ```

2. 观察过去30m内的内存使用趋势（**加速度是否是定值**）

   + rate() 函数计算指标在指定时间窗口内的每秒变化率 (字节/秒)。[30m] 指定 30 分钟的时间窗口。 计算了过去 30 分钟内每个 Pod**内以容器为单位** 的平均内存使用率增长速度 (字节/秒)。

     ```
     rate(container_memory_usage_bytes{container!="",namespace = "roh5", pod =~ "game-roh5-server.*"}[10m])
     ```

   + sum() 函数对 rate() 的结果进行求和，并按 pod 标签分组（假设pod中有两个容器，对其进行求和）

     ```
     sum(rate(container_memory_usage_bytes{container!="",namespace = "roh5", pod =~ "game-roh5-server.*"}[20m])) by (pod) / 1048576
     ```

3. 检测内存是否存在泄露或者飙升（检测是否超过内存限制的90%）

   ```
   sum(container_memory_working_set_bytes{container!="",namespace = "roh5", pod =~ "game-roh5-server.*"}) by (pod) /
   sum(kube_pod_container_resource_limits{resource="memory",namespace = "roh5", pod =~ "game-roh5-server.*"}) by (pod) > 0.9
   ```

   内网测试

   ```
   sum(container_memory_working_set_bytes{namespace = "harbor",pod=~"harbor-trivy-.*"}) by (pod) /
   sum(kube_pod_container_resource_limits{resource="memory",namespace = "harbor",pod=~"harbor-trivy-.*"}) by (pod) > 0.0001
   ```

4. 检测由于oom终止的Pod

   ```
   kube_pod_container_status_last_terminated_reason{reason="OOMKilled"} == 1
   ```

## Pod-CPU

container_cpu_usage_seconds_total 计数器类型的指标

1. 查看某个pod中容器的1m钟内的cpu使用率

   ```
   rate(container_cpu_usage_seconds_total{namespace = "roh5",pod= "game-roh5-server-10001-688cdbf5c8-2nkzc"}[1m]) * 1000
   ```

2. 在过去的5分钟内平均每秒使用的CPU率（以Pod为基准进行求和，类似于kubectl  top pods xxxx）

   ```
   sum(rate(container_cpu_usage_seconds_total{container!="",pod=~"game-roh5-server.*"}[5m])) by (pod)
   ```

3. CPU 使用率占分配资源的百分比

   ```
   sum(rate(container_cpu_usage_seconds_total{container!="",pod=~"game-roh5-server.*"}[5m])) by (pod) /
   sum(kube_pod_container_resource_limits{resource="cpu",container!="",pod=~"game-roh5-server.*"}) by (pod) 
   ```

4. 查找消耗 CPU 资源最多的前 5 个 Pod

   ```
   topk(3, sum(rate(container_cpu_usage_seconds_total{container!="",pod=~"game-roh5-server.*"}[5m])) by (pod))
   ```

## 备注

```
    - action: replace
      regex: kube_pod_container_resource_limits;cpu;core
      replacement: kube_pod_container_resource_limits_cpu_cores
      sourceLabels:
      - __name__
      - resource
      - unit
      targetLabel: __name__
```

这个配置的基本功能是：如果 Prometheus 发现某个指标名称是 `kube_pod_container_resource_limits`，并且它的 `resource` 标签是 `cpu`，它的 `unit` 标签是 `core`，那么它会将该指标的名称替换为 `kube_pod_container_resource_limits_cpu_cores`。这样的操作使得在查询和监控时，指标名称更加直观、统一，并且明确其所代表的含义。