## 审计策略
审计日志根据配置的审计策略捕获。审计策略定义了应记录的事件和包含的数据。审计策略中配置的规则按顺序处理，第一条匹配规则设置事件的审计级别。
每个请求都可记录相关的阶段。以下是可以启用审计记录的阶段：
+ RequestReceived: 审计处理程序收到请求后立即生成事件的阶段
+ ResponseStarted: 一旦发送了响应头，但在发送响应体之前。
+ ResponseComplete: 响应体已完成
+ Panic: 发生恐慌时产生的事件
捕获的事件信息取决于配置的审核级别。您可以使用以下审核级别：
+ None: 不记录符合此规则的事件
+ Metadata: 记录请求事件元数据。不记录请求/响应正文。
+ Request: 日志事件元数据和请求正文
+ RequestResponse: 日志事件元数据以及请求和响应正文

## 审计日志后端实现
Kubernetes 为审计日志提供了两种后端实现：日志文件和 webhook。Kubernetes 审计 webhook 作为一个扩展点，能让你通过外部服务或 webhook  路由和处理审计事件

## 术语

+ `Artifact`: 人工制品是 `falcoctl` 可以操作的元素，目前只考虑 `rulesfiles` 和 `plugins` 。

+ `Index` ：一个.yaml 文件，包含所有可用的 **Artifacts（规则文件、插件）** 的列表，以及它们存储在哪些**注册表（Registry）和仓库（Repository）** 中。该工具的默认配置包含一个索引文件，指向 `falcosecurity` 组织官方支持的工件，请参见[此处](https://github.com/falcosecurity/falcoctl/blob/gh-pages/index.yaml)。用户也可以维护自己的索引文件，指向包含自定义规则文件和插件的注册表和资源库。

+ `Registry` ：注册表存储了 `falcoctl` 所理解的与 [OCI 标准](https://opencontainers.org/)有关的工件，任何符合标准的 OCI 都可以使用。官方注册中心使用 Github Packages

  存储 Falco **Artifacts（规则文件、插件）** 的地方，类似于存储容器镜像的 Docker Hub、Harbor、Quay.io。

+ `Repository` ：**Repository（仓库）** 是 **Registry（注册表）** 中的一个子目录，专门存放**某一类 Artifact**。
