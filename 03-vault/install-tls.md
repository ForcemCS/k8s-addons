### 安装

参见官方[教程](https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-minikube-tls)

```
helm install -n $VAULT_K8S_NAMESPACE $VAULT_HELM_RELEASE_NAME  --version 0.28.0  hashicorp/vault -f ./awskms.yaml
```

### CSR配置文件说明

```
cat > ${WORKDIR}/vault-csr.conf <<EOF
[req]
default_bits = 2048
prompt = no
encrypt_key = yes
default_md = sha256
distinguished_name = kubelet_serving
req_extensions = v3_req
[ kubelet_serving ]
O = system:nodes
CN = system:node:*.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = *.${VAULT_SERVICE_NAME}
DNS.2 = *.${VAULT_SERVICE_NAME}.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}
DNS.3 = *.${VAULT_K8S_NAMESPACE}
IP.1 = 127.0.0.1
EOF
```

**1. [req] 部分**

- `default_bits = 2048`： 指定生成的密钥长度为 2048 位。
- `prompt = no`： 禁止生成过程中出现交互式提示。
- `encrypt_key = yes`： 使用密码加密生成的私钥。
- `default_md = sha256`： 使用 SHA256 算法生成签名。
- `distinguished_name = kubelet_serving`： 指定使用 `[ kubelet_serving ]` 部分中定义的 Distinguished Name (DN) 信息。
- `req_extensions = v3_req`： 指定使用 `[ v3_req ]` 部分中定义的 X.509 v3 扩展。

**2. [ kubelet_serving ] 部分**

- `O = system:nodes`： 指定组织 (Organization) 为 "system:nodes"。

- ```
  CN = system:node:*.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}
  ```

  ： 指定 Common Name (CN) 为 "system:node:*.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}"。

  - `${VAULT_K8S_NAMESPACE}` 和 `${K8S_CLUSTER_NAME}` 是环境变量，分别表示 Vault 所在的 Kubernetes 命名空间和集群名称。
  - `*` 表示通配符，可以匹配任何子域名。

**3. [ v3_req ] 部分**

- `basicConstraints = CA:FALSE`： 指定该证书不能作为证书颁发机构 (CA)。

- ```
  keyUsage = nonRepudiation, digitalSignature, keyEncipherment, dataEncipherment
  ```

  ： 指定证书密钥用途，包括：

  - `nonRepudiation`: 不可否认性
  - `digitalSignature`: 数字签名
  - `keyEncipherment`: 密钥加密
  - `dataEncipherment`: 数据加密

- ```
  extendedKeyUsage = serverAuth, clientAuth
  ```

  1. **证书用途**: 该 CSR 请求的证书用途包括 `serverAuth` 和 `clientAuth`，这意味着生成的证书既可以用在服务器端进行身份验证，也可以用在客户端进行身份验证。

- `subjectAltName = @alt_names`： 指定使用 `[alt_names]` 部分中定义的 Subject Alternative Name (SAN) 信息。

**4. [alt_names] 部分**

1. **备用名称**: 证书包含多个备用名称 (SAN)，包括 DNS 名称和 IP 地址。这意味着客户端可以使用这些名称或 IP 地址中的任何一个来连接 Vault 集群。

- `DNS.1 = *.${VAULT_SERVICE_NAME}`： 添加一个 DNS SAN，匹配以 `${VAULT_SERVICE_NAME}` 结尾的任何域名。
- `DNS.2 = *.${VAULT_SERVICE_NAME}.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}`： 添加一个 DNS SAN，匹配以 `${VAULT_SERVICE_NAME}.${VAULT_K8S_NAMESPACE}.svc.${K8S_CLUSTER_NAME}` 结尾的任何域名。
- `DNS.3 = *.${VAULT_K8S_NAMESPACE}`： 添加一个 DNS SAN，匹配以 `${VAULT_K8S_NAMESPACE}` 结尾的任何域名。
- `IP.1 = 127.0.0.1`： 添加一个 IP SAN，值为本地回环地址。

**总结**

该配置文件定义了一个 CSR 请求，用于获取一个证书，该证书允许 Vault 服务在 Kubernetes 集群中与 kubelet 进行安全通信。证书包含了 Vault 服务的域名和 IP 地址信息，以及相关的密钥用途和扩展信息。

如何外部环境想访问vault集群的话

curl --cacert $WORKDIR/vault.ca    --header "X-Vault-Token: $VAULT_TOKEN"    https://127.0.0.1:8200/v1/secret/data/tls/apitest | jq .data.data

```
cat > ${WORKDIR}/overrides.yaml <<EOF
global:
   enabled: true
   tlsDisable: false
injector:
   enabled: true
server:
   extraEnvironmentVars:
      VAULT_CACERT: /vault/userconfig/vault-ha-tls/vault.ca
      VAULT_TLSCERT: /vault/userconfig/vault-ha-tls/vault.crt
      VAULT_TLSKEY: /vault/userconfig/vault-ha-tls/vault.key
   volumes:
      - name: userconfig-vault-ha-tls
        secret:
         defaultMode: 420
         secretName: vault-ha-tls
   volumeMounts:
      - mountPath: /vault/userconfig/vault-ha-tls
        name: userconfig-vault-ha-tls
        readOnly: true
   standalone:
      enabled: false
   affinity: ""
   ha:
      enabled: true
      replicas: 3
      raft:
         enabled: true
         setNodeId: true
         config: |
            cluster_name = "vault-integrated-storage"
            ui = true
            listener "tcp" {
               tls_disable = 0
               address = "[::]:8200"
               cluster_address = "[::]:8201"
               tls_cert_file = "/vault/userconfig/vault-ha-tls/vault.crt"
               tls_key_file  = "/vault/userconfig/vault-ha-tls/vault.key"
               tls_client_ca_file = "/vault/userconfig/vault-ha-tls/vault.ca"
            }
            storage "raft" {
               path = "/vault/data"
            }
            disable_mlock = true
            service_registration "kubernetes" {}
EOF

```

这段代码定义了一个 YAML 格式的配置文件 `overrides.yaml`，用于配置 Vault 服务，特别是用于 Kubernetes 环境下的高可用性 (HA) 部署。

**配置文件内容解释：**

- **global:**

  - **enabled: true**: 启用全局配置。
  - **tlsDisable: false**: 不禁用 TLS，意味着 Vault 通信将使用 TLS 加密。

- **injector:**

  - **enabled: true**: 启用 Injector，它负责将 Vault secrets 注入到 Pods 中。

- **server:**

  - extraEnvironmentVars

    : 为 Vault Server Pod 设置额外的环境变量。

    - **VAULT_CACERT**: Vault CA 证书路径。
    - **VAULT_TLSCERT**: Vault TLS 证书路径。
    - **VAULT_TLSKEY**: Vault TLS 密钥路径。

  - volumes

    : 定义 Pod 使用的卷。

    - **name: userconfig-vault-ha-tls**: 卷名为 `userconfig-vault-ha-tls`。

    - secret

      : 该卷类型为 secret，存储 Vault TLS 相关的证书和密钥。

      - **defaultMode: 420**: 设置卷的默认权限为 `0644`。
      - **secretName: vault-ha-tls**: 使用名为 `vault-ha-tls` 的 Kubernetes Secret 作为数据源。

  - volumeMounts

    : 定义卷在容器内的挂载点。

    - **mountPath: /vault/userconfig/vault-ha-tls**: 将卷挂载到容器内的 `/vault/userconfig/vault-ha-tls` 目录。
    - **name: userconfig-vault-ha-tls**: 使用名为 `userconfig-vault-ha-tls` 的卷。
    - **readOnly: true**: 将卷挂载为只读模式。

  - standalone

    :

    - **enabled: false**: 禁用单机模式，因为我们正在配置 HA 部署。

  - **affinity**: 设置 Pod 的亲和性规则，这里为空字符串，表示没有特殊要求。

  - ha

    :

    - **enabled: true**: 启用 HA 模式。

    - **replicas: 3**: 部署 3 个 Vault Server 副本，以实现高可用性。

    - raft

      :

      - **enabled: true**: 使用 Raft 协议进行共识和数据复制。

      - **setNodeId: true**: 自动设置 Raft 节点 ID。

      - config

        : Raft 配置块。

        - **cluster_name**: Raft 集群名称，设置为 `vault-integrated-storage`。

        - **ui**: 启用 Vault UI 界面。

        - listener "tcp"

          : 配置 TCP 监听器。

          - **tls_disable**: 禁用 TLS，设置为 `0` 表示启用 TLS。
          - **address**: 监听地址，设置为 `[::]:8200`，表示监听所有 IPv4 和 IPv6 地址的 8200 端口。
          - **cluster_address**: 集群内部通信地址，设置为 `[::]:8201`。
          - **tls_cert_file**: TLS 证书文件路径。
          - **tls_key_file**: TLS 密钥文件路径。
          - **tls_client_ca_file**: TLS 客户端 CA 证书路径。

        - storage "raft"

          : 配置 Raft 存储后端。

          - **path**: Raft 数据存储路径，设置为 `/vault/data`。

        - **disable_mlock**: 禁用内存锁定，设置为 `true` 表示禁用。

        - **service_registration "kubernetes"**: 使用 Kubernetes 进行服务注册。

**总结：**

该配置文件定义了一个使用 Raft 协议实现高可用性的 Vault 集群，并启用了 TLS 加密和 Kubernetes 服务注册。它还配置了 Vault Server Pod 的环境变量、卷和挂载点，以及 Raft 集群的网络和存储配置。

### 执行解密

更多信息请[参考](https://developer.hashicorp.com/vault/tutorials/kubernetes/kubernetes-minikube-tls)

```
kubectl exec -n $VAULT_K8S_NAMESPACE vault-0 -- vault operator init        -format=json > cluster-keys.json
kubectl  -n vault   exec  --stdin=true  --tty=true vault-0   vault operator unseal

需要输入三个密钥
```

#### Join `vault-1` and `vault2` pods to the Raft cluster

不需要执行解密

**vault-1**

```
kubectl exec -n $VAULT_K8S_NAMESPACE -it vault-1 -- /bin/sh
vault operator raft join -address=https://vault-1.vault-internal:8200 -leader-ca-cert="$(cat /vault/userconfig/vault-ha-tls/vault.ca)" -leader-client-cert="$(cat /vault/userconfig/vault-ha-tls/vault.crt)" -leader-client-key="$(cat /vault/userconfig/vault-ha-tls/vault.key)" https://vault-0.vault-internal:8200
```

**vault-2**

```
kubectl exec -n $VAULT_K8S_NAMESPACE -it vault-2 -- /bin/sh
vault operator raft join -address=https://vault-2.vault-internal:8200 -leader-ca-cert="$(cat /vault/userconfig/vault-ha-tls/vault.ca)" -leader-client-cert="$(cat /vault/userconfig/vault-ha-tls/vault.crt)" -leader-client-key="$(cat /vault/userconfig/vault-ha-tls/vault.key)" https://vault-0.vault-internal:8200
```

