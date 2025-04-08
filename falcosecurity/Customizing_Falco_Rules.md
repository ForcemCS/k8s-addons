# 自定义Falco Fules

接下来我们将了解：

- 定制 Falco 规则的不同选项。
- 自定义 Falco 默认规则，以适应您的安全要求。

Falco 预装了一套丰富且不断增加的规则，涵盖了许多重要的使用案例。这些规则被编排在不同的文件中。正如之前所 讲的，Falco 提供了一种创建自己的规则的简便方法，你也可以将这些规则整理到不同的文件中。

在许多情况下，您可能会发现定制现有规则是有益的。例如，您可能希望降低某些规则的噪音，或者您可能希望扩大某些 Falco 检测的范围，以便更好地匹配您的环境。

处理这些情况的一种方法是编辑规则文件。这种方法对默认规则文件效果不佳，因为它们会自动更新，这可能会覆盖您的更改。此外，如果您升级了 Falco，还需要确保您的更改被合并到新文件中。Falco 提供了一种更通用的自定义规则的方式，旨在使您的更改在不同版本之间可维护和重复使用。让我们来看看它是如何工作的。

以下规则可检测 vim 或 nano 文本编辑器是否以 root 身份打开：

```
- list: editors  
  items: [vi, nano]  

- macro: editor_started  
  condition: (evt.type = execve and proc.name in (editors))  

- rule: Text Editor Run by Root  
  desc: the root user opened a text editor  
  condition: editor_started and user.name=root  
  output: root user started an editor (cmd=$proc.cmdline)  
  priority: WARNING  
```

如果我们将该规则保存在名为 my_rules.yaml 的规则文件中，就可以在 Falco 中加载该文件来测试该规则：

```
sudo falco -r my_rules.yaml
```

接下来，我们将了解在不触及 my_rules.yaml 文件的情况下自定义规则的不同方法。

## Append

假设我们想在列表中添加新的编辑器。Falco 提供了 append 关键字，可以将值追加到同一列表中。语法如下

```
- list: editors  
  items: [emacs, subl]  
  append: true  
```

我们可以将此列表保存在名为 other_rules.yaml 的文件中。现在，如果我们运行以下命令行，Falco 将检测 vi、nano、emacs 和 subl 的 root 执行：

```
sudo falco -r my_rules.yaml -r other_rules.yaml
```

**注意：**加载文件的顺序很重要。文件按照配置的顺序加载（目录内的文件按照字母顺序加载）。用户可以自定义预定义的规则，只需在出现在列表后面的文件中覆盖这些规则即可

您也可以对宏和规则进行追加（使用相同的语法）。不过，有几件事需要注意。对于规则，**只能追加到 condition**。附加到其他键（如输出）的尝试将被忽略。此外，**追加的内容会直接接到原始 condition 的末尾**。例如，假设我们扩展了示例中的规则条件，对其进行了如下追加：

```
- rule: Text Editor Run by Root
  condition: or user.name = pablo
  append: true

##最终的条件如下

condition: editor_started and user.name=root or user.name = pablo
```

## Disable

您经常会遇到需要禁用规则集中一条或多条规则的情况。例如，它们太吵或与您的环境无关。禁用的规则不会被引擎加载，在 Falco 运行时也不需要任何资源。Falco 提供了多种禁用规则的方法。

禁用规则的一种方法是将**enabled**字段设置为 **false**。启用是一个可选的规则字段，用于定义规则是启用（true）还是禁用（false）。如果缺少此键，则假定 enabled 为 true。要禁用一条规则，只需写入规则名称并将 enabled 设为 false。例如，要禁用 Text Editor Run by Root 规则，请在 other_rules.yaml 文件中添加以下内容：

```
- rule: Text Editor Run by Root
  enabled: false
```

另一种禁用规则的方法是通过命令行。Falco 提供了两种通过命令行禁用规则的方法。第一种是 -T 标志，它可以禁用带有给定标签的规则。命令行中可以多次使用 -T 来禁用多个标记。例如，要跳过所有带有 k8s 标记、cis 标记或两种标记的规则，可以这样运行 Falco：

```
sudo falco -T k8s -T cis
```

从命令行禁用规则的第二种方法是使用 -D 标志。-D 会禁用所有名称中包含 的规则。-D 也可以用不同的参数多次指定。比如：

```

sudo falco -D Editor
```

这条命令会禁用所有规则名中包含 **Editor** 的规则，比如：

- `Text Editor Run by Root`
- `Editor Executed Unexpectedly`

但不会匹配 `editor`（小写），因为匹配是区分大小写的。

如果使用 Helm，这些参数也可以指定为 Helm 图表值（extraArgs）。

```
helm install falco --set "extra.args={-D,Editor}" falcosecurity/falco
```

但注意：**这个 Helm 的语法要求很严格！**不能有空格

## Replace 

替换列表、宏或规则只需重新声明即可。第二个声明可以在同一文件中，也可以在包含原始声明的文件之后加载的单独文件中。要更改规则以支持不同的文本编辑器，请在 other_rules.yaml 中添加以下内容：

```
- list: editors
  items: [emacs, subl]
```

注意我们是如何重新定义编辑器列表内容的，用 emacs 和 subl 代替了原来的命令名称。现在在原始规则文件后加载 other_rules.yaml：

```
sudo falco -r my_rules.yaml -r other_rules.yaml
```

当 root 运行 emacs 或 subl，但不运行 vi 或 nano 时，Falco 会识别编辑器的第二个定义并生成警报。从本质上讲，我们已经替换了列表的内容。宏和规则也是如此。

## 实验

```
11:41:28.678808064: Critical Executing binary not part of base image (proc_exe=mount proc_sname=kubelet gparent=systemd proc_exe_ino_ctime=1744083442546378144 proc_exe_ino_mtime=1732219314000000000 proc_exe_ino_ctime_duration_proc_start=246132156145 proc_cwd=/ container_start_ts=1744083442288725073 evt_type=execve user=root user_uid=0 user_loginuid=-1 process=mount proc_exepath=/usr/bin/mount parent=kubelet command=mount -t tmpfs -o size=3900100608 tmpfs /var/lib/kubelet/pods/5acf0a36-8563-4126-b23b-4069dafa3e73/volumes/kubernetes.io~projected/kube-api-access-24fqt terminal=0 exe_flags=EXE_WRITABLE|EXE_UPPER_LAYER container_id=1375806db04c container_name=kind-control-plane)
```

事件类型：

```
Critical Executing binary not part of base image
```

Falco 检测到：一个**不属于容器基础镜像的二进制文件被执行了**，这是一个典型的入侵或越权行为检测项，防止在容器中运行未经批准的程序。

上边的事件中有关键字**binary**，可以在[此页面](https://falco.org/docs/reference/rules/default-rules/)搜索相关信息

```
- rule: Drop and execute new binary in container
  override:
    condition: append # 表示追加到现有规则的 condition
  condition: and not (proc.name=mount and proc.aname[2] startswith runc)
  # 注意：不再需要顶层的 append: true
```
