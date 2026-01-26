- [How Terminal work](https://how-terminals-work.vercel.app/)
  - terminal emulator 只是“送原始字符到 pty”；是否触发信号取决于内核 pty/tty 层的标志位与前台进程组设置（tcsetpgrp 等）
  - “EOF 不是信号（not a signal）” EOF 更像是输入流语义，在 TTY 上表现为特定模式下的 read() 行为，而非内核异步信号
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
  - https://eieio.games/blog/ssh-sends-100-packets-per-keystroke/
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
- [snitch](https://github.com/karol-broda/snitch)
  - a friendlier ss / netstat for humans. inspect network connections with a clean tui or styled tables.
- [30个Linux命令组合](https://mp.weixin.qq.com/s/4IIzTv6vA-m7KHpEfNqEtQ)
  - 系统监控与资源排查类（6 个）：快速定位资源瓶颈
    - `top -b -n 1 | grep Cpu | awk '{print "CPU使用率："$2"%"}' && free -h | grep Mem | awk '{print "内存使用率："$3"/"$2"("$7"空闲)"}' && df -h | grep /$ | awk '{print "根分区使用率："$5"("$4"空闲)"}'`
      - 刚登录服务器，想快速看 CPU、内存、根分区的核心状态
    - 快速找出 “吃 CPU 最多的前 10 个进程”，看是不是异常进程
      - `ps -eo pid,ppid,%cpu,%mem,cmd --sort=-%cpu | head -10`
    - 收到 “根分区使用率 90%” 的告警，要快速找到 “大于 100M 的大文件”，看是不是日志没切割或临时文件没清理
      - `find / -type f -size +100M 2>/dev/null | xargs du -sh | sort -hr`
    - 系统响应变慢，但 CPU、内存使用率都不高，怀疑是 “磁盘 IO 瓶颈”，要查看 IO 等待情况
      - `vmstat 1 5 | awk 'NR>1 {print "平均等待IO时间："$16"ms 等待IO进程数："$17"个"}'`
    - 服务器带宽跑满，要查看 “指定网卡（如 eth0）的实时流量”，判断是接收还是发送流量过高
      - `sar -n DEV 1 3 | grep -v Average | awk '/eth0/ {print "网卡eth0：接收"$5"KB/s 发送"$6"KB/s"}'`
    - 要查看 “当前在线的用户”，包括用户名、登录 IP 和时间，判断是否有异常账号。
      - `w | awk 'NR>1 {print "登录用户："$1" 登录IP："$3" 登录时间："$4}'`
  - 日志分析与数据提取类（6 个）：从海量日志中抓关键信息
    - Nginx 服务报 500 错误，要统计 “9 月 8 日当天的 error 日志条数”，判断错误是偶尔出现还是持续爆发。
      - `grep -i "error" /var/log/nginx/error.log | grep -E "2025-09-08" | wc -l`
    - 监控显示 “Nginx 500 错误率上升”，要找出 “出现 500 错误最多的前 10 个 IP 和 URL”，判断是单个 IP 异常访问还是特定接口有问题
      - `grep "500" /var/log/nginx/access.log | awk '{print $1,$7,$9}' | sort | uniq -c | sort -nr | head -10`
    - 想 “实时监控 SSH 登录情况”，一旦有成功登录或失败尝试，立即在终端显示，方便及时发现暴力破解。
      - `tail -f /var/log/messages | grep --line-buffered "ssh" | awk '/Accepted/ {print "正常登录："$0} /Failed/ {print "登录失败："$0}'`
    - 开发反馈 “9 月 8 日 14:00-14:30 之间 Tomcat 报异常”，要提取 “这个时间段内的所有 Exception 日志”，快速定位代码问题
      - `sed -n '/2025-09-08 14:00:00/,/2025-09-08 14:30:00/p' /var/log/tomcat/catalina.out | grep "Exception"`
    - 应用日志用 “|” 分隔字段（如 “时间 | 用户 ID | 接口名 | 耗时”），要统计 “调用次数最多的前 5 个接口”，分析业务热点
      - `awk -F '|' '{print $3}' /var/log/app/log.txt | sort | uniq -c | sort -nr | head -5`
    - 要分析 “昨天压缩后的 Nginx 日志”（后缀.gz）中 “timeout 错误的次数”，不用先解压再 grep（节省磁盘空间）。
      - `zgrep "timeout" /var/log/nginx/access.log-20250907.gz | wc -l`
  - 文件管理与批量操作类（6 个）：告别重复手动操作
    - /data/backup 目录下有每天的备份文件（如 backup_20250901.tar.gz），要 “删除 7 天前的旧备份”，避免占满磁盘
      - `find /data/backup -name "*.tar.gz" -mtime +7 -exec rm -f {} \;`
    - 每天凌晨要 “给 /data/logs 目录下的所有.log 文件加日期后缀”（如 access.log→access.log.20250908），方便后续归档。
      - `for file in /data/logs/*.log; do mv "$file" "$file.$(date +%Y%m%d)"; done`
    - 批量修改 /etc/config 目录下所有.conf 文件中的 IP”（把 192.168.1.10 改成 192.168.1.20）
      - `sed -i 's/old_ip=192.168.1.10/old_ip=192.168.1.20/g' /etc/config/*.conf`
    - 备份 /data/app 目录时，要 “排除 logs 子目录”（日志占空间且无需备份），并给备份包加日期后缀，方便识别。
      - `tar -zcvf /data/backup/app_$(date +%Y%m%d).tar.gz /data/app --exclude=/data/app/logs`
    - /data/blacklist.txt 是 “黑名单 IP 列表”（每行一个 IP），/data/user.txt 是 “用户访问记录”（格式 “时间 IP 用户名”），要 “提取不在黑名单中的用户记录”，过滤异常 IP
      - `awk 'NR==FNR{a[$1];next} !($2 in a)' /data/blacklist.txt /data/user.txt`
  - 进程与服务管理类（4 个）：快速管控服务状态
    - 快速查看 Nginx 服务状态”（是运行中、停止还是失败），并看 “最近 10 分钟的服务日志”，判断服务是否正常。
      - `systemctl status nginx | grep -E "active|inactive|failed" && journalctl -u nginx --since "10 minutes ago" | tail -20`
    - 强制杀死所有 Java 进程”，之后重启服务。
      - `ps -ef | grep java | grep -v grep | awk '{print $2}' | xargs kill -9`
    - 要 “后台启动 /data/app/start.sh 脚本”，并把输出日志写到 /data/logs/app.log，且退出终端后进程不停止
      - `nohup /data/app/start.sh > /data/logs/app.log 2>&1 &`
    - 写定时任务（crontab）时，要 “检查 app.jar 进程是否存在，不存在则启动”，实现服务的简单保活
      - `pgrep -f "app.jar" || /data/app/start.sh`
    - 服务器访问 10.0.0.1（总部网关）超时，要 “跟踪网络链路”，看是哪一跳（路由器）出了问题，延迟过高还是丢包
      - `traceroute 10.0.0.1 | grep -E "^\s*[0-9]+" | awk '{print "跳数："$1" IP："$2" 延迟："$3}'`
    - Nginx 服务（80 端口）的连接数突然飙升，要 “找出连接 80 端口最多的前 10 个 IP”，判断是否有 IP 在恶意发起大量连接（如 CC 攻击）
      - `ss -antp | grep :80 | awk '{print $5}' | cut -d: -f1 | sort | uniq -c | sort -nr | head -10`
  - 权限与安全审计类（4 个）：保障服务器安全
    - 服务器被扫描出 “有文件权限为 777（所有人可读写执行）”，要 “找出 /data/app 目录下所有权限为 777 的文件”，修改为安全权限（如 644）
      - `find /data/app -perm 777 -type f 2>/dev/null`
    - 查看所有用户的最后登录时间 lastlog | grep -v "Never logged in"
    - 服务器上的文件突然丢失，要 “查看 root 用户最近执行的 20 条含 rm -rf、chmod、chown 的命令”，判断是否有人误操作删除文件或修改权限。
      - grep -E "rm -rf|chmod|chown" /root/.bash_history | tail -20
- SSH
  - [史上最全 SSH 暗黑技巧详解](https://plantegg.github.io/2019/06/02/%E5%8F%B2%E4%B8%8A%E6%9C%80%E5%85%A8_SSH_%E6%9A%97%E9%BB%91%E6%8A%80%E5%B7%A7%E8%AF%A6%E8%A7%A3--%E6%94%B6%E8%97%8F%E4%BF%9D%E5%B9%B3%E5%AE%89/)
  - 把远端机器的一个 port 映射到本地的一个 port，比如我在开发 NebulaGraph Catalyst 的时候，会把集群中的一堆服务跑在服务器，映射到本地，主进程在本地开发调试，轻量、方便（local）
  - 把本地端口映射到远端一个端口，比如调试需要公网上的 api hook 的接口时候可用，这里要注意，需要 sshd 开相关的配置允许
  - 把本地端口作为一个 socks5 代理，网络从 sshd 服务端走（dynamic），这是我常用的非全局连公司网络、家庭网落的方法
  - 除了隧道，ssh 还支持 chain，写在配置里非常简单，加一行 ProxyJump foo 的意思是这个 host 通过配置里的另一个 host foo 的隧道连接
  - ssh -w 可以创建tun，自己写脚本就可以建VPN了 proxycommand 非常强大，曾经用这个走squid 穿防火墙
    - authorized_keys的key前面是可以加配置的，比如限定from 的IP，给不给pty
- [grep组合](https://mp.weixin.qq.com/s/ZrGsGvccwMYPS9fvSJXFCQ)
  - 捕获完整异常堆栈（事后分析）
    - grep -A50 "NullPointerException" application.log | less
    - 通过 -A N 输出匹配行后的 N 行，尽量覆盖堆栈深度；用 less 进行翻页与二次搜索。
  - 实时监控异常 + 保留上下文（线上实时）
    - tail -f application.log | grep -Ai30 "ERROR\|Exception"
    - 将 tail -f 的实时输出交给 grep，并试图用 -A30 保留后续上下文，同时用 -i 忽略大小写，\| 组合多关键词。
  - 不解压直接搜压缩历史日志
    - zgrep -H -A50 "OutOfMemoryError" *.gz
    - 通过 zgrep 直接在 .gz 中搜索，并用 -H 显示文件名便于溯源。
  - 异常趋势统计（频次/排序）
    - grep -c "ConnectionTimeout" *.log | sort -nr -t: -k2
    - 用 grep -c 得到“每个文件匹配行数”，再按计数排序，用于发现异常在哪些文件/时间段更集中。
  - 进阶参数与反向过滤
    - “上下文矩阵”（图示）强调 -A/-B/-C 控制上下文范围
    - grep -v "健康检查\|心跳" app.log | grep -A30 "异常"
  - 生产实战进阶：关联分析/性能优化/正则选择
    - grep -C10 "2023-11-15 14:30" app.log | grep -A20 "事务回滚"（按时间窗口找关联）
    - grep -A40 "traceId:0a1b2c3d" service*.log（按 traceId 串联分布式链路）
    - grep -m1000 "ERROR" large.log（限制匹配数量，避免刷屏/加快定位）
    - grep --binary-files=text "异常" binary.log（把二进制当文本处理）
    - grep -E "Timeout\|Reject\|Failure"（扩展正则）
    - fgrep -f patterns.txt app.log（固定字符串匹配提升性能）
  - 扩展工具链：wc/awk/sed 与组合题
    - grep "ERROR" app.log | wc -l（计数）
    - grep "Exception" app.log | awk -F':' '{print $4}' | sort | uniq | wc -l（统计“异常类型数量”的一种写法）
    - awk '$9 != 200 {print $1}' access.log | sort | uniq -c | sort -nr | head -20（Nginx 非 200 的来源 IP TopN）
    - sed -n '/2023-11-15 14:00:00/,/2023-11-15 14:10:00/p' app.log（按时间段截取）
    - sed 's/\([0-9]\{3\}\)[0-9]\{4\}\([0-9]\{4\}\)/\1****\2/g' app.log（手机号脱敏）
    - “每分钟超时数”的组合（grep+sed+sort+uniq）
- [Linux Perf 性能分析工具](https://mp.weixin.qq.com/s/fdVAydAa9OL-dT0eqnovYg)
  - 基于事件采样的原理，以性能事件为基础，支持针对处理器相关性能指标与操作系统相关性能指标的性能剖析。可用于性能瓶颈的查找与热点代码的定位。
  - Perf 工作模式
    - Counting Mode
      - Counting Mode 将会精确统计一段时间内 CPU 相关硬件计数器数值的变化。
      - 为了统计用户感兴趣的事件，Perf Tool 将设置性能控制相关的寄存器。这些寄存器的值将在监控周期结束后被读出。典型工具：Perf Stat。
      - Perf Stat：分析性能。
    - Sampling Mode
      - Sampling Mode 将以定期采样方式获取性能数据。
      - PMU 计数器将为某些特定事件配置溢出周期。当计数器溢出时，相关数据，如 IP、通用寄存器、EFLAG 将会被捕捉到。典型工具：Perf Record。
      - Perf Record：记录一段时间内系统/进程的性能事件, 默认性能事件为 cycles ( CPU 周期数 )
- [tmux](https://x.com/tz_2022/article/2012624953053503541)
  -  tmux new -A -s main (这行命令最骚，如果有叫 main 的房间就进去，没有就新建，极其无脑)
  - 
































