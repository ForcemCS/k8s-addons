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
  #每秒请求速率(次数)，按路径,域名和响应码分类
  #QPS 突然暴跌可能意味着登录服务故障、支付渠道中断、服务器宕机或玩家大量流失。QPS 异常飙升可能预示着遭受 DDoS 攻击或刷量行为。
  ```

2. 请求延迟 (Request Latency - P95/P99)

   95% 或 99% 的请求从 ingress-nginx 收到到返回响应给客户端所花费的时间。这代表了绝大多数用户的延迟体验。

   ```
   histogram_quantile(0.95, sum(rate(nginx_ingress_controller_request_duration_seconds_bucket[5m])) by (le, ingress))
   #Ingress 层面观察到的高延迟，通常暗示着后端 服务处理慢（逻辑复杂、数据库查询慢、锁竞争）、网络传输慢（跨区、网络拥堵）或 Ingress 本身资源不足。它是追踪“卡顿”问题的重要起点。
   ```

3. 网络流量 (Bandwidth)

   **`nginx_ingress_controller_bytes_sent_sum` (传出/下行带宽 - 服务器到客户端)**

   - **含义**: 这个指标是一个 **Counter** (计数器)，记录了自 Nginx Ingress Controller 启动以来，通过它**发送给客户端**的总字节数。这通常是你最关心的带宽，因为它代表了游戏服务器向玩家发送游戏状态、资源等数据的量。

     ```
     sum(
       rate(nginx_ingress_controller_bytes_sent_sum{namespace = "ingress-nginx",host = "server-hd.dhysr.soletower.com" , ingress= "game-hdh5-server-10001-ingress-10003"}[5m])
     ) / (1024 * 1024)
     ```

   **`nginx_ingress_controller_request_size_sum` (传入/上行带宽 - 客户端到服务器)**

   - **含义**: 这个指标也是一个 **Counter**，记录了自 Nginx Ingress Controller 启动以来，从客户端**接收到的请求体**的总字节数。这代表了玩家向游戏服务器发送操作指令等数据的量。对于某些类型的游戏，这个值可能也比较重要，但通常远小于传出带宽。

     ```
     sum by (ingress)(
       rate(nginx_ingress_controller_request_size_sum{namespace = "ingress-nginx",host = "server-hd.dhysr.soletower.com" , ingress= "game-hdh5-server-10001-ingress-10003"}[5m])
     ) / (1024 * 1024)
     ```

4. 连接数 (Connections)

   **`nginx_ingress_controller_nginx_process_connections`**:

   它通过 `state` 标签来区分不同状态的连接：

   - `state="active"`: 当前总的活跃连接数。
   - `state="reading"`: 正在读取请求头的连接数。
   - `state="writing"`: 正在写入响应的连接数。
   - `state="waiting"`: 处于空闲/等待状态的连接数 (Keep-Alive, idle WebSocket)。

   1. **监控总活跃连接数 (跨所有 Pod):**

      ```promql
      sum(nginx_ingress_controller_nginx_process_connections{state="active", controller_namespace="ingress-nginx", controller_class="k8s.io/game-nginx"})
      ```

      *(请确保 `controller_namespace` 和 `controller_class` 与你的环境匹配，用于选定正确的 Ingress Controller 实例)*

   2. **监控总等待连接数 (用于推断长连接):**

      ```promql
      sum(nginx_ingress_controller_nginx_process_connections{state="waiting", controller_namespace="ingress-nginx", controller_class="k8s.io/game-nginx"})
      ```

   3. **监控等待连接占活跃连接的比例 (用于推断长连接):**

      ```promql
      sum(nginx_ingress_controller_nginx_process_connections{state="waiting", controller_namespace="ingress-nginx", controller_class="k8s.io/game-nginx"})
      /
      sum(nginx_ingress_controller_nginx_process_connections{state="active", controller_namespace="ingress-nginx", controller_class="k8s.io/game-nginx"})
      ```

      *(如果这个比例很高，说明长连接/空闲连接占比较大)*

   4. **监控连接建立速率 (用于推断长连接):**

      ```promql
      sum(rate(nginx_ingress_controller_nginx_process_connections_total{state="accepted", controller_namespace="ingress-nginx", controller_class="k8s.io/game-nginx"}[1m]))
      ```

   5. **推断长连接**: 依然需要通过观察 `state="active"` 的值，并结合 `state="waiting"` 的比例以及 `state="accepted"` 的速率来间接判断长连接的情况。

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
