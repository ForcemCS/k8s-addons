# Configuring Falco

接下来我们将

+ 了解并使用 Falco 的命令行选项和环境变量。
+ 根据具体需求优化配置文件。
+ 针对各种使用情况有效配置 Falco。

## Configuration Options

在 Falco 中，配置过程会根据您的部署方法而有所不同--您是直接在主机上运行，还是作为容器运行，或是作为 Kubernetes 集群的一部分运行。此外，您选择的配置界面类型，无论是环境变量、命令行选项还是配置文件，也会影响您的设置方式。不过，无论这些差异有多大，需要配置的基本设置都是一致的

**Configuration Interface** 

+ 环境变量

  设置环境变量有时是更改某些设置的唯一方法，例如，是否跳过**driver loader**。

+ 命令行参数

  命令行参数有时是更改某些设置的唯一方法。通过命令行配置的设置总是优先于从配置文件加载的设置。运行 falco --help 命令可获得 Falco 命令行选项的完整列表。Falco 会按字母顺序打印每个选项（以及简要说明）。可用选项可能会根据 Falco 版本的不同而有所变化

+ 配置文件

  如前所述，配置文件是一个 YAML 文件，默认位于 /etc/falco/falco.yaml。在该文件中，您可以配置规则文件、插件、输出设置和通道、暴露的服务、日志、性能等。默认配置文件包含对所有文件可用设置的详尽解释。根据 Falco 版本的不同，可用设置可能会有所变化

**Configuration Process** 

+ Host

  如果使用软件包管理器安装了 Falco，可以直接在 systemd 单元文件中指定命令行选项和环境变量，这些文件通常位于 /usr/lib/systemd/system/falco-*.service。

  如果不使用软件包管理器，运行 Falco 完全取决于自己，包括传递命令行选项和设置环境变量。在这种情况下，可以手动创建一个 systemd 单元,可以[参考](https://github.com/falcosecurity/falco/tree/master/scripts/systemd)

+ Container 

  Falco 的容器映像允许您指定要运行的命令，默认情况下是 /usr/bin/falco。如果需要传递命令行选项，可以通过容器运行时的 CLI 来实现。

  ```
  docker run --rm -it falcosecurity/falco /usr/bin/falco --version
  ```

  **请注意**，falcosecurity/falco 容器镜像的入口点是一个会尝试自动安装驱动程序的脚本。如果想跳过安装过程，需要将 SKIP_DRIVER_LOADER 环境变量设置为非空值。在 Docker 中，设置环境变量的一种方法是使用 -e 选项。例如，要同时获取版本和跳过驱动程序安装，可以运行

  ```
  docker run --rm -it -e SKIP_DRIVER_LOADER=y \ falcosecurity/falco /usr/bin/falco --version
  ```

  Falco 容器映像还捆绑了默认配置文件和默认规则文件。如果需要修改其中任何一个文件，通常的做法是制作该文件的外部副本（例如 /etc/falco/falco.yaml），然后将其挂载到容器中（-v 选项

+ Kubernetes

  通过清单在 Kubernetes 中部署 Falco 时，可在 DaemonSet 或部署清单中指定环境变量（env）和命令行选项（args）。配置文件设置和规则文件可使用 mount 或 ConfigMaps 进行配置。这个 [DaemonSet 模板](https://github.com/falcosecurity/deploy-kubernetes/tree/main/kubernetes/falco/templates)是一个很好的参考，已经配置了连接容器运行时和 Kubernetes API 服务器的所有选项。

## Command-line Options

1. 使用 -c 设置可使用不同于/etc/falco/falco.yaml 的配置文件

2. 启用或禁用规则：`-D`、`-T`、`-t`

   ```
   falco -D "Unexpected outbound connection"
   ```

   `-D <rule>` 禁用某条规则

   `-T <tag>` 启用某个标签下的所有规则

   `-t <tag>` 禁用某个标签下的所有规则

3. 输出中增加容器或 Kubernetes 信息：`-pc` 和 `-pk`

4. 加载规则文件：`-r`

5. 覆盖配置项：`-o`

   可以使用 `-o <key>=<value>` 的方式临时修改配置文件中的某些配置项。

   ```
   falco -o json_output=true
   ```

假设你在 Kubernetes 中运行 Falco，想加载自定义规则，输出 JSON，同时添加 K8s 信息，可以这样：

```
falco -r /etc/falco/my_rules.yaml -o json_output=true -pk
```

### Data Enrichment

Falco 除了监听系统调用（syscalls），它还会尝试去“丰富”事件的数据，比如容器、Pod、Kubernetes 的元信息。这需要它去连接 container runtime 和 Kubernetes API。

| Option                                     | Description                                                  |
| :----------------------------------------- | :----------------------------------------------------------- |
| **--cri <path>**                           | 使用此选项可指定与 CRI 兼容的容器运行时的 Unix 套接字的路径。如果多次指定，Falco 将按顺序尝试每个给定的路径，并使用第一个连接到的路径。不设置此选项时，Falco 将只尝试使用 /run/containerd/containerd.sock |
| --disable-cri-async                        | 这个选项一般不需要设置，但**如果你发现容器元数据时有时无**，可以试试这个选项。 |
| -k <url>          --k8s-api                | 这可以通过连接到 指定的 Kubernetes API 服务器来启用 Kubernetes 元数据丰富。或者，也可以使用 FALCO_K8S_API 环境变量，该变量接受的值与此选项允许的值相同。 |
| **-K**                      --k8s-api-cert | 使用该选项对 Kubernetes API 服务器进行身份验证。也可以使用 FALCO_K8S_API_CERT 环境变量，该变量接受的值与此选项允许的值相同。使用 --help 选项了解更多详情。 |
| **--k8s-node <node_name>**                 | 该选项可对 Kubernetes 元数据充实进行重要的性能优化。当从 API 服务器请求 Pod 的元数据时，Falco 会使用节点名称作为过滤器，从而丢弃来自其他节点的不必要元数据。应始终设置该选项。否则，Falco 虽然能正常工作，但在大型集群上可能会出现性能问题。 |

### 调试和故障排除

到目前为止，我们所介绍的命令行选项都是您在操作 Falco 时可能会经常用到的。不过，下面列出的另一组选项更适合偶尔使用，比如当你需要有关 Falco 安装的信息或试图解决问题时。可以[参考](https://falco.org/docs/troubleshooting/)

## Configuration File Settings

如前所述，配置文件是一个 YAML 文件，默认位于 /etc/falco/falco.yaml。在该文件中，您可以配置rules files, plugins, output settings and channels, exposed services, logging, performance等

如前所述，Falco 会加载 rules_file 中定义的规则文件（顺序很重要）。**plugins**设置可定义并配置一个或多个插件，但您还必须将其添加到 load_plugins 设置中，以启用该插件。默认情况下，Falco 会监控配置和规则文件的变化，并在检测到任何修改时自动重新加载以应用更新的配置。将 watch_config_files 设置为 false 即可禁用。

### Falco Output Settings

启用 time_format_iso_8601 后，Falco 会以 ISO8601 格式显示日志和输出信息中的时间。默认情况下，时间以本地时区显示。优先级设置可让您根据规则的严重性对其进行过滤和控制，确保只有特定优先级或更高的规则才会被 Falco 激活和评估（默认值为 debug）。

规则条件按照规则文件中定义的顺序进行评估。默认情况下，只有第一个匹配规则会触发，可能会影响其他规则。将实验性的 rule_matching 设置从 first（默认）改为 all，可允许同一事件有多个匹配规则（规则仍按其定义顺序触发）。这消除了基于 "先匹配者胜 "原则的规则优先级问题。不过，启用全部匹配选项可能会导致性能下降。我们建议在生产中部署前仔细测试这一替代设置。

### Falco Logging

默认情况下，Falco 与功能、设置和潜在错误相关的日志会被导入 stderr (log_syslog: true) 和 syslog (log_syslog:true)。log_level 设置（默认为 info）决定了 Falco 与其运行相关的日志的最低日志级别。这些日志与 Falco 的警报输出无关，而是与 Falco 的生命周期有关。

当输出通道未能在指定期限内发送警报时，就会发生超时错误。output_timeout 参数以毫秒为单位指定了在认为超过截止时间之前的等待时间（默认为 2000）。换句话说，Falco 输出的用户可以阻塞 Falco 输出通道长达 2 秒，而不会触发超时通知。

Falco 利用内核和用户空间之间的共享缓冲区来接收系统调用事件。不过，在某些情况下，由于读取事件或需要跳过特定事件的问题，底层库可能会出现超时。虽然 Falco 很少遇到连续事件超时的情况，但它有能力检测这种情况。您可以配置无事件连续超时的最大次数（syscall_event_timeouts.max_consecutives），之后 Falco 将生成警报。默认值为连续 1000 次超时且未收到任何事件。请注意，这需要将 Falco 的 log_level 设置为最小notice级别。

### Syscall Event Drops 

Falco 在内核和用户空间之间使用共享缓冲区来传递系统调用信息。当 Falco 检测到该缓冲区已满且系统调用已被丢弃时，它会采取以下一种或多种措施：

- gnore：什么也不做（当操作列表为空时的默认值）
- **log**: 记录一条 DEBUG 消息，指出缓冲区已满
- **alert**: 发出 Falco 警报，指出缓冲区已满
- **exit**: 以非零的返回代码退出 Falco

当丢弃的系统调用占上一秒事件数的百分比大于给定阈值（范围为 [0, 1] 的双倍值）时，就会发出消息。如果想在出现掉线时收到警报，请将阈值设为 0。为了调试或测试，可以使用 simulate_drops: true 来模拟掉线。在这种情况下，阈值并不适用。默认情况下，syscall_event_drops 通过以下配置启用，但要求 Falco 规则配置优先级设置为debug。

```yaml
syscall_event_drops:  
  threshold: .1  
  actions:  
    - log  
    - alert  
  rate: .03333  
  max_burst: 1  
  simulate_drops: false  
```

此设置会在指定时间段内根据设置发出多次名为“Falco内部：系统调用事件丢弃”的Falco规则。此处的统计数据反映1秒时间段的增量。如果您更倾向于以固定间隔定期查看单调计数器的指标，包括系统调用丢弃统计数据和其他指标，请查看指标配置选项，本课程未涉及该选项，但配置文件中对此进行了详细说明。

### Falco Performance Tuning Settings

下面是性能调整设置的摘要。如需了解更多信息，请阅读配置文件，因为其中包含每个设置的大量详细信息。

+ syscall_buf_size_preset

  syscall_buf_size_preset 设置控制着系统调用缓冲区的大小。该设置使用一个索引系统，每个索引代表一个缓冲区大小的 2 次幂，从索引 1 的 1 MB 开始（默认值为 4，即 8 MB）。缓冲区大小还必须满足与系统页面大小相关的限制条件。例如，页面大小为 1 MB 的系统不能使用小于 4 MB 的缓冲区。该设置允许对性能和可靠性进行微调；增加缓冲区大小可以减少系统调用中断，但可能会降低系统速度。相反，减小缓冲区大小可以加快系统速度，但有可能增加系统调用中断。建议只有在默认设置无法满足您的要求时才调整该参数。

  **请注意**，每个缓冲区在虚拟内存中映射两次，实际上使其占用空间增加了一倍。

+ syscall_drop_failed_exit

  启用试验性的 syscall_drop_failed_exit（系统调用失败退出）设置（默认禁用）后，系统调用失败退出事件在推送到环形缓冲区之前，会在内核驱动程序中被丢弃。这一优化可以降低 CPU 占用率，提高环形缓冲区的使用效率，从而减少事件丢失的次数。但需要注意的是，启用该选项也意味着要牺牲系统的部分可见性。

+ modern_bpf.cpus_for_each_syscall_buffer

  modern_bpf.cpus_for_each_syscall_buffer 设置仅适用于 modern_bpf 驱动程序。它控制为单个系统调用缓冲区分配多少 CPU。该参数对于优化内存分配和性能至关重要，尤其是在 Kubernetes 等受限环境中。默认情况下，该设置将两个 CPU 映射到一个系统调用缓冲区（1:2 映射）。这与传统的 eBPF 驱动程序不同，后者默认为 1:1 映射。该索引的设置范围从 0 到最大在线 CPU 数量。值为 0 意味着所有 CPU 共享一个缓冲区。该设置旨在平衡资源分配和系统速度，减少内存占用和 CPU 竞争。更改该参数时应慎重，因为缓冲区过少可能会导致内核级争用和性能下降。

  