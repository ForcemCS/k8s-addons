## 说明

对普通 Kubernetes 秘密进行加密，并将其转换为加密secrets

```
kubeseal < unsealed_secrets/db-creds.yml > sealed_secrets/db-creds-encrypted.yaml -o yaml
kubeseal < unsealed_secrets/paypal-cert.yml > sealed_secrets/paypal-cert-encrypted.yaml -o yaml
```

然后执行kubectl apply  
