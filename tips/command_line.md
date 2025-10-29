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
    
    







