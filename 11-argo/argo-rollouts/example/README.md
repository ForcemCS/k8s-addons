## Canary

当我们使用kubectl argo rollouts set image rollouts-demo rollouts-demo=argoproj/rollouts-demo:red 升级的时候，发现%20的流量有问题后
可以使用kubectl argo rollouts abort rollouts-demo终止升级，回到之前的版本，此时rollout的状态处于降级状态（Degraded）
我们需要再次执行kubectl argo rollouts set image rollouts-demo rollouts-demo=argoproj/rollouts-demo:yellow  回到健康的状态

## BlueGreen

下面描述了蓝绿更新期间发生的事件顺序。

1. 刚开始activeService和previewService都指向修订版 1 ReplicaSet。

2. 用户通过修改 pod 模板 ( `spec.template.spec` ) 来启动更新。

3. 创建的修订版 2 ReplicaSet 的大小为 0。

4. `previewService`被修改为指向修订版 2 ReplicaSet。 `activeService`仍然指向修订版 1。

5. 修订版 2 ReplicaSet 将缩放为`spec.replicas`或`previewReplicaCount` （如果设置）。

6. 一旦修订版 2 ReplicaSet Pod 完全可用， `prePromotionAnalysis`就会开始。

7. `prePromotionAnalysis`成功后，如果`autoPromotionEnabled`为 false 或`autoPromotionSeconds`非零，蓝色/绿色会暂停。

8. 推出可以由用户手动恢复，也可以通过超过`autoPromotionSeconds`自动恢复。

9. 如果使用了`previewReplicaCount`功能，则修订版 2 ReplicaSet 将缩放为`spec.replicas`

10. 此次推出通过更新`activeService`以指向它来“升级”修订版 2 ReplicaSet。此时，没有任何服务指向修订版 1

11. `postPromotionAnalysis`分析开始

12. 一旦`postPromotionAnalysis`成功完成，更新就会成功，并且修订版 2 ReplicaSet 将被标记为稳定。此次推出被认为已得到全面推广。

13. 等待`scaleDownDelaySeconds` （默认30秒）后，修订版本1的ReplicaSet被缩小

    
