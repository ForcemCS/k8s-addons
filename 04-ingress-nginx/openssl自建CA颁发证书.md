## openssl自建CA颁发证书

### 阶段一：创建自签名根 CA

根 CA 由一个私钥和一个自签名的公共证书组成。

**1. 生成 CA 的私钥**

这是CA 最核心、最敏感的文件。任何拥有它的人都可以冒充你的 CA 签发证书。**务必妥善保管！**

```bash
# 生成一个 4096 位的 RSA 私钥
openssl genpkey -algorithm RSA -out ca.key -pkeyopt rsa_keygen_bits:4096
```

- `ca.key`: 生成的 CA 私钥文件。

**2. 创建 CA 的自签名证书**

我们用上一步的私钥来为自己“签名”，生成一个公共证书。这个证书就是信任的根。

```bash
openssl req -x509 -new -nodes \
  -key ca.key \
  -sha256 \
  -days 3650 \
  -out ca.crt \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=My Org/CN=My Internal Root CA"
```

**命令解释：**

- `-x509`: 表示输出一个自签名的证书，而不是证书请求。

- `-new`: 表示这是一个新的请求。

- `-nodes`: "No DES" 的缩写，表示不为私钥设置密码。这在自动化脚本中很方便，但在生产环境中请考虑移除此项以增加安全性（移除后每次使用都需要输入密码）。

- `-key ca.key`: 用于签名的私钥。

- `-sha256`: 使用 SHA-256 哈希算法。

- `-days 3650`: 证书有效期（10年）。CA 证书有效期通常设置得较长。

- `-out ca.crt`: 输出的 CA 公共证书文件。

- ```
  -subj "..." : 在命令行中直接提供证书的主题信息，避免交互式输入。
  ```

  - `CN` (Common Name): 证书的通用名称，这里我们给它一个描述性的名字。

`ca.key`（CA私钥）和 `ca.crt`（CA公共证书）。**`ca.crt` 就是你要提供给 Ingress-NGINX 用于验证客户端证书的文件。**

------

### 阶段二：为客户端生成证书请求 (CSR)

现在，我们模拟一个客户端（比如一个用户或一个服务）来申请证书。

**1. 生成客户端的私钥**

```bash
# 为客户端生成一个 2048 位的 RSA 私钥
openssl genpkey -algorithm RSA -out client.key -pkeyopt rsa_keygen_bits:2048
```

- `client.key`: 生成的客户端私钥文件。这个私钥将由客户端自己保管。

**2. 创建客户端的证书签名请求 (CSR)**

CSR 文件包含了客户端的公钥和身份信息，它将被发送给 CA 进行签名。

```bash
openssl req -new \
  -key client.key \
  -out client.csr \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=My Team/CN=gameops"
```

- `client.csr`: 生成的 CSR 文件。

- ```
  -subj "..." : 客户端的身份信息。
  ```

  - `CN=user1`: 这里的 Common Name 通常用来标识用户或机器，你的后端服务可以解析这个字段来识别客户端身份。

------

### 阶段三：使用根 CA 签署客户端证书

这是最关键的一步。我们扮演 CA 的角色，使用 CA 的私钥来签署客户端的 CSR，从而生成最终的客户端证书。

**1. 创建一个扩展配置文件（非常重要！）**

为了让生成的证书明确用于“客户端认证”，我们需要一个简单的配置文件。

创建一个名为 `client_ext.cnf` 的文件，内容如下：

```ini
[v3_ext]
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
```

- `extendedKeyUsage = clientAuth`: 这行是核心！它明确指出该证书可用于客户端 TLS 身份验证。没有它，某些严格的服务器可能会拒绝该证书。

**2. 签署证书**

```bash
openssl x509 -req \
  -in client.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out client.crt \
  -days 3650 \
  -sha256 \
  -extfile client_ext.cnf \
  -extensions v3_ext
```

**命令解释：**

- `-req`: 表示输入是一个 CSR 文件。
- `-in client.csr`: 要处理的 CSR 文件。
- `-CA ca.crt -CAkey ca.key`: 指定用于签名的 CA 证书和私钥。
- `-CAcreateserial`: `openssl` 要求每个签发的证书都有唯一的序列号。此选项会自动创建并管理一个序列号文件（`ca.srl`）。第一次运行时会创建它。
- `-out client.crt`: 输出的最终客户端证书。
- `-days 365`: 客户端证书有效期（10年）。
- `-extfile` 和 `-extensions`: 应用我们刚才创建的配置文件中的扩展，确保证书用途正确。

### 验证和使用

至此，所有证书都已生成！

**1. 验证证书链**
我们可以验证客户端证书是否确实是由我们的 CA 正确签发的。

```bash
openssl verify -CAfile ca.crt client.crt
```

如果一切正常，你会看到输出：`client.crt: OK`

**2. 你现在拥有的文件及用途**

| 文件名           | 描述                 | 给谁用？                                                     |
| ---------------- | -------------------- | ------------------------------------------------------------ |
| **`ca.key`**     | **CA 的私钥**        | **绝密！自己保管，不要外泄。** 用于签发更多证书。            |
| **`ca.crt`**     | **CA 的公共证书**    | **给服务器 (Ingress-NGINX)**。放入 `auth-tls-secret` 中，用于验证客户端。 |
| **`client.key`** | **客户端的私钥**     | **给最终用户/客户端**。与 `client.crt` 配对使用。            |
| **`client.crt`** | **客户端的公共证书** | **给最终用户/客户端**。在发起 mTLS 请求时出示给服务器。      |
| `client.csr`     | 客户端证书请求       | 一次性文件，可以删除了。                                     |
| `ca.srl`         | CA 序列号文件        | 保留，用于签发下一个证书。                                   |
| `client_ext.cnf` | 扩展配置文件         | 保留，用于签发下一个客户端证书。                             |

**如何提供给用户？**

- **命令行工具 (如 curl):**
  直接提供 `client.key` 和 `client.crt` 文件。

  ```bash
  curl --key client.key --cert client.crt https://admin.yourcompany.com
  ```

- **浏览器:**
  浏览器通常需要一个 `.p12` (或 `.pfx`) 格式的文件，它将客户端证书和私钥打包在一起并用密码保护。

  ```bash
  # 将 client.key 和 client.crt打包成 client.p12
  openssl pkcs12 -export -out client.p12 -inkey client.key -in client.crt -name "User1 Client Certificate"
  ```

  执行此命令时，会提示你设置一个“导出密码”。用户在将 `client.p12` 文件导入到他们的浏览器或系统钥匙串时，需要输入这个密码。

现在你就可以将 `client.p12` 文件安全地分发给你的用户了！

### Ingress使用示例

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/auth-tls-verify-client: "on"
    nginx.ingress.kubernetes.io/auth-tls-secret: "roh5/ca-client"
    nginx.ingress.kubernetes.io/auth-tls-verify-depth: "1"
    nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream: "true"
    nginx.ingress.kubernetes.io/whitelist-source-range: "xxxxxx"
  name: tool-ingress
  namespace: roh5
spec:
  ingressClassName: nginx
  rules:
  - host: example.cn
    http:
      paths:
      - backend:
          service:
            name: tool-service
            port:
              number: 80
        path: /
        pathType: Prefix
  tls:
  - hosts:
    - example.cn
    secretName: xxx-tls

```

