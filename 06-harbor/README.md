```
root@node01:/etc/containerd/certs.d# ls
harbor.threshold.com:31339  hosts.toml  xxxxxxxxxxx:5000
root@node01:/etc/containerd/certs.d# cat harbor.threshold.com\:31339/hosts.toml 
server = "https://harbor.threshold.com:31339"

[host."https://harbor.threshold.com:31339"]
  ca = "/etc/certs.d/harbor.threshold.com.crt"
root@node01:/etc/containerd/certs.d# cat xxxxxxxxxxxxxxxxxxxxxxxxxx\:5000/hosts.toml 
server = "https://xxxxxxxxxxxxxxxxxxxxx.com:5000"

[host."https://xxxxxxxxxxxxxxxxxxxxx.com:5000"]
  ca = "/etc/certs.d/tx.crt"
```

```
root@docker:/etc/docker# ls
certs.d  daemon.json
root@docker:/etc/docker# cat daemon.json 
{   "insecure-registries" : ["xxxxxxxxxxxxxxxxxxxxx:5000","harbor.threshold.com:31339"],
    "registry-mirrors": ["https://docker.m.daocloud.io"]
}
root@docker:/etc/docker# cd certs.d/
root@docker:/etc/docker/certs.d# ls
xxxxxxxxxxxxxxxxxxxxx:5000
root@docker:/etc/docker/certs.d# cd xxxxxxxxxxxxxxxxxxxxx\:5000/
root@docker:/etc/docker/certs.d/xxxxxxxxxxxxxxxxxxxxx:5000# ls
xxxxxxxxxxxxxxxxxxxxx.crt

```

上边的代码分别是containerd和docker运行时认证到私有仓库的一些参考示例