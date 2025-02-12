## 说明
Loki为简单可扩展的模式
## 测试
在loki部署为多租户模式的时候，可以使用如下命令进行测试
```
curl   -u "user:password" -X POST "http://loki-gateway/loki/api/v1/push" \
  -H "Content-Type: application/json" \
  -H "X-Scope-OrgID: proj-hd" \
  -d '{
    "streams": [
      {
        "stream": {
          "job": "manual-push"
        },
        "values": [
          ["1680000000000000000", "hello, loki!"]
        ]
      }
    ]
  }'
```
