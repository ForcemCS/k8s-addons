### buildkitd.toml

这个文件是用来配置 BuildKit 守护进程 (`buildkitd`) 的。你可以把它想象成是 Docker 守护进程 (`dockerd`) 的 `daemon.json` 文件，但它专门负责镜像构建。

文件分为几个部分：全局设置和多个特定功能的配置块（用 `[section]` 表示）。

```toml
# debug = true: 开启调试日志。会输出更详细的日志信息，方便排查问题。
debug = true
# 开启追踪日志。比 debug 更加详细，会记录每一个操作的细节，可能会影响性能，仅在深度调试时使用。
trace = true
# BuildKit 的“大本营”。所有状态、缓存、快照等数据都存放在这个目录下。
root = "/var/lib/buildkit"
# 允许在构建过程中使用一些“不安全”的特权。默认是禁用的，因为它们可能带来安全风险。
#    network.host: 允许构建容器使用主机的网络栈。
#    security.insecure: 允许构建步骤获得更高的权限。
insecure-entitlements = [ "network.host", "security.insecure", "device" ]

[log]
  # format = "text": 设置日志输出格式，可以是 json 或 text。
  format = "text"

[dns]
  #nameservers: 为构建容器指定 DNS 服务器。
  #options: DNS 解析选项。
  #searchDomains: DNS 搜索域。
  nameservers=["1.1.1.1","8.8.8.8"]
  options=["edns0"]
  searchDomains=["example.com"]

#gRPC 是 BuildKit 客户端（如 buildctl 或 docker）与守护进程通信的方式。
[grpc]
  #表示在所有网络接口的 1234 端口上监听 TCP 连接。
  address = [ "tcp://0.0.0.0:1234" ]
  #用于 Go 程序性能分析和调试的地址。
  debugAddress = "0.0.0.0:6060"
  #运行 gRPC 服务的用户和组 ID。
  uid = 0
  gid = 0
  #配置 TLS 加密，用于安全通信
  [grpc.tls]
    cert = "/etc/buildkit/tls.crt"
    key = "/etc/buildkit/tls.key"
    ca = "/etc/buildkit/tlsca.crt"

[otel]
  #  OpenTelemetry 配置，用于将构建的追踪数据导出到监控系统。
  socketPath = "/run/buildkit/otel-grpc.sock"

[cdi]
  # Container Device Interface 配置，用于向构建容器中注入硬件设备（如 GPU），默认禁用。
  disabled = true
  # List of directories to scan for CDI spec files. For more details about CDI
  # specification, please refer to https://github.com/cncf-tags/container-device-interface/blob/main/SPEC.md#cdi-json-specification
  specDirs = ["/etc/cdi", "/var/run/cdi", "/etc/buildkit/cdi"]

# config for build history API that stores information about completed build commands
[history]
  # 历史记录的最大保留时间（秒）。
  maxAge = 172800
  # 最多保留多少条历史记录。
  maxEntries = 50
  
#Worker 是实际执行构建步骤的后端。BuildKit 支持多种 Worker，最常见的是 oci 和 containerd。
#使用一个符合 OCI 标准的运行时（如 runc）来创建和管理构建容器。这是比较独立的模式。
[worker.oci]
  enabled = true
  # 手动指定这个 worker 支持的平台架构（如 linux/amd64）。
  platforms = [ "linux/amd64", "linux/arm64" ]
  snapshotter = "auto" # 底层文件系统快照技术，overlayfs 是最常用的。
  rootless = false # see docs/rootless.md for the details on rootless mode.
  # Whether run subprocesses in main pid namespace or not, this is useful for
  # running rootless buildkit inside a container.
  noProcessSandbox = false
  # 启用垃圾回收（Garbage Collection），自动清理不用的缓存。
  gc = true
  #非常重要的磁盘空间管理参数。定义了 BuildKit 的垃圾回收策略。
  #reservedSpace: GC 无论如何都会保留的最小空间。
  #maxUsedSpace: BuildKit 使用空间超过这个阈值，就会触发 GC。
  #minFreeSpace: GC 的目标是让磁盘可用空间达到这个值。
  reservedSpace = "30%"
  maxUsedSpace = "60%"
  minFreeSpace = "20GB"
  # alternate OCI worker binary name(example 'crun'), by default either 
  # buildkit-runc or runc binary is used
  binary = ""
  # name of the apparmor profile that should be used to constrain build containers.
  # the profile should already be loaded (by a higher level system) before creating a worker.
  apparmor-profile = ""
  # limit the number of parallel build steps that can run at the same time
  max-parallelism = 4
  # maintain a pool of reusable CNI network namespaces to amortize the overhead
  # of allocating and releasing the namespaces
  cniPoolSize = 16

  [worker.oci.labels]
    "foo" = "bar"

  [[worker.oci.gcpolicy]]
    # reservedSpace is the minimum amount of disk space guaranteed to be
    # retained by this policy - any usage below this threshold will not be
    # reclaimed during # garbage collection.
    reservedSpace = "512MB"
    # maxUsedSpace is the maximum amount of disk space that may be used by this
    # policy - any usage above this threshold will be reclaimed during garbage
    # collection.
    maxUsedSpace = "1GB"
    # minFreeSpace is the target amount of free disk space that the garbage
    # collector will attempt to leave - however, it will never be bought below
    # reservedSpace.
    minFreeSpace = "10GB"
    # keepDuration can be an integer number of seconds (e.g. 172800), or a
    # string duration (e.g. "48h")
    keepDuration = "48h"
    filters = [ "type==source.local", "type==exec.cachemount", "type==source.git.checkout"]
  [[worker.oci.gcpolicy]]
    all = true
    reservedSpace = 1024000000
#使用一个已经存在的 containerd 实例作为后端。如果你已经在用 containerd（比如 Kubernetes 的 CRI），这种方式可以更好地集成。
[worker.containerd]
  address = "/run/containerd/containerd.sock"
  enabled = true
  platforms = [ "linux/amd64", "linux/arm64" ]
  namespace = "buildkit"

  # gc enables/disables garbage collection
  gc = true
  # reservedSpace is the minimum amount of disk space guaranteed to be
  # retained by this buildkit worker - any usage below this threshold will not
  # be reclaimed during garbage collection.
  # all disk space parameters can be an integer number of bytes (e.g.
  # 512000000), a string with a unit (e.g. "512MB"), or a string percentage
  # of the total disk space (e.g. "10%")
  reservedSpace = "30%"
  # maxUsedSpace is the maximum amount of disk space that may be used by
  # this buildkit worker - any usage above this threshold will be reclaimed
  # during garbage collection.
  maxUsedSpace = "60%"
  # minFreeSpace is the target amount of free disk space that the garbage
  # collector will attempt to leave - however, it will never be bought below
  # reservedSpace.
  minFreeSpace = "20GB"
  # limit the number of parallel build steps that can run at the same time
  max-parallelism = 4
  # maintain a pool of reusable CNI network namespaces to amortize the overhead
  # of allocating and releasing the namespaces
  cniPoolSize = 16
  # defaultCgroupParent sets the parent cgroup of all containers.
  defaultCgroupParent = "buildkit"

  [worker.containerd.labels]
    "foo" = "bar"

  # configure the containerd runtime
  [worker.containerd.runtime]
    name = "io.containerd.runc.v2"
    path = "/path/to/containerd/runc/shim"
    options = { BinaryName = "runc" }

  [[worker.containerd.gcpolicy]]
    reservedSpace = 512000000
    keepDuration = 172800
    filters = [ "type==source.local", "type==exec.cachemount", "type==source.git.checkout"]
  [[worker.containerd.gcpolicy]]
    all = true
    reservedSpace = 1024000000

# 这部分配置 BuildKit 如何与 Docker Registry 交互，特别是处理镜像拉取、推送和缓存。
[registry."docker.io"]
  # mirror configuration to handle path in case a mirror registry requires a /project path rather than just a host:port
  mirrors = ["yourmirror.local:5000", "core.harbor.domain/proxy.docker.io"]
  # Use plain HTTP to connect to the mirrors.
  http = true
  # Use HTTPS with self-signed certificates. Do not enable this together with `http`.
  insecure = true
  # If you use token auth with self-signed certificates,
  # then buildctl also needs to trust the token provider CA (for example, certificates that are configured for registry)
  # because buildctl pulls tokens directly without daemon process
  #配置自定义的 CA 证书和客户端证书，用于认证。
  ca=["/etc/config/myca.pem"]
  [[registry."docker.io".keypair]]
    key="/etc/config/key.pem"
    cert="/etc/config/cert.pem"

# optionally mirror configuration can be done by defining it as a registry.
[registry."yourmirror.local:5000"]
  http = true
#Frontend 是 BuildKit 的一个重要概念。它负责将用户提供的构建指令（如 Dockerfile）转换成 BuildKit 内部的底层指令图（LLB）。

[frontend."dockerfile.v0"]
  enabled = true

[frontend."gateway.v0"]
  enabled = true
  # If allowedRepositories is empty, all gateway sources are allowed.
  # Otherwise, only the listed repositories are allowed as a gateway source.
  # 
  # NOTE: Only the repository name (without tag) is compared.
  #
  # Example:
  # allowedRepositories = [ "docker-registry.wikimedia.org/repos/releng/blubber/buildkit" ]
  allowedRepositories = []

[system]
  # how often buildkit scans for changes in the supported emulated platforms
  platformsCacheMaxAge = "1h"
```

