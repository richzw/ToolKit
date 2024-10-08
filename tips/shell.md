
- [find 命令的 7 种用法](https://mp.weixin.qq.com/s/XS2KOhBeGeviusIqIBnxKg)
  - 如果你的 Linux 服务器上有一个名为 logs 的目录，如何删除该目录下最后一次访问时间超过一年的日志文件呢 `find . -type f -atime +365 -exec rm -rf {} \`
  - 按名称或正则表达式查找文件 `find ./yang/books -type f -name "*.pdf"`
  - 查找不同类型的文件 - 指定 -type 选项来搜索其他类型的文件 ``
  - 按指定的时间戳查找文件
    - 访问时间戳（atime）：最后一次读取文件的时间。
    - 修改时间戳（mtime）：文件内容最后一次被修改的时间。
    - 更改时间戳（ctime）：上次更改文件元数据的时间（如，所有权、位置、文件类型和权限设置）
    - + 表示“大于”，- 表示“小于”。所以我们可以搜索 ctime 在 5~10 天前的文 `find . -type f -ctime +5 -ctime -10`
  - 按大小查找文件 - 要查找大小为 10 MB ~ 1 GB 的文件 `find . -type f -size +10M -size -1G`
  - 按所有权查找文件 - 以下命令将查找所有属于 yang 的文件 `find -type f -user yang`
  - 在找到文件后执行命令 - `-exec` 选项后面的命令必须以分号（;）结束 `find . -type f -atime +5 -exec ls {} \;`
- [Top 20 Network Monitoring Tools in Linux](https://linoxide.com/network-monitoring-tools-linux/)
  - [Translation](https://mp.weixin.qq.com/s/VwPxTr5tBdteE2aJg1cYUA)
- 进程上次的启动时间
  - `ps -o lstart {pid}`
- [ip rate limit](https://making.pusher.com/per-ip-rate-limiting-with-iptables/index.html)
  ```shell
  $ iptables --new-chain SOCAT-RATE-LIMIT
  $ iptables --append SOCAT-RATE-LIMIT \
      --match hashlimit \
      --hashlimit-mode srcip \
      --hashlimit-upto 50/sec \
      --hashlimit-burst 100 \
      --hashlimit-name conn_rate_limit \
      --jump ACCEPT
  $ iptables --append SOCAT-RATE-LIMIT --jump DROP
  $ iptables -I INPUT -p tcp --dport 1234 --jump SOCAT-RATE-LIMIT
  ```
- [iptablse常用命令](https://mp.weixin.qq.com/s/1RIR_AECgDr45EENtPRK9w)
  - 清除iptables（常用） `iptables -F`
  - 备份iptables（常用） `iptables-save > iptables.txt`
  - 导入iptables（常用） `iptables-restore < iptables.txt`
  - 机器重启自动生效（常用） `service iptables save`
  - 清空某条规则： `iptables -t filter -D INPUT -s 1.2.3.4 -j DROP`
  - 禁止某个ip（下面用$ip表示）访问本机： `iptables -I INPUT -s $ip -j DROP`
  - 禁止某个ip段（下面用 mask表示，其中$mask是掩码）访问本机： `iptables -I INPUT -s $ip/$mask -j DROP`
  - 禁止本机访问某个ip（下面用$ip表示）： `iptables -A OUTPUT -d $ip -j DROP`
  -  禁止某个ip（下面用$ip表示）访问本机的80端口的tcp服务： `iptables -I INPUT -p tcp –dport 80 -s $ip -j DROP`
  - 禁止所有ip访问本机的80端口的tcp服务： `iptables -A INPUT -p tcp --dport 80 -j DROP`
  - 禁止所有ip访问本机的所有端口： `iptables -A INPUT -j DROP`
  - 禁止除了某个ip（下面用$ip表示）之外其他ip都无法访问本机的3306端口（常用）：
    - （1）首先 禁止所有 `iptables -I INPUT -p tcp --dport 3306 -j DROP`
    - （2）然后 开放个别 `iptables -I INPUT -s $ip -p tcp --dport 3306 -j ACCEPT`
- bash
  - 通过增加 -v 选项，即可开启详细模式，用于查看所执行的命令
  - 通过增加 -x 参数来进入 xtrace 模式，用于调试执行阶段的变量值。
  - 增加 -u 选项， 可以检查变量是否未定义/绑定
  - 组合使用 -vu 就可以直接看到具体出现问题的代码是什么内容
  - 在需要调试的位置设置 set -x ，在结束的位置设置 set +x ，这样调试日志中就只会记录我需要调试部分的日志了
  - set -e 选项。该选项在遇到首个 非0 值的时候会直接退出
- [SSH Tunnels](https://iximiuz.com/en/posts/ssh-tunnels/)
  - ![img.png](shell_ssh_tunnel.png)
    ```shell
    remote# nc -lvk 7780
    
    local# ssh -N -L 7777:localhost:7780 root@159.89.238.232
    
    local# echo "foo bar" | nc -N localhost 7777
    ```
- [git commands](https://mp.weixin.qq.com/s/EXHboxE0talIZakIitWp6w)
  - 我的提交信息(commit message)写错了，我想要修改它
    - `git commit --amend --only -m 'xxxxxxx'`
  - 我想从一个提交(commit)里移除一个文件
    - `$ git checkout HEAD^ myfile
      $ git add -A
      $ git commit --amend`
  - 我想删除我的的最后一次提交(commit) 如果你还没有推到远程, 把Git重置(reset)到你最后一次提交前的状态就可以
    - `$ git reset --soft HEAD@{1}`
  - 我意外的做了一次硬重置(hard reset)，我想找回我的内容
    - `$ git reflog`
    - `$ git checkout -b temp <commit-hash>`
    - `$ git checkout master`
    - `$ git merge temp`
    - `$ git branch -d temp`
  - 我想要暂存一个新文件的一部分，而不是这个文件的全部内容
    - `$ git add --patch filename.x`
  - 我想把在一个文件里的变化(changes)加到两个提交(commit)里
    - git add 会把整个文件加入到一个提交. git add -p 允许交互式的选择你想要提交的部分.
  - 我从错误的分支拉取了内容，或把内容拉取到了错误的分支
    - 这是另外一种使用 git reflog 情况，找到在这次错误拉(pull) 之前HEAD的指向。然后把HEAD重置到那个指向
    - `$ git reset --hard <commit-hash>`
  - 我想撤销rebase/merge
    - Git 在进行危险操作的时候会把原始的HEAD保存在一个叫ORIG_HEAD的变量里, 所以要把分支恢复到rebase/merge前的状态是很容易的
    - `$ git reset --hard ORIG_HEAD`
  - 我需要组合(combine)几个提交(commit) 
    - 假设你的工作分支将会做对于 main 的pull-request。一般情况下你不关心提交(commit)的时间戳，只想组合 所有 提交(commit) 到一个单独的里面, 然后重置(reset)重提交(recommit)。确保主(main)分支是最新的和你的变化都已经提交了, 然后
    - `(my-branch)$ git reset --soft main
      (my-branch)$ git commit -am "New awesome feature"`
    - 如果你想要保留你的提交(commit)的时间戳，你可以使用 git rebase -i main
  - 安全合并(merging)策略
    - --no-commit 执行合并(merge)但不自动提交, 给用户在做提交前检查和修改的机会。no-ff 会为特性分支(feature branch)的存在过留下证据, 保持项目历史一致
    - `(main)$ git merge --no-ff --no-commit my-branch`
  - 检查是否分支上的所有提交(commit)都合并(merge)过了
    - `(main)$ git log --graph --left-right --cherry-pick --oneline HEAD...feature/120-on-scroll`
- [Shell Commands]()
  - 查看2015年8月16日14时这一个小时内有多少IP访问: 
    -  `cat access.log | awk '{print $1}' | sort | uniq -c | awk '{print $2}' | grep -E '2015-08-16 14:' | wc -l`
  - 查看访问前十个ip地址 
    - `cat access.log | awk '{print $1}' | sort | uniq -c | sort -nr | head -10 | awk '{print $2}'`
  - 列出当前服务器每一进程运行的数量，倒序排列 
    - `ps aux | awk '{print $11}' | sort | uniq -c | sort -nr`
- [Vim cheatsheet](https://mp.weixin.qq.com/s/BkJnbXvuVZIAExOkgVqPWw)
- [perf](https://mp.weixin.qq.com/s/5Y_ZyDPM6OcejvktyoZzDw)
  - Perf 工作模式
    - Couting Mode 
      - Counting Mode 将会精确统计一段时间内 CPU 相关硬件计数器数值的变化。为了统计用户感兴趣的事件，Perf Tool 将设置性能控制相关的寄存器。这些寄存器的值将在监控周期结束后被读出。典型工具：Perf Stat。
    - Sampling Mode
      - Sampling Mode 将以定期采样方式获取性能数据。PMU 计数器将为某些特定事件配置溢出周期。当计数器溢出时，相关数据，如 IP、通用寄存器、EFLAG 将会被捕捉到。典型工具：Perf Record。
  - Perf Tool 
    - Perf Stat：分析性能。
      ```shell
      perf stat -p $pid -d     # 进程级别统计
      perf stat -a -d sleep 5  # 系统整体统计
      perf stat -p $pid -e 'syscalls:sys_enter' sleep 10  #分析进程调用系统调用的情形
      ```
    - Perf Top：实时性能分析。
      ```shell
      perf top -p $pid -g -F 99
      ```
    - Perf Record：记录性能数据。记录一段时间内系统/进程的性能事件, 默认性能事件为 cycles ( CPU 周期数 )
      ```shell
      perf record -p $pid -g -e cycles -e cs #进程采样
      perf record -a -g -e cycles -e cs #系统整体采样
      ```
- [git 瘦身](https://mp.weixin.qq.com/s/nNxKzniBeygreKLAAX0Mqw)
  - ulimit -n 9999999 # 解决可能出现的报错too many open files的问题
  - git rev-list --objects --all | sort -k 2 > allfileshas.txt # 获取所有文件的hash值
  -  使用 git filter-branch 截断历史记录
  - 使用 git-filter-repo 清理截断日期前的所有历史记录，并将截断节点的提交信息修改
- [常用 Shell 脚本](https://mp.weixin.qq.com/s/hPqiHB4g2IcP0PLFc1UgWQ)
- [分析日志](https://mp.weixin.qq.com/s/5wyuAuG1rPWtVuZ2P2c9mA)
  - 从分隔日志中解析字段 `awk -F, '{print $2, $5}' access.log`


