## 需求

假设我们的开发环境在一台ubuntu服务器安装了K3S，但是没有公网IP地址。我们可以通过一台云服务商带有public IP的虚拟机实现转发。

1. 我们K3S内网的一个服务，例如center-hdh5.sandbox-204.h.xinghuihuyu.cn

   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     annotations:
       meta.helm.sh/release-name: center
       meta.helm.sh/release-namespace: hdh5
     creationTimestamp: "2025-02-19T07:05:40Z"
     generation: 1
     labels:
       app.kubernetes.io/managed-by: Helm
     name: center-ro3-micro-ingress
     namespace: hdh5
     resourceVersion: "194480"
     uid: a00c31d9-2220-44c4-958d-a13ba7d18f69
   spec:
     ingressClassName: nginx
     rules:
     - host: center-hdh5.sandbox-204.h.xinghuihuyu.cn
       http:
         paths:
         - backend:
             service:
               name: center-ro3-micro-svc
               port:
                 number: 18000
           path: /
           pathType: Prefix
   ```

2. 然后在这台服务器上启动frpc

   ```ini
   root@debian-12:~/frpc# cat  /etc/systemd/system/frpc.service 
   [Unit]
   Description=Frp Client Service
   Documentation=https://github.com/fatedier/frp
   Wants=network-online.target
   After=network-online.target
   
   [Install]
   WantedBy=multi-user.target
   
   [Service]
   Type=simple
   # Having non-zero Limit*s causes performance problems due to accounting overhead
   # in the kernel. We recommend using cgroups to do container-local accounting.
   LimitNOFILE=1048576
   #LimitNPROC=infinity
   #LimitCORE=infinity
   #TasksMax=infinity
   #TimeoutStartSec=0
   Restart=on-failure
   RestartSec=5s
   #ExecStartPre=/sbin/modprobe br_netfilter
   #ExecStartPre=/sbin/modprobe overlay
   ExecStart=/usr/bin/frpc -c /root/frpc/frpc.toml
   
   ```

3. 让center-hdh5.sandbox-204.h.xinghuihuyu.cn解析到公网地址上

4. 假设这台公有云EC2上也部署了K3S，并且作为frps(假设认为是30001传输数据的，请查看配置文件)

   ```ini
   [root@VM-12-16-centos frps]# cat  /etc/systemd/system/frps.service 
   [Unit]
   Description=Frp Server Service
   Documentation=https://github.com/fatedier/frp
   Wants=network-online.target
   After=network-online.target
   
   [Install]
   WantedBy=multi-user.target
   
   [Service]
   Type=simple
   # Having non-zero Limit*s causes performance problems due to accounting overhead
   # in the kernel. We recommend using cgroups to do container-local accounting.
   LimitNOFILE=1048576
   #LimitNPROC=infinity
   #LimitCORE=infinity
   #TasksMax=infinity
   #TimeoutStartSec=0
   Restart=on-failure
   RestartSec=5s
   #ExecStartPre=/sbin/modprobe br_netfilter
   #ExecStartPre=/sbin/modprobe overlay
   ExecStart=/usr/bin/frps -c /root/frps/frps.toml
   ```

5. 现在我们可以认为是访问<公网:30001>就是转发到了内网的服务。但是现在必须执行curl -kivL -H 'Host: centerwx.h2.xinghuihuyu.cn' 

6. 因为我们EC2是也部署了K3S，所以可以用traefik再次做转发，具体详情请惨开配置文件