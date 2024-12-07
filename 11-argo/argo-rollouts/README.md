## Blue/Green部署

蓝/绿部署是最小化部署停机时间的最简单方法之一。蓝/绿部署并非 Kubernetes 特有，甚至可用于驻留在虚拟机上的传统应用程序。

1. 一开始，应用程序的所有用户都会被路由到当前版本（显示为蓝色）。关键是所有流量都从负载平衡器通过。
2. 部署新版本（显示为绿色）。该版本不接收任何实时流量，因此所有用户仍由先前/稳定版本提供服务
3. 开发人员可以在内部测试新颜色并验证其正确性。如果正确，流量就会切换到新版本
4. 如果一切顺利，旧版本将被完全丢弃。我们又回到了初始状态（颜色的顺序并不重要）。

这种模式的主要好处是，如果新版本在任何时候出现问题，所有用户都可以（通过负载平衡器）切换回以前的版本。切换负载平衡器比重新部署新版本要快得多，从而将对现有用户的影响降到最低。

这种模式有多种变体。在某些情况下，旧版本不会被销毁，而是继续在后台运行。您还可以在线保留旧版本（可能占用空间较小），以便轻松切换到任何以前的应用程序版本。

## Canary 部署

蓝/绿部署可以最大限度地减少部署后的停机时间，但并不完美。如果您的新版本有一个隐藏的问题，在一段时间后才显现出来（即您的烟雾测试没有检测到），那么您的所有用户都会受到影响，因为流量切换是全有或全无。

一种改进的部署方法是金丝雀部署。它的功能类似于蓝/绿部署，但不是一次性将 100% 的实时流量切换到新版本，而是只移动一部分用户。

## Argo Rollouts介绍

Argo Rollouts 是专为 Kubernetes 设计的渐进式交付控制器。它采用循序渐进的部署方式，而不是 “一蹴而就 ”的方式，让您能够以最少/零停机时间部署应用程序。

Argo Rollouts 可为您的 Kubernetes 集群增效，除了滚动更新外，您现在还可以进行以下操作
+ 蓝/绿部署
+ 金丝雀部署
+ A/B 测试
+ 自动回滚
+ Integrated Metric analysis

Argo Rollouts 的改进 - 预览服务 (Preview Service): Argo Rollouts 不仅仅是简单的切换。在切换流量到绿色（新版本）之前，它会创建一个额外的“preview” Kubernetes service。这个预览服务指向绿色版本，但不会接收生产流量。

一段时间后（具体时间可在 Argo Rollouts 中配置），旧版本会被完全缩减（以保护资源）

```shell
假设我们有了第一个版本，现在使用下边的命令，进行版本升级
root@kubernetes-vm:~/workdir# kubectl argo rollouts set image simple-rollout webserver-simple=docker.io/kostiscodefresh/gitops-canary-app:v2.0
rollout "simple-rollout" image updated
root@kubernetes-vm:~/workdir# kubectl argo rollouts get rollout simple-rollout
Name:            simple-rollout
Namespace:       default
Status:          ॥ Paused
Message:         BlueGreenPause
Strategy:        BlueGreen
Images:          docker.io/kostiscodefresh/gitops-canary-app:v1.0 (stable, active)
                 docker.io/kostiscodefresh/gitops-canary-app:v2.0 (preview)
Replicas:
  Desired:       2
  Current:       4
  Updated:       2
  Ready:         2
  Available:     2

NAME                                        KIND        STATUS     AGE    INFO
⟳ simple-rollout                            Rollout     ॥ Paused   3m45s  
├──# revision:2                                                           
│  └──⧉ simple-rollout-597df99d85           ReplicaSet  ✔ Healthy  30s    preview
│     ├──□ simple-rollout-597df99d85-kct7w  Pod         ✔ Running  30s    ready:1/1
│     └──□ simple-rollout-597df99d85-p9drd  Pod         ✔ Running  30s    ready:1/1
└──# revision:1                                                           
   └──⧉ simple-rollout-b68b5bffb            ReplicaSet  ✔ Healthy  3m45s  stable,active
      ├──□ simple-rollout-b68b5bffb-bgtr6   Pod         ✔ Running  3m44s  ready:1/1
      └──□ simple-rollout-b68b5bffb-gvr2x   Pod         ✔ Running  3m44s  ready:1/1
```

更改镜像后，会发生以下情况

+ Argo Rollouts 使用新版本创建另一个副本集
+ 旧版本仍然存在，并获得live/active流量
+ ArgoCD 会将应用程序标记为不同步
+ ArgoCD 还会将应用程序的健康状况标记为 “suspended”，因为我们已将新颜色设置为等待
  请注意，尽管我们已经部署了应用程序的下一个版本，但所有实时流量都会流向旧版本。

你可以查看 “live traffic”选项卡来验证这一点。

```
kubectl argo rollouts promote simple-rollout
kubectl argo rollouts get rollout simple-rollout --watch
```

一段时间后，您就会看到旧版本的 pod 被销毁。从 “实时流量 ”选项卡中可以看到，现在所有实时流量都转到了新版本。

Argo Rollouts 上边所述的基本canary 模式，还提供了丰富的自定义选项。最重要的新增功能之一是，除了用于实时流量的服务外，还能通过引入 “preview”Kubernetes service来 “测试 ”即将推出的版本。执行部署的团队可以使用该预览服务，在新版本被实时用户子集使用时对其进行验证。

现在有三个 Kubernetes 服务。

+ rollout-canary-all-traffic 服务正在捕获来自应用程序实际用户的所有实时流量。
+ rollout-canary-preview，它只会将流量导向金丝雀/新版本。
+ rollout-canary-active。它将始终指向软件的stable/previous一版本。

正常情况下，所有 3 个服务都指向同一个版，一旦开始部署，就会创建一个新的 “版本”。此时rollout-canary-all-traffic流量可以按照比例发送到新版本和旧版本

其余的两种情况不变

蓝/绿部署可以在任何 Kubernetes 集群上运行。但对于金丝雀（Canary）部署，你需要一个智能服务层，它可以将流量逐渐转移到canary pods ，同时仍将其余流量保留到old/stable的吊舱。为此，Argo Rollouts 支持多种服务网格和网关。我们将使用流行的 Kubernetes Ambassador API Gateway，在金丝雀和旧版/稳定版之间拆分实时流量。

```
kubectl argo rollouts set image simple-rollout webserver-simple=docker.io/kostiscodefresh/gitops-canary-app:v2.0
kubectl argo rollouts get rollout simple-rollout
```

更改镜像后，会发生以下情况

Argo Rollouts 使用新版本创建另一个副本集
旧版本仍然存在，并获得实时/活动流量
金丝雀版本获得 30% 的实时流量。
ArgoCD 会将应用程序标记为不同步
ArgoCD 还会将应用程序的健康状况标记为 “暂停”，因为我们已将新颜色设置为suspended
请注意，即使我们已经部署了应用程序的下一个版本，实时流量还是会同时流向新/旧版本。您可以查看 “实时流量 ”选项卡来验证这一点。

要手动推广部署并将 60% 切换到新版本，请输入

```
kubectl argo rollouts promote simple-rollout
kubectl argo rollouts get rollout simple-rollout --watch
```

## Automated Rollbacks with Metrics

虽然您可以在不同阶段之间使用简单pauses 的金丝雀，但 Argo Rollouts 提供了强大的功能，可以查看应用程序指标并自动决定是否继续部署。

这种方法背后的理念是实现金丝雀部署的完全自动化。您可以设置不同的阈值来定义部署是否 “成功”，而无需人工运行smoke 测试或查看图表。

Argo Rollouts 已支持多个指标提供程序，如 Prometheus、DataDog、NewRelic、Cloudwatch 等。

```
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
spec:
  args:
  - name: service-name
  metrics:
  - name: success-rate
    interval: 2m
    count: 2
    # NOTE: prometheus queries return results in the form of a vector.
    # So it is common to access the index 0 of the returned array to obtain the value
    successCondition: result[0] >= 0.95
    provider:
      prometheus:
        address: http://prom-release-prometheus-server.prom.svc.cluster.local:80
        query: sum(response_status{app="",role="canary",status=~"2.*"})/sum(response_status{app="",role="canary"}
```

该 Prometheus 分析报告指出 ：将为 Prometheus 获取响应状态指标（200 个响应与总响应）。如果结果大于 0.95，Deployment 将继续成功进行。如果结果小于 0.95，那么金丝雀将失败，整个部署将自动回滚。
