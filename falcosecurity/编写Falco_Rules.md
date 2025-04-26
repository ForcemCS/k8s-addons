## 编写Falco Rules

接下来我们将能够：

+ Replicate the desired detection event.

+ 按照建议的方法编写一条新规则。
+ 避免编写 Falco 规则时的常见错误。

虽然 Falco 的默认规则集非常丰富，而且还在不断扩展，但经常会遇到需要对这些规则进行自定义的情况，或者需要制定新规则的情况。编写新规则的核心是精心设计条件和输出，因此在概念上是一个非常简单的过程。但在实际操作中，有几个因素需要考虑。Falco Book中提出的规则开发方法包括九个步骤(这里省略了8和9)

1. 复现要检测的事件。
2. 捕捉事件并将其保存到跟踪文件中。
3. 借助 sysdig 制作和测试条件过滤器。
4. 借助 sysdig 制作并测试输出结果。
5. 将 sysdig 命令行转换为规则。
6. 在 Falco 中验证规则。
7. 模块化并优化规则。

下面说明如何创建一条新规则，检测在 /proc、/bin 和 /etc 目录中创建符号链接的尝试。

### 复现事件以检测

如果不对规则进行测试和验证，几乎不可能创建出可靠的规则，因此第一步是重新创建规则应检测的场景。在本例中，要检测的是在三个特定目录中创建符号链接。使用 ln 命令在终端中重新创建该场景：

```
ln -s ~ /proc/evillink
ln -s ~ /bin/evillink
ln -s ~ /etc/evillink
```

### **捕捉事件并将其保存到跟踪文件中**

现在使用 sysdig 捕捉可疑活动。sysdig 允许使用 -w 命令行标志将活动存储在跟踪文件中：

```
sysdig -w evillinks.scap
```

在另一个终端，再次运行三条 ln 命令，然后回到第一个终端，用 Ctrl-C 停止 sysdig。现在，这些活动都记录在跟踪文件中，可以根据需要多次查看：

```
sysdig -r evillinks.scap
```

请注意，跟踪文件很大，包含了主机的所有活动，而不仅仅是 ln 命令。在运行捕获时，使用过滤器可使文件更小、更易于检查：

```
sysdig -w evillinks.scap proc.name=ln
```

新的捕获是一个大小小于 1 MB 的无噪音文件，只包含制定规则所需的特定活动。在跟踪文件中保存触发规则的活动有几个优点：

+ **复杂行为只需要复现一次**：有些可疑行为不像 `ln` 命令这么简单，捕获一次就可以反复分析，不需要每次都重新执行。

+ **可以集中精力分析事件**：不用频繁在多个终端来回切换执行命令。

+ **可以在别的机器上开发规则**：比如你在一台云服务器或者边缘设备上捕获 trace 文件后，可以把文件拷贝到自己的开发机上进行规则开发，不需要在目标机器上部署配置 Falco。

+ **普通用户权限就能完成**：不需要 root 权限，也能进行捕获和分析。

+ **分析过程更一致、可回溯**：不仅方便开发规则，也方便后续做回归测试。

### **用 sysdig 来编写和测试条件过滤器**

前面我们已经用 `sysdig` 把可疑的系统调用活动记录下来了，接下来要做的是**编写检测条件（filter）**。在这一步，通常要先回答两个关键问题：

1. **需要针对哪种类型的系统调用？**
   - 不是所有 Falco 规则都是基于系统调用的（有些基于插件），但大多数情况下，首先要确定**哪类事件**应该触发规则。
2. **在这些事件里，需要检查哪些参数或字段？**
   - 就是要知道，在哪些字段中判断是否存在可疑行为。

这时候，`sysdig` 就能派上用场了。你可以用它读取并解析之前保存的捕获文件：

```
sysdig -r evillinks.scap
```

在输出的最后部分，可以看到重点内容，比如：

```
2313 11:21:22.782601383 1 ln (23859) > symlinkat
2314 11:21:22.782662611 1 ln (23859) < symlinkat res=0 target=/home/foo linkdirfd=-100(AT_FDCWD) linkpath=/e
```

这段输出里，重要的信息是：

- **系统调用是 `symlinkat`**。
- `man symlinkat` 手册说明，这是 `symlink` 的一种变体。
- **`linkpath` 参数**保存了符号链接（symlink）的目标路径，比如 `/etc/evillink`。

所以，**我们要关注的是 `symlink` 和 `symlinkat` 两种系统调用，并且要检查它们的 `linkpath` 参数**。

因此，可以设计出如下的过滤器：

```
(evt.type=symlink or evt.type=symlinkat) and (
  evt.arg.linkpath startswith /proc/ or
  evt.arg.linkpath startswith /bin/ or
  evt.arg.linkpath startswith /etc/
)
```

这个意思是：

- 如果事件类型是 `symlink` 或 `symlinkat`
- 并且 `linkpath` 参数是以 `/proc/`、`/bin/` 或 `/etc/` 开头的
- 那么就判定为可疑事件

**用 sysdig 验证过滤器**

可以使用下面的命令测试上面的过滤条件：

```
sysdig -r evillinks.scap \
  "(evt.type=symlink or evt.type=symlinkat) and \
   (evt.arg.linkpath startswith /proc/ or \
    evt.arg.linkpath startswith /bin/ or \
    evt.arg.linkpath startswith /etc/)"
```

执行后输出：

```
438 11:21:13.204948767 2 ln (23814) < symlinkat res=-2(ENOENT) target=/home/foo linkdirfd=-100(AT_FDCWD) linkpath=/proc/evillink
1679 11:21:19.420360948 0 ln (23850) < symlinkat res=0 target=/home/foo linkdirfd=-100(AT_FDCWD) linkpath=/bin/evillink
2314 11:21:22.782662611 1 ln (23859) < symlinkat res=0 target=/home/foo linkdirfd=-100(AT_FDCWD) linkpath=/etc/evillink
```

可以看到输出的三条记录，正好对应我们用 `ln` 命令做的三个可疑的符号链接操作，说明这个过滤条件是正确的！

### 用 sysdig 来编写和测试规则的输出信息

在前面我们已经确定了过滤条件（filter），现在需要**设计规则触发时显示的输出内容**。这一步是为了让 Falco 触发警报时，能清楚地告诉我们到底发生了什么。

这里，`sysdig` 也能帮上忙！

比如，为了这条规则，想要输出的信息是这样：

```
a symlink was created in a sensitive directory
(link=%evt.arg.linkpath, target=%evt.arg.target, cmd=%proc.cmdline)
```

可以用下面的命令，在 `sysdig` 里一起测试过滤器和输出：

```
sysdig -r evillinks.scap \
  -p"a symlink was created in a sensitive directory\
  (link=%evt.arg.linkpath, target=%evt.arg.target, cmd=%proc.cmdline)" \
  "(evt.type=symlink or evt.type=symlinkat) and \
  (evt.arg.linkpath startswith /proc/ or \
  evt.arg.linkpath startswith /bin/ or \
  evt.arg.linkpath startswith /etc/)"
```

执行后输出结果是：

```
a symlink was created in a sensitive directory (link=/proc/evillink, target=/home/foo, cmd=ln -s /home/foo /proc/evillink)
a symlink was created in a sensitive directory (link=/bin/evillink, target=/home/foo, cmd=ln -s /home/foo /bin/evillink)
a symlink was created in a sensitive directory (link=/etc/evillink, target=/home/foo, cmd=ln -s /home/foo
```

### 把 sysdig 命令转换成正式的 Falco 规则

前面我们用 `sysdig` 测试了过滤条件（condition）和输出内容（output），现在要把它们正式整理成一个 **Falco 规则**。

Falco 规则通常包括几个部分：

1. **rule**（规则名）
2. **desc**（描述信息）
3. **condition**（匹配条件）
4. **output**（触发时显示的内容）
5. **priority**（优先级，比如 WARNING、ERROR、CRITICAL）

根据之前的工作，把它整理成一个 Falco YAML 规则，大概是这样：

```
- rule: Symlink in a Sensitive Directory
  desc: Detect the creation of a symbolic link in a sensitive directory.
  condition: >
    (evt.type=symlink or evt.type=symlinkat) and (
     evt.arg.linkpath startswith /proc/ or
     evt.arg.linkpath startswith /bin/ or
     evt.arg.linkpath startswith /etc/)
  output: >
    a symlink was created in a sensitive directory
     (link=%evt.arg.linkpath, target=%evt.arg.target, cmd=%proc.cmdline)
  priority: WARNING
```

### 验证 Falco 规则是否生效

前面我们已经把规则写好了，现在要做的就是**在 Falco 里验证**这个规则能否正确工作。

具体步骤如下：

1. 把规则保存成一个 YAML 文件

比如保存成 `symlink.yaml` 文件，内容就是前面写的那个规则。

2. 让 Falco 加载这个规则，并用之前录制的捕获文件 `evillinks.scap` 来测试

执行这个命令：

```
falco -r symlink.yaml -e evillinks.scap
```

3. 运行后，Falco 应该会输出类似下面的日志：

```
Fri Sep 15 15:41:23 2023: Falco version: 0.35.1 (x86_64)
Fri Sep 15 15:41:23 2023: Falco initialized with configuration file: /etc/falco/falco.yaml
Fri Sep 15 15:41:23 2023: Loading rules from file symlink.yaml
Fri Sep 15 15:41:23 2023: Reading system call events from file: evillinks.scap
15:40:05.272773097: Warning a symlink was created in a sensitive directory (link=/proc/evillink, target=/root, cmd=ln -s /root /proc/evillink)
15:40:56.979673067: Warning a symlink was created in a sensitive directory (link=/etc/evillink, target=/root, cmd=ln -s /root /etc/evillink)
15:41:02.021948830: Warning a symlink was created in a sensitive directory (link=/bin/evillink, target=/root, cmd=ln -s /root /bin/evillink)
```

可以看到：

- **规则成功触发了3次**（因为之前在 `/proc`、`/etc`、`/bin` 都各建了一个符号链接）。
- 输出的警告内容正是我们在 `output` 字段里定义的格式。

还有一些统计信息，比如：

- `event drop detected: 0 occurrences`：说明没有丢失事件，数据完整。
- `Events detected: 3`：检测到了 3 个符合条件的事件。
- `Triggered rules by rule name:`：列出哪个规则被触发了几次，比如 `Symlink in a Sensitive Directory: 3`。

### 让规则更模块化、可维护、优化性能

1. 第一阶段：提取公共条件，使用 **macro（宏）**

最初的规则中，`condition` 写得很长很直白，比如：

```
(evt.type=symlink or evt.type=symlinkat) and (
 evt.arg.linkpath startswith /proc/ or
 evt.arg.linkpath startswith /bin/ or
 evt.arg.linkpath startswith /etc/
)
```

为了让条件更清晰、好维护，可以把每一部分抽出来，分别定义成宏：

```
- macro: sensitive_sylink_dir
  condition: >
    (evt.arg.linkpath startswith /proc/ or
     evt.arg.linkpath startswith /bin/ or
     evt.arg.linkpath startswith /etc/)

- macro: create_symlink
  condition: (evt.type=symlink or evt.type=symlinkat)
```

然后，主规则就可以直接用这些宏组合条件了：

```
- rule: Symlink in a Sensitive Directory
  desc: Detect the creation of a symbolic link in a sensitive directory.
  condition: create_symlink and sensitive_sylink_dir
  output: >
    a symlink was created in a sensitive directory
    (link=%evt.arg.linkpath, target=%evt.arg.target, cmd=%proc.cmdline)
  priority: WARNING
```

这样写有什么好处？

- **可读性强**：一眼就能看出规则在做什么。
- **易维护**：如果将来想修改敏感目录，只要改宏就行，不需要到处改。

------

2. 第二阶段：继续优化，使用 **list（列表）**

更进一步，还可以把**固定的值**（比如系统调用名、目录名）提取成单独的列表：

```
- list: symlink_syscalls
  items: [symlink, symlinkat]

- list: sensitive_dirs
  items: [/proc/, /bin/, /etc/]
```

然后让宏使用新的高级操作符：

- `evt.type in (symlink_syscalls)` ：检查事件类型是不是在 symlink 的系统调用里。
- `evt.arg.linkpath pmatch (sensitive_dirs)` ：检查链接的路径是否以 `/proc/`、`/bin/` 或 `/etc/` 开头。

对应的宏变成：

```
- macro: sensitive_sylink_dir
  condition: (evt.arg.linkpath pmatch (sensitive_dirs))

- macro: create_symlink
  condition: evt.type in (symlink_syscalls)
```

规则本身还是一样：

```
- rule: Symlink in a Sensitive Directory
  desc: Detect the creation of a symbolic link in a sensitive directory.
  condition: create_symlink and sensitive_sylink_dir
  output: >
    a symlink was created in a sensitive directory
    (link=%evt.arg.linkpath, target=%evt.arg.target, cmd=%proc.cmdline)
  priority: WARNING
```