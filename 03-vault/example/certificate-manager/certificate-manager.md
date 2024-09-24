## Configure PKI secrets engine

1. 在其默认路径中启用 PKI 秘密引擎。

2. 将最大租约生存时间 (TTL) 配置为 `8760h` 。

3. 生成一个新的根证书颁发机构（CA）证书(内部使用的CA)

   ```
   vault write pki/root/generate/internal \
       common_name=threshold.com \
       ttl=8760h
   
   ```

4. 配置 PKI 秘密引擎证书颁发和证书吊销列表 (CRL) 端点

   ```
   vault write pki/config/urls \
       issuing_certificates="http://vault.vault:8200/v1/pki/ca" \
       crl_distribution_points="http://vault.vault:8200/v1/pki/crl"
   ```

5. 配置名为 `threshold-dot-com` 的角色，该角色允许创建带有任何子域的证书 `threshold-dot-com` 域。

   ```
   vault write pki/roles/threshold-dot-com \
       allowed_domains=threshold.com \
       allow_subdomains=true \
       max_ttl=4320h
   ```

6. 创建名为 `pki` 的策略，以启用对 PKI 机密引擎路径的读取访问。

   ```
   $ vault policy write pki - <<EOF
   path "pki*"                        { capabilities = ["read", "list"] }
   path "pki/sign/threshold-dot-com"    { capabilities = ["create", "update"] }
   path "pki/issue/threshold-dot-com"   { capabilities = ["create"] }
   EOF
   ```

## Configure Kubernetes authentication

请参考官方文档

1. 创建名为 `issuer` 的 Kubernetes 身份验证角色，将 `pki` 策略与名为 `issuer` 的 Kubernetes 服务帐户绑定。

   ```
   vault write auth/kubernetes/role/issuer \
       bound_service_account_names=issuer \
       bound_service_account_namespaces=threshold \
       policies=pki \
       ttl=20m
   ```

## Deploy Cert Manager

请参考官方文档

## Configure an issuer and generate a certificate

cert-manager 使用 `Issuer` 资源来定义如何获取证书。当你想要使用 Vault 生成证书时，你需要创建一个 `Issuer` 对象，并将其配置为连接到你的 Vault 服务器。

