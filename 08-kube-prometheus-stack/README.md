## 说明
thanos sidecar将Prometheus指标公开为通用Thanos StoreAPI,StoreAPI 是一个通用的 gRPC API，允许 Thanos 组件从各种系统和后端获取指标
Thanos sidecar允许将 Prometheus 保留的所有块备份到磁盘。为了实现这一目标，我们需要确保：
  Sidecar 可以直接访问 Prometheus 数据目录  参数 --tsdb.path
  指定存储桶的配置                          参数 --objstore.config-file
  sidecar 启动时上传已经压缩的块            参数 --shipper.upload-compacted

Querier 组件本质上是一个普通的 PromQL Prometheus 引擎，它从任何实现 Thanos StoreAPI的服务中获取数据。这意味着 Querier 公开 Prometheus HTTP v1 API，以通用 PromQL 语言查询数据。这允许与 Grafana 或 Prometheus API 的其他使用者兼容。
此外，Querier 能够对同一 HA 组中的 StoreAPI 进行重复数据删除。我们将确保将查询器指向所有sidecar 的 gRPC 端点即可进行查询（假设你有两个Prom 集群，一个是单示例，一个是HA架构）,Querier也公开StoreAPI
