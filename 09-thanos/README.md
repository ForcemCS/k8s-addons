## 说明
1. 借助 Thanos sidecar 将 Prometheus 数据无缝上传到廉价的对象存储。
2. 借助 Thanos Store Gateway 进一步查询对象存储中的数据
3. 借助 Thanos Store Gateway 进一步查询对象存储中的数据。
4. 如何通过 Thanos Querier 轻松查询新数据和旧数据。

所有这些都允许您将指标保存在廉价且可靠的对象存储中，从而为 Prometheus 提供几乎无限的指标保留

在 Prometheus 中维护长久的数据是可行的，但并不容易。调整数据大小、备份或长期维护这些数据是很棘手的。最重要的是，Prometheus 不执行任何复制，因此 Prometheus 的任何不可用都会导致查询不可用。我们可以允许 Thanos Sidecar 持续上传由 Prometheus 定期保存到磁盘的指标块。
注意：Prometheus 在抓取数据时，最初会聚合内存和 WAL（on-disk write-head-log）中的所有样本。仅在 2-3 小时后，它才会以 2 小时 TSDB 块的形式将数据“压缩”到磁盘中。这就是为什么我们仍然需要查询 Prometheus 来获取最新数据，但总的来说，通过此更改，我们可以将 Prometheus 保留率保持在最低限度。在这种情况下，建议将 Prometheus 保留时间至少为 6 小时，以便为潜在的网络分区事件提供安全缓冲。

##Thanos组件
所有的组件都是通过thanos镜像实现的，只是不同的参数实现不同的功能

1.Store Gateway
该组件在对象存储桶中的历史数据之上实现 Store API。它主要充当 API 网关，因此不需要大量本地磁盘空间，我们可以通过querier指定参数Store Gateway gRPC 地址（--store ip:port） 来查询
它在本地磁盘上保留有关所有远程块的少量信息，并使其与存储桶保持同步。通常可以安全地在重新启动后删除这些数据，但代价是增加启动时
现在，Querier同时查询 Prometheus 实例(thanos sidecar)和 Thanos Store，从而提供对存档块和实时指标的透明访问。 Query 中使用的普通 PromQL Prometheus 引擎可以推断出我们需要获取数据的时间序列和时间范围。Prometheus+sidecar 结果和存储网关之间潜在的数据重复将被透明处理，并且在结果中不可见
思考一个问题：在使用 Thanos sidecar 的 HA Prometheus 设置中，多个 sidecar 尝试将相同的数据块上传到对象存储是否会出现问题？

2.Thanos Compactor
它对 TSDB 块数据对象存储应用压缩、保留、删除和下采样操作。(Thanos Compactor 目前设计为作为单例运行。不可用（小时）不是问题，因为它不满足任何实时请求。)
Compactor是一个重要组件，可对单个对象存储桶进行压缩、下采样，并对其中的 TSDB 块应用保留功能，从而使历史数据的查询更加高效。它创建旧指标的聚合（基于规则）
它还负责数据的下采样，40小时后执行5m下采样，10天后执行1h下采样
注意：如果您使用对象存储，则 Thanos Compactor 是必需的，否则 Thanos Store Gateway 在不使用 Compactor 的情况下会太慢。

有几个参数的说明
The flag wait is used to make sure all compactions have been processed while --wait-interval is kept in 30s to perform all the compactions and downsampling very quickly. Also, this only works when when --wait flag is specified. Another flag --consistency-delay is basically used for buckets which are not consistent strongly. It is the minimum age of non-compacted blocks before they are being processed. Here, we kept the delay at 0s assuming the bucket is consistent.

Compaction and Downsampling（压缩和下采样）
当我们查询大量历史数据时，我们需要遍历数百万个样本，这使得当我们检索一年的数据时查询速度越来越慢，因此，Thanos 使用称为下采样的技术（降低信号采样率的过程）来保持查询响应，并且不需要特殊配置来执行此过程。
Compactor对桶数据进行压缩，同时完成历史数据的下采样。


有关更多信息请[参考](https://thanos.io/tip/components/compact.md/)

3.Query Frontend
在Querier前面部署和配置 Thanos Query Frontend，以便我们可以缓存查询响应,以实现低延迟查询
假设我们必须向多个团队提供集中式指标平台。对于每个团队，我们都会有一个专门的 Prometheus。这些可以位于相同环境或不同环境（数据中心、区域、集群等）。

如果我们能在查询前设置一个入口点，而不是单独的查询器，会怎么样？这样，我们就可以根据时间对查询进行切分，并在查询器之间进行分配，以平衡负载？此外，为什么不缓存这些响应，以便下次有人询问相同的时间范围时，我们可以直接从内存中提供。这样不是更快吗？


