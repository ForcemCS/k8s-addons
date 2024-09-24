## 使用说明
```shell
helm install consul hashicorp/consul  --version 1.5.0 --namespace consul --values values.yaml
```


## ACL
Consul 是一座图书馆，里面存放着各种书籍（资源）。ACL 就像图书馆的门禁系统，Token 就像你的借书证。

+ 没有借书证（Token），你就无法进入图书馆（访问资源）。
+ 不同的借书证（Token）拥有不同的权限，例如只能借阅特定类型的书籍。
+ 图书管理员的万能钥匙（Bootstrap Token）可以打开图书馆的所有房间和书柜。

## 配置 CLI 以与 Consul 数据中心交互

```
export CONSUL_HTTP_TOKEN=9fd4899e-744b-xxxxxxxxxxxx
export CONSUL_HTTP_ADDR=https://12.0.0.21:31960
export CONSUL_HTTP_SSL_VERIFY=false
```

更多内容请[参考](https://developer.hashicorp.com/consul/tutorials/get-started-kubernetes/kubernetes-gs-deploy)

## Consul API 网关与 Consul 的关系详解

Consul API 网关是 HashiCorp Consul 的一个组件，它充当 Consul 服务网格中所有外部请求的单一入口点。简单来说，**Consul API 网关是建立在 Consul 之上的一个反向代理和 API 管理层**，利用 Consul 的服务发现、健康检查和配置管理功能，简化了微服务架构中 API 的访问和管理。

**Consul 与 Consul API 网关的关系可以概括为：**

- **Consul 是基础：** Consul 提供了服务发现、配置管理、健康检查等核心功能，为 Consul API 网关提供了必要的基础设施。
- **API 网关是扩展：** Consul API 网关构建在 Consul 之上，利用其功能来管理和路由 API 流量，并提供额外的安全性和可观察性。

**具体来说，Consul API 网关利用 Consul 的以下功能：**

- **服务发现：** Consul API 网关利用 Consul 的服务注册和发现功能，动态地将请求路由到后端服务，无需手动配置。
- **健康检查：** Consul API 网关使用 Consul 的健康检查机制，自动识别和隔离不健康的实例，确保只有健康的实例接收流量。
- **配置管理：** Consul API 网关可以从 Consul 中读取配置信息，例如路由规则、访问控制策略等，实现动态配置更新，无需重启网关。

**除了利用 Consul 的核心功能外，Consul API 网关还提供以下功能：**

- **反向代理：** 接收外部请求，并将其转发到相应的微服务。
- **路由：** 根据路径、头部、方法等条件，将请求路由到不同的后端服务。
- **负载均衡：** 在多个服务实例之间分发流量，提高服务可用性和性能。
- **安全认证和授权：** 集成多种认证和授权机制，例如 JWT、OAuth2、mTLS 等，保护 API 免受未经授权的访问。
- **流量控制：** 提供速率限制、熔断等机制，防止服务过载和级联故障。
- **可观察性：** 提供指标收集、日志记录、追踪等功能，帮助用户监控和排查 API 问题。

**总而言之，Consul API 网关与 Consul 紧密集成，利用其强大的功能简化了微服务架构中 API 的管理和访问。** 它为开发者提供了一个统一的入口点来管理 API 流量，并提供了丰富的功能来保障 API 的安全性、可靠性和可观测性。

## 配置参数说明

在 Consul server agent 配置文件中，`client_addr` 和 `cluster_addr` 是两个重要的网络地址配置，用于控制 Consul server 如何与客户端和集群中其他 server 进行通信。

**1. `client_addr`:**

- **作用:** 指定 Consul server 监听客户端请求的地址和端口。
- **对象:** 主要用于 Consul 客户端（例如，使用 Consul API 或 DNS 接口的服务）连接到 server 并与其进行交互。
- **示例:** `client_addr = "0.0.0.0:8500"` (监听所有接口的 8500 端口)
- 注意:
  - 如果你的 Consul 客户端和 server 运行在同一台机器上，可以直接使用 `127.0.0.1`。
  - 如果你的 Consul server 运行在私有网络中，并且你希望外部客户端可以通过负载均衡器访问它，则应将 `client_addr` 设置为负载均衡器的地址。

**2. `cluster_addr`:**

- **作用:** 指定 Consul server 用于与集群中其他 server 进行通信的地址和端口。
- **对象:** 用于 server 之间的内部通信，例如领导者选举、数据同步和状态传播等。
- **示例:** `cluster_addr = "10.0.0.5:8300"` (使用私有网络地址 10.0.0.5 的 8300 端口)
- 注意:
  - `cluster_addr` 必须是集群内所有 server 都可以访问的地址。
  - 通常建议使用私有网络地址来提高安全性并减少网络延迟。

**总结:**

- `client_addr` 用于客户端与 Consul server 的通信，而 `cluster_addr` 用于 Consul server 之间的内部通信。
- 为了确保 Consul 集群的正常运行，请根据您的网络环境和安全需求正确配置这两个地址。

希望这个解释能够帮助您理解 `client_addr` 和 `cluster_addr` 的区别和作用。如果您还有其他问题，请随时提出。

## 客户端代理

HashiCorp Consul 中**客户端代理（client agent）**的作用和部署建议。

**简单来说：**

- **Consul 客户端代理**就像“眼线”，部署在各个服务器节点上，负责监控节点和服务的健康状况，并汇报给 Consul 集群。
- 客户端代理通过 **RPC（远程过程调用）** 与 Consul 服务器通信，默认端口为 8300。
- Consul **没有客户端代理和服务数量限制**，但为了系统稳定性，建议将服务分布在多个数据中心，每个数据中心最多部署 5000 个客户端代理。
- 除了使用客户端代理，还可以使用 **Envoy 代理** 来搭建 Consul 服务网格，这种方式更简化，不需要部署客户端代理。

