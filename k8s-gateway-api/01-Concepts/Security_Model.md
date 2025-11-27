# Security Model

## Crossing Namespace Boundaries

默认情况下，为了安全，A 房间的资源是不应该随意引用 B 房间的资源的。Gateway API 打破了这个限制，允许跨房间操作，但为了防止乱来，它设计了一套非常严格的 **“握手（Handshake）”机制**。

### 1. 更加精细的“准入控制” (Route Binding)

在上一个回答中我们提到了 `allowedRoutes`，这里文档进一步深入探讨了 **“怎么选才安全”**。

#### 这里的例子 (`matchExpressions`)

```yaml
selector:
  matchExpressions:
  - key: kubernetes.io/metadata.name
    operator: In
    values:
    - foo
    - bar
```

**解释：**
Gateway 管理员制定了一条规则：“我只允许来自 `foo` 房间和 `bar` 房间的路由连上来。”
这是通过 K8s 自动生成的标签 `kubernetes.io/metadata.name` 来实现的，这个标签的值永远等于 Namespace 的名字。

#### ⚠️ 重点警告：自定义标签的风险 (Risks of Other Labels)

文档特意写了一段 **“Risks of Other Labels”**，这是一个非常重要的安全提示。

- **安全做法：** 使用 `kubernetes.io/metadata.name`。这是 K8s 系统打的标签，普通用户改不了。这就像查身份证，身份证是公安局发的，造不了假。

- 危险做法：

   使用自定义标签，比如 `env: prod`

  - **场景：** 网关管理员设置规则：“只要 Namespace 有 `env: prod` 标签，就允许连接”。
  - **漏洞：** 如果你是集群里的一个普通用户，虽然你没有网关权限，但你有权给自己创建的 Namespace 打标签。你偷偷给自己打个 `env: prod`，你的测试路由就混进了生产网关！
  - **比喻：** 这就像查“红帽子”。如果规则是“戴红帽子的人能进”，那坏人自己买顶红帽子戴上就混进去了。

------

### 2. 通用签证：ReferenceGrant (引用授权)

这是 Gateway API 引入的一个**全新资源对象**，用来解决除了 Gateway<->Route 之外的跨界引用问题。

**场景：**

- **Route (在 `prod` 命名空间):** 想把流量转发给 **Service (在 `backend` 命名空间)**。
- **问题：** 默认情况下，Route 只能转发给同屋的 Service。跨屋转发是被禁止的。
- **解决：** `backend` 命名空间必须发一张“签证” (`ReferenceGrant`)。

#### 配置解读

```yaml
kind: ReferenceGrant
metadata:
  name: allow-prod-traffic
  namespace: backend  # 这个资源必须建在“被引用”的那一方（目标方）
spec:
  from:  # 谁想来？
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    namespace: prod
  to:    # 允许访问什么？
  - group: ""
    kind: Service
```

**翻译过来就是：**
我（`backend` 命名空间）特此授权：来自 `prod` 命名空间的 `HTTPRoute` 资源，可以引用（访问）我的 `Service` 资源。

### 3. 尚未解决的难题：限制 GatewayClass 的使用范围

文档最后一段提到了一个高级概念。

- **问题：** 假设你定义了一个 `GatewayClass` 叫 "AWS-Super-Expensive-LB"（超贵的负载均衡器）。目前 Gateway API 的标准里，**没有内置的方法**去限制说：“这个贵的 Class 只能给 VIP 团队用，实习生团队不能用”。
- **现状：** 只要实习生有创建 `Gateway` 的 RBAC 权限，他就可以指定用这个贵的 Class。
- **建议方案：** 官方建议使用第三方的 **Policy Agent（策略引擎）**，比如 **OPA (Open Policy Agent)** 或 **Gatekeeper**。你需要写一个策略脚本来强制执行：“如果是 `dev` 命名空间，禁止使用 `AWS-Super-Expensive-LB`”。