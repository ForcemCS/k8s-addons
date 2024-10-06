## 说明

安装 [Argo Events](https://argoproj.github.io/argo-events/installation/)

在安装了events之后，我们将为 Webhook 设置传感器sensor和event-source。请参考[Here](https://argoproj.github.io/argo-events/tutorials/01-introduction/),目标是根据 HTTP Post 请求触发 Argo workflow。记得设置RBAC权限

然后通过ingress 公开事件源 pod

## Parameterization

### Webhook Event Payload

Webhook 事件源通过 HTTP 请求接收事件并将其转换为 CloudEvents。Webhook 传感器通过事件总线从事件源接收的事件结构如下、

```
{
    "context": {
      "type": "type_of_event_source",
      "specversion": "cloud_events_version",
      "source": "name_of_the_event_source",
      "id": "unique_event_id",
      "time": "event_time",
      "datacontenttype": "type_of_data",
      "subject": "name_of_the_configuration_within_event_source"
    },
    "data": {
      "header": {},
      "body": {},
    }
}
```

1. `Context` ：这是 CloudEvent 上下文，无论 HTTP 请求的类型如何，它都由事件源填充。
2. `Data` ：数据包含以下字段。
   + `Header` ：事件`data`中的`header`包含分派到事件源的 HTTP 请求中的标头。事件源从请求中提取标头并将其放入事件`data` `header`中。
   + `Body` ：这是 HTTP 请求的请求负载。
