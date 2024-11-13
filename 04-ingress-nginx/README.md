## 安装

```
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --version 4.11.3 \
  --namespace ingress-nginx --create-namespace \
  -f values.yaml

helm fetch ingress-nginx/ingress-nginx --version 4.11.3

helm  -n monitoring  upgrade --install monitoring-ingress-nginx  ingress-nginx-4.11.3.tgz   --values values.yaml
```
此目录下的内容参考于腾讯云
