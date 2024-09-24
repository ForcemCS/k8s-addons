### Vault agent injector TLS configuration

以下说明演示如何使用K8S CA手动配置 Vault Agent 注入器。

1. First, create a private key for the certificate:

   ```
   $ openssl genrsa -out tls.key 2048
   ```

2. create a certificate signing request (CSR) to be used when signing the certificate:

   ```
   openssl req \
      -new \
      -key tls.key \
      -out tls.csr \
      -subj "/C=US/ST=CA/L=San Francisco/O=HashiCorp/CN=vault-agent-injector-svc"
   ```

3. 创建 CSR 后，创建扩展文件以配置用于签署证书的其他参数。

   ```
   $ cat <<EOF >csr.conf
   authorityKeyIdentifier=keyid,issuer
   basicConstraints=CA:FALSE
   keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
   subjectAltName = @alt_names
   
   [alt_names]
   DNS.1 = vault-agent-injector-svc
   DNS.2 = vault-agent-injector-svc.vault
   DNS.3 = vault-agent-injector-svc.vault.svc
   DNS.4 = vault-agent-injector-svc.vault.svc.cluster.local
   EOF
   ```

4. 签署证书：

   ```
   openssl x509 \
     -req \
     -in tls.csr \
     -CA injector-ca.crt \
     -CAkey injector-ca.key \
     -CAcreateserial \
     -out tls.crt \
     -days 1825 \
     -sha256 \
     -extfile csr.conf
   
   ```

### Configuration 

```
kubectl create secret generic injector-tls \
    --from-file tls.crt \
    --from-file tls.key \
    --namespace=vault
export CA_BUNDLE=$(cat injector-ca.crt | base64)
```

