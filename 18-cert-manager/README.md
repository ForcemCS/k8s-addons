## ACME

+ ACME Issuer 和 ACME 协议

  - **ACME (Automated Certificate Management Environment):** 这是一种协议，用于自动化证书颁发机构 (CA) 与用户（通常是网站或服务的管理员）之间的证书申请、验证和颁发过程。它由 Let's Encrypt 等机构推广，旨在简化 SSL/TLS 证书的管理，让 HTTPS 更加普及。
  - **ACME Issuer:** 在 cert-manager 中，`ACME Issuer` 是一种资源类型。它代表你在某个支持 ACME 协议的 CA 服务器上的一个账户。当你创建一个 `ACME Issuer` 时，cert-manager 会为你生成一个私钥，这个私钥用于你在该 CA 服务器上的身份识别。 你可以把它想象成你在 Let's Encrypt 上的一个账号。

  - **cert-manager**: 是一个 Kubernetes 的插件，用来自动化管理和颁发来自各种来源的 TLS 证书，这里它支持ACME协议，可以和 Let's Encrypt 等 CA 集成。

+ Challenges

  + **HTTP01**

    + 你向 ACME 服务器（如 Let's Encrypt）请求证书，ACME 服务器会提供一个**验证 token**。
    + 你需要将这个 token 放到域名对应的网站上，特定的路径下（通常是 `http://yourdomain/.well-known/acme-challenge/`）。
    + ACME 服务器会尝试访问这个 URL，如果能够正确获取 token，就证明你控制了这个域名，验证通过。
    + 证书签发成功。

    - cert-manager 的作用:

      当你使用 cert-manager 和 HTTP01 挑战时，cert-manager 会自动完成以下操作：

      1. 生成需要放置在特定 URL 的密钥。
      2. 创建一个临时的、小型 Web 服务器来提供这个密钥。
      3. 配置你的 Kubernetes Ingress（如果使用 Ingress），将对该特定 URL 的请求路由到这个临时 Web 服务器。
      4. 一旦验证完成，cert-manager 会清理这些临时资源。

  + **DNS-01**

    + ACME 服务器会要求你在你的域名的 DNS 记录中添加一条特定的 TXT 记录（包含一个计算出的密钥）。如果 ACME 服务器能够通过 DNS 查询找到这条 TXT 记录，并验证其中的密钥，就认为你拥有这个域名。
    + cert-manager 会（在你有适当权限的情况下）自动完成以下操作：
      1. 生成需要添加到 DNS TXT 记录的密钥。
      2. 使用你配置的 DNS 提供商（例如 Cloudflare、Route 53、Google Cloud DNS 等）的 API，自动在你的 DNS 记录中添加这条 TXT 记录。

### Configuring Issuers

#### HTTP-01

参考示例如下

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration
    email: xxxx@example.cn
    # 如果你丢失了这个 Secret（例如，删除了它）就相当于丢失了 ACME 账户的密钥。
    # 虽然你可以重新创建一个新的 Secret 和密钥，并使用新账户申请证书，但你将无法撤销（revoke）使用旧账户生成的任何证书
    privateKeySecretRef:
      name: letsencrypt-prod
    # 这意味着你需要一个 Ingress 控制器（如 Nginx Ingress Controller）来将流量路由到 cert-manager 创建的临时 Pod，
    # 该 Pod 将提供挑战所需的响应。
    solvers:
      - http01:
          ingress:
            ingressClassName: traefik

```

我们需要注意的是可以通过`selector` 字段配置适用于哪些 Certificate。请[参考](https://cert-manager.io/docs/configuration/acme/)

也可以使用**Private ACME Servers**

#### DNS-01

### 申请证书

