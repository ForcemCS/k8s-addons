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

5. 查看容器是否频繁受到 CPU 节流，如果部署cpu limit的原因，应该分析应用程序代码

   ```
   sum(rate(container_cpu_cfs_throttled_seconds_total{namespace=~"hdh5", pod=~"game-hdh5-server-10003-6b8db8dcbf-nlvh4", image!="", container!="", cluster="cls-po5g6rh8"}[$__rate_interval])) by (container)
   ```
## Ingress-Nginx
1. 请求总数和速率 (Request Rate / QPS)
每秒钟有多少 HTTP 请求通过 ingress-nginx 到达你的后端服务
```
sum(rate(nginx_ingress_controller_requests[1m])) by (path ,host, status)
```

每秒请求速率(次数)，按路径,域名和响应码分类

QPS 突然暴跌可能意味着登录服务故障、支付渠道中断、服务器宕机或玩家大量流失。QPS 异常飙升可能预示着遭受 DDoS 攻击或刷量行为。


2.请求延迟 (Request Latency - P95/P99)
95% 或 99% 的请求从 ingress-nginx 收到到返回响应给客户端所花费的时间。这代表了绝大多数用户的延迟体验。
Ingress 层面观察到的高延迟，通常暗示着后端 服务处理慢（逻辑复杂、数据库查询慢、锁竞争）、网络传输慢（跨区、网络拥堵）或 Ingress 本身资源不足。它是追踪“卡顿”问题的重要起点。
```
histogram_quantile(0.95, sum(rate(nginx_ingress_controller_request_duration_seconds_bucket[5m])) by (le, ingress))
```

3.网络流量 (Bandwidth)
  nginx_ingress_controller_bytes_sent (Counter) - 发送给客户端的总字节数。
  nginx_ingress_controller_request_size (Histogram/Summary) - 接收到的请求体大小。
  nginx_ingress_controller_response_size (Histogram/Summary) - 发送的响应体大小。

PromQL (计算指定游戏服务的出口带宽 - Bytes/sec):
```
sum(rate(nginx_ingress_controller_bytes_sent{service="<your-game-svc-name>", namespace="<your-game-namespace>"}[5m]))
```
识别异常的大响应（如过大的 JSON/HTML）。
```
histogram_quantile(0.99, sum(rate(nginx_ingress_controller_response_size_bucket[5m])) by (le, ingress))
```

计算指定游戏服务的平均响应大小 - Bytes
```
sum(rate(nginx_ingress_controller_response_size_sum{service="<your-game-svc-name>", namespace="<your-game-namespace>"}[5m]))
/
sum(rate(nginx_ingress_controller_response_size_count{service="<your-game-svc-name>", namespace="<your-game-namespace>"}[5m]))
```

过高的带宽使用可能占满服务器网卡或网络链路，间接导致延迟增加或丢包，影响玩家体验。尤其是在玩家密集区域或进行大规模团战时。

4.连接数 (Connections)
指标: nginx_ingress_controller_connections_active (Gauge) - 当前活动的客户端连接数。
目的: 了解并发连接水平。游戏服务（尤其是长连接类型）可能会维持大量并发连接。

PromQL (查看指定游戏服务 Ingress Pod 的总活跃连接数):
```
sum(nginx_ingress_controller_connections_active{class="nginx", service="<your-game-svc-name>", namespace="<your-game-namespace>"})
```

入口资源消耗: 每个连接都需要消耗 ingress-nginx 的内存和文件句柄资源。过高的连接数可能导致 Ingress Controller 性能下降或崩溃。
后端压力前兆: Ingress 层的连接数是后端 Skynet 服务需要处理的连接数的前哨。监控它可以预警后端连接池或处理能力的压力。

6.请求转发到后端的延迟
```
histogram_quantile(0.95, sum(rate(nginx_ingress_controller_ingress_upstream_latency_seconds_bucket[5m])) by (le, ingress))
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
