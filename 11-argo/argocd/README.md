## The reconciliation loop
Argo CD 自动同步应用程序的时间可以手动配置，但是又局限性，我们可以使用GIT Commit的方式
[Git Webhook](https://argo-cd.readthedocs.io/en/stable/operator-manual/webhook/)

使用 webhook 非常高效，因为现在当你向 Git 提交东西时，你的 Argo CD 安装绝不会延迟。如果只使用默认的轮询方式，那么 Argo CD 可能需要等待长达 3 分钟（或您设置的任何同步时间）才能检测到更改。有了 Webhook，只要 Git 有任何变动，Argo CD 就会立即运行同步。

## Application Health
可能的值有:
“Healthy”     -> Resource is 100% healthy
“Progressing” -> 资源不健康，但仍有机会达到健康状态
“Suspended”   -> 资源已暂停或暂停。典型的例子是 cron 作业
“Missing”     -> 资源不在群集中
“Degraded”    -> 资源状态显示失败或资源无法及时达到健康状态
“Unknown”     -> 健康评估失败，实际健康状态未知

对于自定义 Kubernetes 资源，健康状况是在 Lua 脚本中定义的。具有自定义健康状况定义的资源包括

- Argo Rollout (and associated Analysis and Experiments)
- Bitnami Sealed secrets
- Cert Manager
- Elastic Search
- Jaeger
- Kafka
- CrossPlane provider

您可以在 https://github.com/argoproj/argo-cd/tree/master/resource_customizations 查看所有自定义健康检查的列表。Lua 是一种独立的脚本语言，通过在 Lua 中实施检查，您可以为自定义资源添加自己的健康检查。

下面是 Argo Rollouts Experiments 健康检查的一个[简单示例](https://github.com/argoproj/argo-cd/blob/master/resource_customizations/argoproj.io/Experiment/health.lua)

```lua
local hs = {}
if obj.status ~= nil then
    if obj.status.phase == "Pending" then
        hs.status = "Progressing"
        hs.message = "Experiment is pending"
    end
    if obj.status.phase == "Running" then
        hs.status = "Progressing"
        hs.message = "Experiment is running"
    end
    if obj.status.phase == "Successful" then
        hs.status = "Healthy"
        hs.message = "Experiment is successful"
    end
    if obj.status.phase == "Failed" then
        hs.status = "Degraded"
        hs.message = "Experiment has failed"
    end
    if obj.status.phase == "Error" then
        hs.status = "Degraded"
        hs.message = "Experiment had an error"
    end
    return hs
end

hs.status = "Progressing"
hs.message = "Waiting for experiment to finish: status has not been reconciled."
return hs
```

Argo CD 健康检查完全独立于 Kubernetes 健康探测器。

## Sync Strategies

请[参考](https://argo-cd.readthedocs.io/en/stable/user-guide/auto_sync/)

在定义同步策略时，有 3 个参数可以更改：

1. **Manual or automatic sync**.
2. **Auto-pruning of resources** - this is only applicable for automatic sync.
3. **Self-Heal of cluster**(集群自愈) - this is only applicable for automatic sync.

手动或自动同步定义了 Argo CD 在 Git 中发现应用程序新版本时的操作。如果设置为自动，Argo CD 将应用更改，然后在集群中更新/创建新资源。如果设置为手动，Argo CD 将检测到更改，但不会更改集群中的任何内容。

自动剪枝（Auto-pruning）定义了当你从 Git 中移除/删除文件时 Argo CD 会做什么。如果启用，Argo CD 也会删除集群中的相应资源。如果禁用，Argo CD 将永远不会删除集群中的任何内容。

自愈定义了当你直接对集群（通过 kubectl 或其他方式）进行更改时，Argo CD 会做什么。请注意，如果要遵循 GitOps 原则（因为所有更改都应通过 Git），则不建议在集群中进行手动更改。如果启用，Argo CD 就会丢弃多余的更改，并将集群恢复到 Git 中描述的状态。

这意味着你可以对所有设置进行多种组合，如下表所示：

| Policy        | A      | B        | C        | D        | E       |
| ------------- | ------ | -------- | -------- | -------- | ------- |
| Sync Strategy | Manual | Auto     | Auto     | Auto     | Auto    |
| Auto-prune    | N/A    | Disabled | Enabled  | Disabled | Enabled |
| Self-heal     | N/A    | Disabled | Disabled | Enabled  | Enabled |

策略 A（Argo CD 不做任何事情）可能是您开始采用 Argo CD 的方式，尤其是当您想在现有项目中应用 GitOps 时。这是您了解 Argo CD 如何工作的机会，不会对您的部署造成实际影响。

策略 B（自动同步）是利用 GitOps 自动化功能的第一步。只要 Git 发生任何变化，集群就会自动更新。禁用自动修剪意味着仍需手动删除资源，禁用自愈意味着仍可对集群进行手动更改（如果想有一个迁移期）。

策略 C 和 D 是正交的，可以作为通向策略 E 的垫脚石。
在该策略下，所有事情都是双向自动化的。Git 中的更改也会自动反映到集群中（包括删除资源），而集群中的手动更改则会被直接丢弃。

## Managing Secrets

我们将使用 [Bitnami Sealed secrets](https://github.com/bitnami-labs/sealed-secrets) 控制器，在将秘密存储到 Git 之前对其进行加密，并在将其传递给应用程序之前对其进行解密。

在集群外部对它们进行加密，这样我们就能按照 GitOps 的创始原则之一（所有东西都存储在 Git 中）

Sealed secrets只是建立在 Kubernetes 本机机密之上的扩展。这意味着，在encryption/decryption完成后，所有秘密都能像普通 Kubernetes 秘密一样运行，这就是你的应用程序访问它们的方式。如果你不喜欢纯 Kubernetes 秘密的工作方式，那么你需要找到一种替代的安全机制。

Bitnami Sealed secrets 控制器是你安装在集群上的 Kubernetes 控制器，它只执行一项任务。它能将sealed secrets（可提交至 Git）转换为plain secrets（可在应用程序中使用）。

```shell
helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
helm repo update
helm install sealed-secrets-controller sealed-secrets/sealed-secrets
```

安装后，控制器会自行创建两个密钥：

1. 私钥用于 secret decryption。This key应保留在群集内部，不得泄露给任何人。
2. 公钥用于secret encryption。它可以（也将会）在群集外使用，因此可以将它交给其他人。

如果你的应用程序可以使用 Kubernetes 的普通机密，那么它也可以使用sealed secrets。

## Encrypting your secrets

详细信息请[参考](https://github.com/bitnami-labs/sealed-secrets/blob/main/README.md#installation)

Kubeseal 与控制器的操作正好相反。它使用现有的 Kubernetes secrets 并对其进行加密。Kubeseal 向群集申请安装过程中创建的公钥，并用该密钥加密所有secrets 。

Kubeseal 需要访问集群才能加密机密。(它需要一个类似 kubectl 的 kubeconfig）。加密后的secrets 只能在加密过程中使用的集群中使用。

最后一点非常重要，因为它意味着所有secrets 都是特定于群集的。应用程序的命名空间也是默认使用的，因此秘密是群集和命名空间特定的。

因此，使用 kubeseal，只需获取 yaml 或 json 格式的任何现有秘密并对其加密即可：

```
kubeseal -n my-namespace <.db-creds.yml> db-creds.json
```

这将创建一个 SealedSecret，它是控制器专用的自定义 Kubernetes 资源。该文件可以安全地提交到 Git 或存储在其他外部系统中。
然后，你就可以在集群上应用该secret 了：

```shell
kubectl apply -f db-creds.json -n my-namespace
```

secret 现在是群集的一部分，当应用程序需要时，控制器将对其进行解密。

完整的过程如下 ：

1. 在本地创建一个普通的 Kubernetes 秘密。千万不要在任何地方提交。
2. 使用 kubeseal 在 SealedSecret 中加密密文。
3. 从工作空间删除原始密文，并将密封密文应用到群集。
4. 你可以选择将密封的秘密提交到 Git。
5. 部署应用程序，该应用程序需要正常的 Kubernetes 秘密才能运行。(应用程序不需要任何修改）。千万不要在任何地方提交。
6. 控制器会解密密封秘钥，并将其作为普通秘钥传递给你的应用程序。
7. 应用程序照常运行。

通过使用 Sealed Secrets 控制器，我们最终可以在 Git 中（以加密形式）存储所有机密，并与应用程序配置一起存储。

## Declarative 配置

```
argocd app create demo \
--project default \
--repo https://github.com/ForcemCS/gitops-certification-examples.git \
--path "./helm-app/" \
--sync-policy auto \
--dest-namespace default \
--dest-server https://kubernetes.default.svc
```

```
argocd app create demo \
--project default \
--repo https://github.com/ForcemCS/gitops-certification-examples.git \
--path ./kustomize-app/overlays/staging \
--sync-policy auto \
--dest-namespace default \
--dest-server https://kubernetes.default.svc
```

