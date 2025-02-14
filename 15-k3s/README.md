## 说明
禁用traefik,可以使用ingress nginx
```
curl -sfL https://get.k3s.io | sh -s - --disable traefik
```
修改ingress nginx svc 为LB
```
kubectl patch svc ingress-nginx-controller -n ingress-nginx -p '{"spec":{"type":"LoadBalancer","ports":[{"port":80,"targetPort":80,"protocol":"TCP","name":"http"},{"port":443,"targetPort":443,"protocol":"TCP","name":"https"}]}}'

```
