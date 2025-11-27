# API Overview

在 K8s Gateway API 的设计模式中，有三个核心概念：`GatewayClass`、`Gateway` 和 `Route`。

- **GatewayClass (模板/菜单):** 由集群管理员定义。比如定义一个叫 "AWS-LB" 的类，或者 "Nginx-Proxy" 的类。这代表了底层基础设施的能力。

- **Gateway (订单):** 它是你发出的“订单”。你告诉 K8s：“基于 'AWS-LB' 这个模板，给我创建一个负载均衡器，监听 80 端口。”

  Gateway 只是负责“把门打开”（监听端口）。但进门之后，流量该往左走还是往右走？这是由 **Route (路由)** 决定的。

  + **Gateway:** “我在监听 80 端口。”
  + **HTTPRoute:** “如果是访问 `yoursite.com/api` 的流量，请转发给 `backend-service`。”

- **Controller (厨师):** 控制器看到你的 `Gateway` 订单后，就会去实际配置底层的网络设备（比如去 AWS控制台创建一个真实的 Load Balancer）。

## Route Resources

Gateway 只负责开门，具体的转发规则（去哪个 Service）要根据协议类型，由不同的 Route 资源来定义。

### 1. HTTPRoute (最常用，最智能)

> *状态：Standard (标准版/稳定版)*

这是平时开发最常用的资源。它是工作在 **七层 (Layer 7)** 的路由。

- **它的能力：** 它可以像把手术刀一样，拆开 HTTP 请求的包裹，看里面的细节。
- **它的判断依据：** 它可以根据 **URL 路径**（如 `/api` vs `/login`）、**Header 头**（如 `User-Agent: iPhone`）、或者 **Host 域名** 来转发流量。
- 场景举例：
  - 用户访问 `example.com/shop` -> 转发给 `shop-service`。
  - 用户访问 `example.com/admin` -> 转发给 `admin-service`。
  - 想在转发前给请求强行加一个 Header？它可以做。
- **"Terminated HTTPS":** 意思是 HTTPS 加密在 Gateway 处解密（卸载证书），变成了明文 HTTP，然后由 HTTPRoute 智能分析内容并转发。

```yaml
kind: Gateway
metadata:
  name: my-gateway
spec:
  gatewayClassName: my-gateway-class
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate  # <--- 重点：这里叫 "终结"，意思是在这解密
      certificateRefs:
      - name: my-tls-secret # <--- 证书引用放在这里！
```

### 2. TLSRoute (加密透传，半智能)

> *状态：Experimental (实验性)*

这是工作在 **TLS 层** 的路由。它比 TCP 高级一点，但比 HTTP 低级一点。

- **它的能力：** 它**不看** HTTP 的具体内容（因为被加密了），它只看 TLS 握手时的“外包装”。

- **它的判断依据：** 主要靠 **SNI (Server Name Indication)**。简单说，就是握手时客户端会喊一声：“我要找 `a.com`！” TLSRoute 听到后，就把流量转给负责 `a.com` 的后端。

- 两种模式：

  1. **Passthrough (透传):** Gateway **不解密**。它就像传声筒，直接把加密的数据流扔给后端 Pod，后端 Pod 自己负责解密。这对安全性要求极高的场景很有用（端到端加密）。

     想象一种情况：你的应用是银行核心系统，安全性要求极高。**你不信任集群管理员，你也不信任网关（Nginx/Envoy）。** 你要求数据从离开用户浏览器开始，一直到进入你的 Pod 内部，全程都是加密的。网关不能解密，不能看里面的数据。

     + **用什么 Route？** 必须用 `TLSRoute`。
     + **证书配在哪？** **配在你的 Pod 应用程序里！** (比如你的 Spring Boot 或 Go 程序启动时加载证书)。Gateway 根本不需要（也没资格）拿证书。
     + **流程：** 用户 -> (HTTPS) -> **Gateway (看不懂内容，原样转发)** -> (HTTPS) -> **Pod (自己解密)**。

  2. **Terminate (终止):** 在 Gateway 解密，然后转发（这就有点像 HTTPRoute 了，但通常 TLSRoute 用于非 HTTP 的 TLS 协议，比如数据库连接的 TLS）。

     这是**TLS 层**结合**四层（Layer 4）**处理。

     - **动作：** Gateway 解密 TLS 流量。
     - 之后发生的事：Gateway 完全不看里面的内容是什么协议。
       - 它不管里面是 HTTP，还是 MongoDB 的命令，还是 Redis 的指令。
       - 它只负责把“加密外壳”剥掉，露出里面的“裸数据流（TCP Stream）”。
     - **路由能力：** 因为它不解析内容，所以它**不能**根据 URL 路径（如 `/api`）转发。它**只能**根据握手时的 SNI（域名，如 `db.example.com`）来决定把这股数据流转发给谁。
     - **后端收到的是：** 一股解密后的 TCP 字节流。

### 3. TCPRoute / UDPRoute (最底层，不管内容)

> *状态：Experimental (实验性)*

这是工作在 **四层 (Layer 4)** 的路由。最简单粗暴。

- **它的能力：** 它根本不管你发的是图片、网页还是数据库查询指令，它只认 **端口 (Port)**。

- 它的限制：

   文档特别提到 

  "there is no discriminator"

  （没有鉴别器）。

  - 什么意思呢？在 HTTP 里，80 端口可以同时给 `a.com` 和 `b.com` 用（靠域名区分）。
  - 但在纯 TCP 模式下，流量来了就是来了，没有域名、没有路径。所以通常 **一个端口只能对应一个后端服务**。比如 Gateway 开了 3306 端口，就只能全转给 MySQL，不能分流。

- **场景举例：** 你的服务不是 HTTP 协议的（比如 Redis、MySQL、游戏服务器），或者你就是想单纯地把某个端口的流量全部导向某个 Pod。

## Attaching Routes to Gateways

### 1. 什么是 "Attaching" (绑定/连接)？

**解释：**
你可以把 `Gateway` 想象成一台**“裸机”**（刚买回来的路由器），它只有硬件能力（监听 IP 和端口），但脑子是空的。
当我们将一个 `Route` \**“绑定”\** 到 Gateway 上时，就相当于给这台路由器**插上了一张“功能卡”**。

- Gateway 依然由运维团队管理。
- Route 中的规则（比如 `/api` 转发给 A 服务）会被注入到 Gateway 底层的负载均衡器中生效。

**关键点：** 这不是随随便便就能连的。 *"built-in controls"*（内置控制），这是一种**“双向握手”**机制：

1. **Route 说：** “我想挂载到名为 `main-gateway` 的网关上。”（通过 `parentRefs` 字段）
2. **Gateway 说：** “我允许来自 `project-a` 命名空间的 `HTTPRoute` 挂载进来。”（通过 `listeners` 里的 `allowedRoutes` 字段）
   **只有双方都同意，连接才会建立。** 这就防止了别的团队乱碰你的网关。

### 2. 三种连接模式 (核心内容)

文档列出了三种模式，分别对应三种不同的**组织架构**和**业务场景**：

#### 模式一：One-to-one (一对一)

> *场景：个人开发者、小型应用、独占网关*

- **结构：** 1 个 Gateway <--> 1 个 Route。
- **解释：** 整个链路归**一个人**（或一个团队）管。
- **例子：** 你自己搭了个博客。你创建了一个网关专门给博客用，又写了一个路由规则。这种模式最简单，没有复杂的权限隔离，适合“全栈”场景。

#### 模式二：One-to-many (一对多 / 网关共享) —— **最重要！**

> *场景：大公司、微服务架构、多租户*

- **结构：** 1 个 Gateway <--> N 个 Route (来自不同 Namespace)。

- 解释：

   这是 Gateway API 设计的精髓。

  - **运维团队（Platform Team）：** 负责维护**这一个**昂贵的 Gateway（配置 SSL、防火墙、高可用 IP）。它是所有流量的统一入口。
  - **业务团队 A（支付组）：** 在自己的 Namespace 里写一个 Route，处理 `/pay`。
  - **业务团队 B（订单组）：** 在自己的 Namespace 里写一个 Route，处理 `/order`。

- **好处：** 业务团队不需要关心底层的负载均衡器怎么配，他们只需要把自己的 Route **“插”** 到运维提供的公共 Gateway 上即可。大家共用一个入口，但互不干扰。

#### 模式三：Many-to-one (多对一 / 多入口复用)

> *场景：混合云、内网+外网同时访问*

- **结构：** N 个 Gateway <--> 1 个 Route。
- **解释：** 你写好了一个 Route 规则（比如：访问 `myapp.com` 转发给 `my-service`）。你想让这个规则**同时**在多个不同的网络环境下生效。
- 例子：
  - **Gateway A (公网 LB):** 给互联网用户访问。
  - **Gateway B (内网 LB):** 给公司内部员工访问（VPN）。
  - 你只需要写**一个** `HTTPRoute`，然后在配置里写上：“我要绑定到 Gateway A **和** Gateway B”。
- **好处：** 一次配置，多处生效。你不需要为了内网和外网分别写两套一模一样的转发规则。

### Example

#### 例子 1：精确点名（最严格的握手）

这个场景是：

- **Route (业务方):** 在 `gateway-api-example-ns2` 命名空间。
- **Gateway (基建方):** 在 `gateway-api-example-ns1` 命名空间。
- **目的:** 业务方想跨界使用基建方的网关。

##### 1. 业务方的申请书 (`HTTPRoute` YAML)

```yaml
spec:
  parentRefs:
  - kind: Gateway
    name: foo-gateway
    namespace: gateway-api-example-ns1  # <--- 重点在这里！
```

- **解释：** 这里的 `parentRefs` 就是“申请书”。它明确写着：“我要挂载到 `ns1` 命名空间里的 `foo-gateway` 上”。
- **潜台词：** “隔壁老王家的网关，借我用用。”

##### 2. 基建方的审批章 (`Gateway` YAML)

仅仅有申请是不够的，Gateway 必须明确同意，否则谁都能来蹭网关，岂不乱套了？

```yaml
listeners:
  - name: prod-web
    allowedRoutes:  # <--- 审批规则
      kinds:
      - kind: HTTPRoute
      namespaces:
        from: Selector
        selector:
          matchLabels:
            # 这是一个 K8s 自动生成的标签，值就是命名空间的名字
            kubernetes.io/metadata.name: gateway-api-example-ns2
```

- **解释：** `allowedRoutes` 定义了谁可以连进来。
- **关键点 `matchLabels`：** 这里用了一个特殊的技巧。K8s 会自动给每个 Namespace 打上一个标签，key 是 `kubernetes.io/metadata.name`，value 是命名空间的名字。
- **翻译过来就是：** “我只接受来自 `gateway-api-example-ns2` 这个命名空间的路由连接，其他的免谈。”

**总结：** 这是一个**一对一的、指名道姓的**授权。A 申请连 B，B 说“我只允许 A 连进来”。

------

#### 例子 2：凭“证”入内（更灵活的握手）

如果有 100 个业务团队（100 个命名空间）都要用这个网关，按上面的方法，Gateway 管理员得把这 100 个名字都写进去，累死了。

于是有了第二个例子：**基于标签（Label）的授权**。

##### Gateway YAML 配置

```yaml
listeners:
  - name: prod-web
    allowedRoutes:
      namespaces:
        from: Selector
        selector:
          matchLabels:
            expose-apps: "true"  # <--- 重点在这里！
```

- **解释：** 这次 Gateway 不再点名了，而是制定了一个规则：“不管你是哪个命名空间的，只要你的命名空间上贴了 `expose-apps: "true"` 这个标签，我就允许你连进来。”

#### 请求流程 (Request flow)

让我们把这个过程想象成**送信**：

1. **Client Request (寄信):**

   - 用户在浏览器输入 `http://foo.example.com`。这封信要寄出去了。

2. **DNS Resolution (查地址):**

   - 浏览器问 DNS 服务器：“`foo.example.com` 住哪？”
   - DNS 回复的 IP 地址，正是 K8s 中 **`Gateway` 资源** 所拥有的 IP（通常是 Load Balancer 的 VIP）。

3. **Listener & Host Matching (大门保安):**

   - 流量到达 Gateway 的 IP。
   - Gateway 的 `Listener`（监听器）开始工作。它看了一眼信封上的 **Host Header**（`foo.example.com`）。
   - 它会去查找：**“有没有哪个 `HTTPRoute` 是认领 `foo.example.com` 这个域名的？”**
   - 找到了！把信转交给对应的 `HTTPRoute` 处理。

4. **Path/Header Matching (分拣员 - 初筛):**

   - 现在轮到 `HTTPRoute` 工作了。
   - 文档提到的 *"match rules"*。路由规则开始看信的内容：
   - “URL 路径是 `/api/v1` 吗？” “Header 里有 `User-ID` 吗？”
   - 如果匹配上了，就进入下一步。

5. **Filters (分拣员 - 加工 - 可选):**

   - 文档提到的 *"modify the request"*。

   - 在转发之前，

     ```
     HTTPRoute
     ```

      还可以对请求做点手脚。比如：

     - “添加一个 Header: `X-Proxy-Time: Now`”
     - “把路径里的 `/api` 去掉”
     - “做一个重定向”

6. **Forwarding (送达):**

   - 最后一步。`HTTPRoute` 根据 `backendRefs`（后端引用），决定把这封信送给哪个 K8s **Service**。
   - Service 再通过 Endpoints 找到具体的 **Pod**。
   - **任务完成。**

### TLS Configuration

我们在之前讨论过，TLS 证书通常配置在 `Gateway` 上（监听器里）。这里补充了一个非常重要的企业级特性：**跨命名空间引用证书**。

- **场景：** 你们公司有一个“安全团队”，他们把所有的 SSL 证书（Secret）都锁在一个叫 `security-secrets` 的命名空间里，普通开发人员根本进不去。
- **问题：** 你的 Gateway 跑在 `prod-app` 命名空间，怎么读取那个被锁住的证书？
- **解决：** Gateway API 允许你在配置 `tls` 时，通过 `ReferenceGrant`（引用授权）机制，去引用另一个命名空间里的 Secret。
- **价值：** 证书集中管理，更安全。

## Attaching Routes to Services

**这是 Gateway API 的一个重大转折点！** 也就是业内常说的 **GAMMA** 计划。

- **以前（标准模式 / 南北向流量）：**
  - 流量从**集群外**进来。
  - `HTTPRoute` 绑定到 `Gateway`（大门）。
  - **路径：** 外部 -> Gateway -> Route -> Service。
- **现在（Service Mesh 模式 / 东西向流量）：**
  - 流量是**集群内部**服务之间的调用（比如 A 服务调 B 服务）。
  - 既然是内部通信，就不经过那个“大门”（Gateway）了。
  - **那路由规则挂哪？** 文档说：**直接挂在 Service 上！**
  - **路径：** Service A -> (Route 规则生效) -> Service B。

**通俗解释：**

- **Gateway 模式：** 像是**公司的前台**。外卖小哥（外部流量）来了，前台根据规则指引他去哪。
- **Service Mesh 模式：** 像是**办公室内部的座机**。你（Service A）拨打同事（Service B）的分机号时，电话系统直接根据规则（Route）进行转接（比如转接到他的手机，或者录音）。这个规则直接绑定在电话系统上，不需要经过前台。

## Extension points (扩展点 / 插件系统)

K8s 官方制定标准时，不可能把世界上所有的功能都塞进去（那样会变得极其臃肿）。所以，他们留了很多**“接口” (Interface)**，允许厂商（AWS, Google, Istio, Kong, Nginx）去填补自定义的功能。

文档列出了三个主要的扩展插槽：

### A. BackendRefs (不仅是 Service)

- **标准情况：** `HTTPRoute` 把流量转发给 K8s `Service`。
- **扩展能力：** 它可以转发给**任何东西**！
- 举例：
  - 转发给一个 **AWS Lambda** 函数。
  - 转发给一个 **S3 Bucket**（直接由网关读取静态文件）。
  - 转发给一个 **集群外部的 IP 地址**（传统数据库）。
  - *只要厂商的 Controller 支持，后端可以是万物。*

### B. HTTPRouteFilter (自定义过滤器)

- **标准情况：** 支持修改 Header、重定向等基础操作。
- **扩展能力：** 允许插入更复杂的逻辑钩子。
- 举例：
  - **鉴权 (Auth):** 在转发前，先去调一下 OAuth 服务，看看用户有没有登录。
  - **限流 (RateLimit):** “这个 IP 每秒只能访问 10 次”。
  - **WAF:** 检查请求里有没有 SQL 注入攻击代码。

### C. Custom Routes (自定义路由协议)

- **标准情况：** 官方只提供了 `HTTPRoute`, `TLSRoute`, `TCP/UDPRoute`, `GRPCRoute`。
- **扩展能力：** 如果你想支持一种很小众的协议怎么办？你自己造一个 Route！
- 举例：
  - **KafkaRoute:** 专门处理 Kafka 消息的路由。
  - **MqttRoute:** 物联网设备专用的路由。
  - **RtmpRoute:** 视频直播流的路由。
  - *前提是你的 Gateway 实现者（比如 Kong 或 Istio）写了代码来支持这种新资源。*