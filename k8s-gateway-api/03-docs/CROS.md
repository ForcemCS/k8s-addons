## CROS

我们直接用一个**实际的电商开发场景**来拆解：

------

### 1. 场景设定：你的项目有两个域名

- **前端域名：** `https://www.my-mall.com` （用户在浏览器里访问的地址）
- **后端 API 域名：** `https://api.my-mall.com` （Envoy Gateway 守着的网关地址）

**冲突点：** 浏览器发现这两个域名的“前缀”（子域名）不一样。根据“同源策略”，浏览器会怀疑：`api.my-mall.com` 的数据是不是真的想给 `www.my-mall.com` 看？

------

### 2. 没有配置 CORS 时会发生什么？

1. 用户打开 `www.my-mall.com`，点击“查看订单”。
2. 网页里的 JavaScript 代码发起了一个请求：`GET https://api.my-mall.com/orders`。
3. **浏览器拦截了回复：** 其实请求发过去了，后端也回了数据。但是浏览器看了一眼发现：**“响应头里没说允许 my-mall.com 访问啊！”**
4. **报错：** 浏览器直接把数据丢掉，并在控制台给你报一个红色的 `CORS Error`。

------

### 3. 配置了 CORS（通过 Envoy Gateway）会发生什么？

这时候，你在 Envoy Gateway 的 `SecurityPolicy` 里配置了 CORS：

- **允许的源 (Allow Origin):** `https://www.my-mall.com`
- **允许的方法 (Allow Methods):** `GET, POST`

#### 实际对话流程：

1. **浏览器先“探路” (Preflight):** 浏览器不会直接发请求，它先发一个 **OPTIONS** 请求问 Envoy：“嘿，网关，我这儿有个 `www.my-mall.com` 的网页想调你的 `/orders` 接口，你准吗？”
2. **Envoy 回复：** “准了！我这本本上记着呢，`www.my-mall.com` 是自己人。”（返回 200 OK，并带上 `Access-Control-Allow-Origin: https://www.my-mall.com`）
3. **浏览器正式发起请求：** 浏览器看到回复后，才真正发送 `GET /orders`。
4. **成功：** 用户顺利看到了订单。

------

### 4. Envoy Gateway 提供的两种方式怎么选？

#### 方式 A：`SecurityPolicy`（全能管家）

- **例子：** 你在整栋大楼的 **总监控室** 里写了一条规定：“凡是挂着 `my-mall.com` 工牌的人，可以进这栋楼的所有办公室（Gateway 级别），或者某几间办公室（Route 级别）。”
- **特点：** 功能极强。比如你可以规定，跨域请求的有效期是 1 小时（避免浏览器频繁探路），或者允许带上某些特殊的自定义加密头。

#### 方式 B：`HTTPCORSFilter`（房间贴纸）

- **例子：** 你在 **具体的某一间办公室**（HTTPRoute）门口贴了个条子：“本房间允许隔壁的小王进入。”
- **特点：** 简单、标准。它是 Gateway API 原生支持的，如果你以后不打算用 Envoy 了，换成别的网关，这个配置可能照样能用。

------

### 5. 为什么要在网关做，而不是在后端代码做？

假设你公司有 3 个后端小组：

- **A 组 (Java):** 写订单接口。
- **B 组 (Go):** 写用户接口。
- **C 组 (Python):** 写库存接口。

**如果不通过 Envoy 网关统一配：** 每个组都要在自己的代码里写一遍 CORS 逻辑。一旦前端域名改了（比如换成 `.cn`），3 个组都要改代码、重新发布。

**如果在 Envoy Gateway 配：** 你只需要改一下 `SecurityPolicy` 的配置文件。后端程序员甚至都不知道 CORS 的存在，他们只需要专心写业务逻辑。

------

### 总结建议

- **初学者/简单场景：** 用 `HTTPCORSFilter` 挂在 `HTTPRoute` 上，简单省事。
- **复杂场景/公司级规范：** 用 `SecurityPolicy`，可以在一个地方管住整个网关的跨域规则，更专业也更安全。

## 示例
```
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: cors-example
spec:
  # 精准作用到这个服务
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  cors:
    #指定哪些网站可以跨域调用你的 API。
    allowOrigins:
        - "http://*.foo.com"
        allowMethods:
        - GET
        - POST
        allowHeaders:
        - "Authorization"    # 允许前端发送登录 Token
        - "Content-Type"     # 允许前端发送 JSON 数据
        exposeHeaders:
```
