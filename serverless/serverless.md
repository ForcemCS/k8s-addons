## serverless简介

**无服务器架构**已成为云计算领域最具争议的术语之一，关于软件是否属于“无服务器”范畴，存在两种截然不同的观点。此外，关于函数即服务 (FaaS) 和无服务器架构之间的区别也存在一些讨论。

首先，无服务器是一个抽象概念，不应被字面理解——尽管许多人确实如此。正如云计算并非指涉天空，无服务器也并非意味着代码无需服务器即可运行。它仅指用户或客户的体验，代表着用户需要与硬件及基础设施密切协作的连续体。

我们发现，随着对基础设施的关注度降低，工作负载也在分解，规模也在缩小。因此，如果说无服务器是一种方法和架构模式，那么函数和函数即服务 (FaaS) 就是它的一个子集，为应用这项技术和理念提供了一种具体的方式。

**函数的特征**

与单体应用相比，函数式应用往往具有以下特点：

- 让开发人员专注于代码编写，而不是基础设施和部署工件。
- 创建更小的工件和更少的代码行，因为它们承担的职责更少（甚至可能只有一个职责）。
- 它们可以是事件驱动的（由数据库、对象存储、发布/订阅触发），也可以部署为 REST 端点并通过 HTTP 访问。
- 它们易于管理，因为它们不依赖于底层存储。
- 它们是同构的——这意味着每个函数都必须符合给定的运行时接口。

### Serverless Platforms 

无服务器平台负责自动扩展，这种扩展通常有两种形式：

1. 在函数空闲时，缩减函数规模可以节省成本并降低系统负载。并非所有平台都支持缩减到零，而且从零副本缩减通常会导致延迟，称为“ *冷启动* ”，因为代码需要重新部署或重新初始化。
2. 随着对特定终端的需求增加，按比例扩展功能。

了解了 Serverless 和 FaaS 的定义之后，让我们从社区的角度来探讨两种观点。拥有并运营无服务器 SaaS 产品的云供应商会让你相信，一个平台只有在满足以下条件时才是无服务器的：

- 开发者只需编写称为函数的小型代码块。
- 客户按使用量（毫秒，无空闲时间）计费。
- 开发者和运营者都不应该对运行代码的服务器有任何访问权限或了解。
- 该代码可以扩展到大规模并行执行。
- 几乎不需要在云平台之间迁移代码。

其他无服务器架构的实践者，例如云原生计算基金会 (CNCF) 无服务器工作组的成员，认为虽然这种观点很流行，但它并不是一个二元对立的概念，无服务器架构更多的是一种哲学方法，而不是一套字面意义上的约束：

- 开发者只需编写称为函数或微服务的小型代码块。
- Serverless 是一种开发者体验，而不是一系列字面意义上的限制。
- 开发人员对服务器的担忧减少了，甚至完全消除了，而运营商可能仍然需要整合资源。
- 代码可以利用现有资源进行大规模并行扩展，也可以通过及时在云端配置新资源来实现。
- 按秒计费对开发者来说不是问题。
- 如有需要，代码可以在不同的系统或云端之间迁移。

在某种程度上，讨论的焦点从 *“Serverless 是否是一套硬性规定”* 转移到了 *“Serverless 系统的外观和使用体验是什么样的？运维人员和开发人员的体验如何？”* 这两个角度的共同点是，开发人员应该减少对基础设施的关注，编写易于部署的小块代码。

在 CNCF 内部，有许多项目基于 Kubernetes 构建，其中一些项目提供类似 Serverless 的体验。客户使用这些平台是因为它们具有可移植性，无需更改应用程序即可在客户的数据中心或选定的云平台上运行。

## CNCF 中的无服务器现状

### Kubernetes 中 Serverless 的现状

在Kubernetes的范畴内讨论无服务器架构时，CNCF无服务器生态图谱将相关服务划分为两类：一类可“安装”并部署在集群上运行，另一类则作为SaaS产品托管提供。

Kubernetes 中工作负载的最小基本单元是 Pod，它可以由一个或多个容器组成——例如，一个容器包含主工作负载，另一个容器则包含辅助容器，例如代理或日志收集器。Pod 从存储在容器镜像仓库中的容器镜像加载代码和用户空间。

要访问集群中的 Pod，需要用到各种 Kubernetes 网络原语，例如 Service、LoadBalancer 和 Ingress。它们各自提供了一种略有不同的方式将流量发送到正在运行的工作负载。

我们在 Kubernetes 上实现 Serverless 的基本要素包括：

- A container image with function code or an executable inside.
- A registry to host the container image.
- A Pod to run the container image.
- A Service to access the Pod.

通常，项目会在此堆栈之上添加更多组件，例如 UI、API 网关、Ingress 自动化、自动扩展、API 等等。

基于上述基础组件构建的项目具有相当高的兼容性；例如，为OpenFaaS构建的工作负载可在Knative上运行，反之亦然。

### Serverless 1.0 与 Serverless 2.0

在 Serverless 1.0 时代，云厂商开发独立产品时并未考虑可移植性或产品间迁移。通常情况下，从 AWS Lambda 迁移到 Azure Functions 需要重新设计签名、修改 zip 文件的构造方式、调整托管服务的访问方式，还要应对不同的可用区、区域以及内存和函数镜像大小等硬性限制。

对于某些企业客户而言，256MB 的功能映像、5 分钟的最大调用时长和 2GB 的最大内存容量限制过大，无法满足他们的需求。此外，Serverless 1.0 也无法支持微服务架构，而许多公司都在使用微服务。

<img src="C:\Users\ForceCS\Desktop\serverless\img\1.png" alt="1" style="zoom:50%;" />

Serverless 2.0 工作负载比 1.0 版本更具可移植性。无服务器2.0的定义如下：

+ 存储于OCI兼容的容器镜像中。
+ 在8080端口暴露HTTP服务器。
+ 可通过环境变量进行配置。

正是这个相对简短的定义，使得用 Node.js、Go、Python 或任何其他二进制语言编写的代码能够轻松地在 FaaS 平台之间迁移。Serverless 2.0 则*赋予了 OpenFaaS 函数可移植性* 。此外，它还带来了运行微服务（例如 Ruby on Rails、Express.js、Vert.x 和 Micronaut）的额外优势。

### OpenFaaS

OpenFaaS 的创建是为了让开发者能够利用自己的硬件，通过 Docker 容器来运行函数。它在 GitHub 上拥有超过 2.2 万颗星，并拥有一个由独立软件开发者和最终用户（如英国电信 BT、思科 Citrix、LivePerson 和 Vision Banco）组成的繁荣社区

+ Build templates 

  OpenFaaS 提供了一套用于生成函数的构建模板。这些模板也可以通过模板商店在线查找和共享。

+ Serving runtime 

  OpenFaaS 是最成熟的服务器运行时环境，也是该领域星标最多的项目。它支持使用 Prometheus 或 Kubernetes HPAv2 进行自动扩缩容，并可实现资源缩减至零后再恢复，从而节省资源。该服务器运行时环境提供 REST API、CLI 和 UI。OpenFaaS 支持使用 Istio 或 Linkerd2 进行流量转移。OpenFaaS 符合 Serverless 2.0 定义

+ Events

  OpenFaaS 支持多种事件触发器，包括 Apache Kafka、NATS、AWS SQS、Cron、MQTT 等。此外，我们还提供了一个 Golang **连接器 SDK** ，方便用户快速编写新的连接器。

+ Scale from Zero 

  支持从零开始和向零开始的缩放。

+ Managed

  OpenFaaS Cloud 是一个基于 OpenFaaS 直接构建的托管服务，提供多用户支持、丰富的仪表盘、指标以及与 GitHub/GitLab 的集成。OpenFaaS Cloud 是开源的，可以托管在本地。

### Knative

[Knative](https://github.com/knative/serving) 最初由 Google 开发，现在也得到了来自 IBM、Red Hat 和 VMware 等公司的贡献。Knative 在发布之初的功能涵盖了事件、构建和服务。自发布以来，其构建项目已被弃用，取而代之的是独立的 Tekton 项目，用于创建构建和持续集成 (CI) 流水线

+ Build templates

  Knative 没有自己的模板集，但可以使用 OpenFaaS 或 Buildpacks。大多数用户维护自己的 Dockerfile 和样板代码。Knative 符合 Serverless 2.0 的定义。

+ Serving runtime

  Knative 的服务运行时称为“Knative Serving”，需要单独的 API 网关（例如 Ambassador、Istio 或 Gloo）才能运行。Knative 通常与 Istio 一起安装，以实现流量转移。每次服务部署都是不可变的，内部会创建一个版本号，使用户能够回滚到之前的代码版本。

+ Events 

  Knative 事件处理项目涵盖了所有与事件的集成。它包含 4-5 个基本组件，用于从事件代理接收事件。配置完成后，即可从各种不同的事件源触发函数。

+ Scale from Zero 

  Knative 支持从零开始或从零开始扩展。冷启动时间明显较长；不过，该项目正在积极寻找降低延迟的方法。