## 说明
Loki为简单可扩展的模式
## 测试
在loki部署为多租户模式的时候，可以使用如下命令进行测试
```
curl -u "proj-hd-inland:xxxxxxxxx" -X POST "http://loki-gateway.monitoring.svc/loki/api/v1/push" \
  -H "Content-Type: application/json" \
  -H "X-Scope-OrgID: proj-hd-inland" \
  -d '{"streams":[{"stream":{"job":"test-data"},"values":[["'"$(date +%s)000000000"'","hi , everyone!"]]}]}'


curl -G "http://loki-gateway.monitoring.svc/loki/api/v1/query_range" \
  --data-urlencode 'query={job="test-data"}' \
  -H "X-Scope-OrgID: proj-hd-inland" \
  -u "proj-hd-inland:xxxxxxxxx" | jq .data.result
```
