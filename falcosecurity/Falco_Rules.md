# Falco Rules

接下来我们将了解：

-  Falco Rules语法。
- Falco 规则如何与标准安全框架保持一致。
- Falco 规则在不同文件中的组织方式。
- 利用默认的 Falco 规则监控可疑活动。

## 规则要点

Falco 规则用于指定特定的检测场景。我们之前讨论过这些规则如何包含满足条件时触发用户通知的条件，以及通知中将使用的输出信息。不过，规则还有其他字段。举例说明，请看一条简化的规则，它的目的是在容器内启动终端 shell 时发出警报：

```yaml
- rule: Terminal shell in container  
  desc: 容器中产生了一个交互式接口
  condition: > 
    #execve 和 execveat 事件代表 进程执行了一个新的程序，在 Linux 系统调用中，它们用于创建新的进程（包括 shell）。
    #事件的方向是 <（表示 "进入"），意味着该命令是 正在启动新的进程，而不是进程退出。
    evt.type in (execve, execveat) and evt.dir=<   
    #这个条件确保该事件 发生在容器内部，而不是宿主机上。
    and (container.id != host)  
    #proc.name 表示进程的名称。
    and proc.name in {bash, sh, zsh}  
  output: >  
    A shell was spawned in a container (user=%user.name  
    user_loginid=%user.loginid &container.info shell=%proc.name  
    parent=%proc.pname cmdline=%proc.cmdline container_id=%container.id)  
  priority: WARNING  
  tags: [container, shell, mitre_execution]  
```

| Rule Fields     | Description                                                  |
| :-------------- | :----------------------------------------------------------- |
| **rule\***      | 用简短的句子描述规则并对其进行唯一标识。                     |
| **desc\***      | 更长的描述，更详细地说明规则检测的内容。                     |
| **condition\*** | 为触发规则而必须满足的规则条件。                             |
| **output\***    | 规则触发时 Falco 发出的信息。                                |
| **priority\***  | 触发规则时生成警报的优先级。Falco 使用系统日志式优先级，因此该键可接受以下值：EMERGENCY（紧急）、ALERT（警报）、CRITICAL（重要）、ERROR（错误）、WARNING（警告）、NOTICE（通知）、INFORMATIONAL（信息）和 DEBUG（故障）。 |
| **source**      | 应用该规则的数据源。默认为 syscall。插件可定义自己的源类型（如 aws_cloudtrail 和 k8s_audit）。 |
| **enabled**     | 用于启用规则的布尔值键。默认为 true。禁用的规则不会被引擎加载，在 Falco 运行时也不需要任何资源。 |
| **exception**   | 导致规则不生成警报的一组例外情况。                           |
| **tags**        | 与此规则相关联的标记列表                                     |

✳号为必须值

### Macros

Falco 宏（Macros） 允许我们把规则中的条件拆分成 更小、更可复用的组件，提高 可读性**、**可维护性和 一致性。

为了让规则 **更清晰、可维护、可复用**，我们可以把 每个独立的逻辑拆分出来：

**1. 定义宏**

```
#✅ shell_procs`代表 `"proc.name in (bash, sh, zsh)"`
#✅ 作用：检查进程是否是 Shell（`bash`、`sh`、`zsh`）。
yaml复制编辑- macro: shell_procs
  condition: proc.name in (bash, sh, zsh)
```

```
- macro: container
  condition: (container.id != host)
```

```
#spawned_process 代表 "evt.type in (execve, execveat) and evt.dir=<"
#✅ 作用：检测 是否有新进程被创建。
- macro: spawned_process
  condition: evt.type in (execve, execveat) and evt.dir=<
```

**2. 使用宏优化规则**

有了宏之后，我们的 Falco 规则可以这样写：

```
- rule: Terminal shell in container
  desc: A shell has been spawned in a container.
  condition: spawned_process and container and shell_procs
  output: A shell was spawned in a container...
  priority: WARNING
  tags: [container, shell, mitre_execution]

```

### Lists

在 Falco 规则集中，`macros`（宏）和 `lists`（列表）被广泛使用：

- 宏用于定义可复用的过滤条件，类似于代码中的函数
- 列表类似于数组，存储多个值，可以在规则、宏等地方引用

列表的作用是让规则更清晰、易维护、可扩展，尤其当需要匹配多个值时。

**使用 `lists`（列表）优化规则**

我们可以用 `lists` 来存储 shell 名称：

**1. 定义 `lists`（列表）**

```
- list: shell_binaries
  items: [bash, sh, zsh]
```

`shell_binaries` 列表现在包含所有我们要检测的 shell 进程名称。
 未来如果要新增 shell 类型（如 `fish`、`dash`），只需要修改这里。

------

**2. 修改 `macros`（宏）使用 `lists`**

```
- macro: shell_procs
  condition: proc.name in (shell_binaries)
```

之前的 `"proc.name in (bash, sh, zsh)"` 直接换成 `"proc.name in (shell_binaries)"`。
 如果列表 `shell_binaries` 变化，`shell_procs` 也自动更新，不需要改动宏。

------

**3. 其他 `macros` 保持不变**

```
- macro: container
  condition: (container.id != host)
```

`container` 仍然用于判断进程是否在容器内运行。

```
- macro: spawned_process
  condition: evt.type in (execve, execveat) and evt.dir=<
```

`spawned_process` 仍然用于检测是否有新进程被创建。

**4. 规则（Rule）使用优化后的 `macros`**

```
- rule: Terminal shell in container
  desc: A shell has been spawned in a container.
  condition: spawned_process and container and shell_procs
  output: A shell was spawned in a container...
  priority: WARNING
  tags: [container, shell, mitre_execution]
```

## Rule Tagging

Falco 规则中的标签操作类似于 AWS 或 Kubernetes 等云环境中的资源标签。它的作用是对规则进行分类，以便于管理和更有效地利用。标签具有重要的应用价值：

1. 根据使用案例对规则进行分类，提高规则的可发现性和采用率。
2. 通过添加上下文信息来丰富输出数据，从而提高可解释性、过滤性和优先级。
3. 支持动态过滤，可以基于标签选择性地加载特定类别的规则（比如，只加载 `mitre_execution` 相关的规则，falco -t mitre_execution）。

**示例规则：`Terminal shell in container`**

来看一个默认规则示例：

```
- rule: Terminal shell in container
  desc: A shell was used as the entrypoint/exec point...
  condition: >
    spawned_process
    and container
    and shell_procs
    and proc.tty != 0
    and container_entrypoint
    and not user_expected_terminal_shell_in_container_conditions
  output: A shell was spawned in a container...
  priority: NOTICE
  tags: [maturity_stable, container, shell, mitre_execution, T1059]

```

这个规则用于**检测容器内是否有终端 shell 被打开**，并且**不符合预期的情况**（例如，不是用户主动运行的 shell）。
 最后，它带有多个标签，具体作用如下：

- `maturity_stable` —— 表示这个规则已经稳定，适用于生产环境。
- `container` —— 说明这个规则适用于容器环境。
- `shell` —— 这个规则专门用于检测 shell 相关活动。
- `mitre_execution` —— 这个规则属于 MITRE ATT&CK 框架中的 **Execution（执行）** 阶段。
- `T1059` —— 具体对应 **MITRE ATT&CK 技术编号 T1059**（Command and Scripting Interpreter，即命令和脚本解释器执行）。

### Default Tags

规则标签遵循[标准集](https://falco.org/docs/concepts/rules/controlling-rules/)，但并不是固定的。用户可以使用任何标签来标注规则，但应创建有意义的标签。最后，标签与 Falco 一起发展

### Rules Maturity Framework 

**Falco 0.36** 引入了**[规则成熟度框架](https://github.com/falcosecurity/rules/blob/main/CONTRIBUTING.md#rules-maturity-framework)**，目的是让用户更容易**评估和采用**默认的安全规则（非自定义规则）。这个框架将规则划分为不同的成熟度级别，帮助安全团队判断规则的适用性，以及是否需要额外的开发和调整。

**规则的四个成熟度级别：**

+ **maturity_stable (成熟稳定)**

  规则由专家严格审核，以确保其稳健性和最佳实践。它们主要侧重于通用的系统级检测，如通用反向外壳或容器逃逸，为不同行业的威胁检测建立了坚实的基线。

+ **maturity_incubating (成熟孵化)**

  规则针对特定用例，具有良好的稳健性，并遵循最佳实践，但可能并不普遍适用。预计这类规则将包括更多针对特定应用的规则。

+ **maturity_sandbox (成熟_沙箱)**

  规则是试验性的，符合最低接受标准，但仍在评估是否有更广泛的用途。

+ **maturity_deprecated**

  规则已过时或不太适用，保留供参考，但不积极支持。

规则成熟度框架使采用者更容易理解每条规则，并了解其在不同情况下的益处。此外，有了这个框架，采用者就能更清楚地了解哪些规则可以按原样采用，哪些规则可能需要大量工程工作才能评估和采用。

## Falco Rules Files

Falco 规则通常打包在规则文件中，Falco 在启动时读取这些文件。规则文件是一个 YAML 文件，可以包含一个或多个规则，每个规则都是 YAML 主体中的一个节点。

Falco 在 /etc/falco 目录中附带了一套由社区编辑的规则文件。Falco 启动时会自动加载默认规则文件。Falco 的每个新版本都会更新这些文件。

```shell
#启动时，Falco 会告诉你已经加载了哪些规则文件：
sudo falco --modern-bpf
Thu Sep 28 03:25:16 2023: Falco version: 0.36.0 (x86_64)
Thu Sep 28 03:25:16 2023: Falco initialized with configuration file: /etc/falco/falco.yaml
Thu Sep 28 03:25:16 2023: Loading rules from file /etc/falco/falco_rules.yaml
Thu Sep 28 03:25:16 2023: Loading rules from file /etc/falco/falco_rules.local.yaml
```

默认情况下，Falco 会加载两个文件：falco_rules.yaml 和 falco_rules.local.yaml。此外，Falco 还挂载 rules.d 目录，您可以使用该目录扩展规则集，而无需更改命令行或配置文件。

**需要注意的是：**文件加载顺序很重要。文件按配置的顺序加载（目录内的文件按字母顺序加载）

### 加载不同的规则文件

通常情况下，您需要加载自己的规则文件，而不是默认文件。有两种不同的方法

```
sudo falco -r my_rule1.yaml -r my_rule2.yaml
```

第二种方法是修改 Falco 配置文件（通常位于 /etc/falco/falco.yaml）中的 rules_file 部分。您可以添加、删除或修改该部分的条目，以控制 Falco 加载哪些规则文件：

```
rules_file:
  - /etc/falco/falco_rules.yaml
  - /etc/falco/falco_rules.local.yaml
  - /etc/falco/rules.d
```

请注意，使用这两种方法，您都可以指定一个目录，而不是单个文件。这非常方便，因为您只需更改目录内容，就可以添加或删除规则文件，而无需重新配置 Falco。

```
sudo falco -r ~/my_rules_dir
```

### Kubernetes 中加载规则文件

如果要在 Kubernetes 集群上部署 Falco，很可能会使用 Helm 进行安装。在这种情况下，与其直接在 /etc/falco/rules.d 目录中放置自定义规则文件，不如将其添加到提供给 helm 命令的 values.yaml 文件中

找到 customRules：{} 行，并将其替换为类似下面的配置：

```
customRules:  
  my_rules1.yaml: |-  
    - rule: Example rule  
      desc: ...  
    ...  
    - rule: Example rule 2  
      ...  
  my_rules2.yaml: |-  
    - rule: Example rule 3  
      ...  
```

这将指示 Helm 在 /etc/falco/rules.d 目录下的 Falco Pods 中创建尽可能多的可访问规则文件。

###  默认规则文件

如前所述，Falco 的默认规则文件通常安装在 /etc/falco 目录下。该目录包含的文件对 Falco 在不同环境下的运行至关重要。下表概述了其中最重要的文件。

| File Name                     | Description                                                  |
| :---------------------------- | :----------------------------------------------------------- |
| **rfalco_rules.yaml**         | Falco 的主规则文件，包含一套稳定的基于系统调用的主机和容器规则。 |
| **falco_rules.local.yaml**    | 在这里，您可以修改现有规则或添加自己的规则                   |
| **k8s_audit_rules.yaml**      | Kubernetes 审计日志规则文件。                                |
| **aws_cloudtrail_rules.yaml** | AWS CloudTrail 日志规则。                                    |
| **rules.d**                   | Falco 默认配置中包含了这个空目录。这意味着您可以在此目录下添加文件（或在此目录下创建规则文件的符号链接），Falco 会自动加载这些文件。 |

**需要注意的是：**默认规则文件 (falco_rules.yaml) 只包含稳定成熟度级别的规则。要启用其他成熟度级别，请使用 falcosidekick 安装它们

### Engine Version

在文本编辑器中打开 Falco 规则文件时，经常会遇到这样的首行：

**required_engine_version: 26**

虽然是可选项，但指明所需的最低引擎版本对于确保运行的 Falco 版本与所含规则之间的兼容性至关重要。旧版本的 Falco 可能不支持规则集中的字段，或者规则可能依赖于最近添加的系统调用。不正确的版本设置可能会导致失败或结果不准确。如果规则文件指定的引擎版本高于 Falco 支持的版本，Falco 将抛出错误并中止启动过程。

**需要注意的是：**Falco 随附的默认规则文件均已版本化。不要忘了在您的每个规则文件中也这样做！

### Plugin Version

除了指定引擎版本外，Falco 规则文件还可以使用 required_plugin_versions 字段定义兼容的插件版本。语法如下

```
- required_plugin_versions:  
  - name: <plugin_name>  
    version: <x.y.z>  
  - name: <other_plugin_name>  
    version: <x.y.z>  
```

这个可选字段包含一个对象列表，每个对象都有两个键值对：名称（name）和版本（version）。如果加载了插件，且其条目存在于 required_plugin_versions 中，则插件的版本必须符合相对于指定版本的语义版本控制（semver）

## 管理 Falco 规则

### falcoctl

更多内容请[参考](https://falco.org/blog/falcoctl-install-manage-rules-plugins/)

如前所述，falcoctl 是一种命令行界面（CLI）工具，旨在通过命令行界面管理 Falco 规则和插件。换句话说，falcoctl 可以帮助下载、安装和更新规则与插件，从而简化用户体验。它有四个主要组件：

- **Artifact**
  工件是 falcoctl 可以操作的元素（目前是规则文件和插件）。
- **Repository**
  资源库包含一件人工制品的一个或多个版本（标签）。
- **Registry**
  注册表存储了 falcoctl 通过一个或多个资源库操作的工件。
- **Index**
  YAML 文件，其中包含可用工件及其注册表和存储库的列表。

falcoctl 默认配置包含一个索引文件，指向 falcosecurity 组织官方支持的工具