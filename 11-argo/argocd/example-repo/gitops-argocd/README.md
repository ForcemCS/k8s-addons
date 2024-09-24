## 第一部分
curl -sSL -o /usr/local/bin/argocd https://github.com/argoproj/argo-cd/releases/download/v2.4.11/argocd-linux-amd64
chmod +x /usr/local/bin/argocd

kubectl -n argocd get secrets argocd-initial-admin-secret -o json | jq .data.password -r | tr -d '\n'  | base64 -d


argocd app create solar-system-app-2 \
--repo https://3000-port-4d52f1c2a8ed4e5a.labs.kodekloud.com/bob/gitops-argocd.git \
--path ./solar-system \
--dest-namespace solar-system \
--dest-server https://kubernetes.default.svc


argocd app sync solar-system-app-2

## 第二部分
  data:
    resource.customizations.health.ConfigMap: |
      hs = {}
      hs.status = "Healthy"
       if obj.data.TRIANGLE_COLOR == "white" then
          hs.status = "Degraded"
          hs.message = "Use any color other than White "
       end
      return hs


kubectl patch configmap argocd-cm -n argocd --patch-file patch.yaml

piVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  # ... other application configurations ...
  syncPolicy:
    # ... other sync policy configurations ...
    prune: true  # 启用 prune
    syncOptions:
    - CreateNamespace=true
    - PruneLast=true
    - ApplyOutOfSyncOnly=true
    - RespectIgnoreDifferences=true
    - Force=true # 启用 force

#
