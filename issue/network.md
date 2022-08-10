
---------
- Case 1 [大量 fin-wait2](https://mp.weixin.qq.com/s?__biz=MjM5MDUwNTQwMQ==&mid=2257486478&idx=1&sn=1f037a765e023d9b321081ae017bc850&chksm=a539e258924e6b4e464122e12ce009c913edb1c671672dbbfce0f22e5d8e81adf33a6cb13c24&cur_album_id=1690026440752168967&scene=190#rd)

  - 分析
      - 分析业务日志发现了大量的接口超时问题，连接的地址跟`netstat`中`fin-wait2`目的地址是一致的
      - 通过`strace`追踪socket的系统调用，发现golang的socket读写超时没有使用setsockopt so_sndtimeo so_revtimeo参数
      ```bash
        [pid 34262] epoll_ctl(3, EPOLL_CTL_ADD, 6, {EPOLLIN|EPOLLOUT|EPOLLRDHUP|EPOLLET, {u32=1310076696, u64=140244877192984}}) = 0
        [pid 34265] epoll_pwait(3,  <unfinished ...>
        [pid 34262] <... getsockname resumed>{sa_family=AF_INET, sin_port=htons(45242), sin_addr=inet_addr("127.0.0.1")}, [112->16]) = 0
        [pid 34264] epoll_pwait(3,  <unfinished ...>
        [pid 34262] setsockopt(6, SOL_TCP, TCP_NODELAY, [1], 4 <unfinished ...>
        [pid 34262] setsockopt(6, SOL_SOCKET, SO_KEEPALIVE, [1], 4 <unfinished ...>
        [pid 34264] read(4,  <unfinished ...>
        [pid 34262] setsockopt(6, SOL_TCP, TCP_KEEPINTVL, [30], 4 <unfinished ...>
      ```
      - 在连接的roundTrip方法里有超时引发关闭连接的逻辑。由于http的语义不支持多路复用，所以为了规避超时后再回来的数据造成混乱，索性直接关闭连接

  - Solution
    - 要么加大客户端的超时时间，要么优化对端的获取数据的逻辑，总之减少超时的触发。
    
---------
- Case 2 [AWS ALB 502](https://adamcrowder.net/posts/node-express-api-and-aws-alb-502/)

  - Details
    
    The `502` Bad Gateway error is caused when the ALB sends a request to a service at the same time that the service closes the connection by sending the `FIN` segment to the ALB socket. The ALB socket receives `FIN`, acknowledges, and starts a new handshake procedure.

    Meanwhile, the socket on the service side has just received a data request referencing the previous (now closed) connection. Because it can’t handle it, it sends an RST segment back to the ALB, and then the ALB returns a 502 to the user.

  - Solution

    Just make sure that the service doesn’t send the `FIN` segment before the ALB sends a `FIN` segment to the service. In other words, make sure the service doesn’t close the HTTP **Keep-Alive connection** before the ALB.

    The default timeout for the AWS Application Load Balancer is 60 seconds, so we changed the service timeouts to 65 seconds. Barring two hiccoughs shortly after deploying, this has totally fixed it.
    
--------
- Case 3 [HOL Blocking](https://mp.weixin.qq.com/s?__biz=Mzg5MTYyNzM3OQ==&mid=2247483985&idx=1&sn=9546ced2f5b9df02537769ff167c9db9&chksm=cfcb304df8bcb95b0527e4ca94ecd66325d8db0b8c7a5eecb92bdd9ed573ce750c1e751809e2&cur_album_id=1899309536088293384&scene=190#rd)

  - tcpdump 定位 root cause 是 tcp retransmit 引起的 HOL(head of line) blocking

--------
- Case 4 [iptables redirect](https://mp.weixin.qq.com/s/fmrw-33cbKLdAkkMHyOrbw)

  - 问题： 由于流量突增临时扩充多个node部署服务，但遇到一个问题全量接口调用失败总是返回无关的返回结果。简单说在服务里本调用其他服务接口，返回的结果莫名其妙。

  - 分析：
    - 本机dig dns解析无异常
    - 通过 lsof 和 netstat 可以看到已建立连接是正常的解析ip，但是对端确实没有收到该请求
    - strace是可以看到服务请求过程中所涉及到的系统调用
    - 尝试使用tcpdump来抓包。每次请求时都会跟127.0.0.1:80建连，请求体也会转到127.0.0.1:80上。这类情况很像是做了端口劫持跳转
    - 在iptables里发现了redirect跳转。所有output请求会转到sidecar_outbound自定义链，在sidecar自定义链中又把目标地址中80的请求转到本地的80端口上。
  - 测试
    - script
      ```shell
      iptables -t nat -N SIDECAR_OUTBOUND
      iptables -t nat -A OUTPUT -p tcp -j SIDECAR_OUTBOUND
      iptables -t nat -A SIDECAR_OUTBOUND -p tcp -d 123.56.0.0 --dport 80 -j REDIRECT --to-port 80
      ```

- UDP优化

  - UDP 存在粘包半包问题？
    - tcp 是无边界的，tcp 是基于流传输的，tcp 报头没有长度这个变量，而 udp 是有边界的，基于消息的，是可以解决粘包问题的。
    - udp 并没有完美的解决应用层粘包半包的问题。如果你的 go udp server 的读缓冲是 1024，那么 client 发送的数据不能超过 server read buf 定义的 1024 byte，不然还是要处理半包了。如果发送的数据小于 1024 byte，倒是不会出现粘包的问题
    - 借助 strace 发现 syscall read fd 的时候，最大只获取 1024 个字节。这个 1024 就是上面配置的读缓冲大小
  - golang udp 的锁竞争
    - udp 压力测试的时候，发现 client 和 server 都跑不满 cpu 的情况。尝试使用 iperf 进行 udp 压测，golang udp server 的压力直接干到了满负载
    - 尝试在 go udp client 里增加了多协程写入，10 个 goroutine，100 个 goroutine，500 个 goroutine，都没有好的明显的提升效果，而且性能抖动很明显
    - 通过 lsof 分析 client 进程的描述符列表，client 连接 udp server 只有一个连接。也就是说，500 个协程共用一个连接
    - 使用 strace 做 syscall 系统调用统计，发现 futex 和 pselect6 系统调用特别多，这一看就是存在过大的锁竞争
  - 优化
    - 实例化多个 udp 连接到一个数组池子里，在客户端代码里随机使用 udp 连接。这样就能减少锁的竞争了。
    - udp 在合理的 size 情况下是不需要依赖应用层协议解析包问题。那么我们只需要在 client 端控制 send 包的大小，server 端控制接收大小，就可以节省应用层协议带来的性能高效

- [`gnet` TCP流协议解析程序](https://mp.weixin.qq.com/s/Hrh63H1f1dmxAL9qt6bJng)

- [Go 语言的网络编程模型](https://mp.weixin.qq.com/s/YT71cQr6vbz4wsF5I3flNA)
  - Go 语言的网络编程模型是同步网络编程。它基于 协程 + I/O 多路复用 （linux 下 epoll，darwin 下 kqueue，windows 下 iocp，通过网络轮询器 netpoller 进行封装），结合网络轮询器与调度器实现。
  - 用户层 goroutine 中的 block socket，实际上是通过 netpoller 模拟出来的。runtime 拦截了底层 socket 系统调用的错误码，并通过 netpoller 和 goroutine 调度让 goroutine 阻塞在用户层得到的 socket fd 上
  - Go 将网络编程的复杂性隐藏于 runtime 中：开发者不用关注 socket 是否是 non-block 的，也不用处理回调，只需在每个连接对应的 goroutine 中以 block I/O 的方式对待 socket 即可。
  - ![img.png](network_netpoll.png)
  - `tcpdump -S -nn -vvv -i lo0 port 8000` -vvv是为了打印更多的详细描述信息，-S 显示序列号绝对值

- [Forcefully Close TCP Connections in Golang](https://itnext.io/forcefully-close-tcp-connections-in-golang-e5f5b1b14ce6)
  - the traditional default close 

    when we execute our `net.Conn.Close()` method, the TCP session we execute it against will start a connection termination sequence which includes handling (discarding) any outstanding data. That is, until we receive the final FIN-ACK packet
    `tcpdump -n -vvv -i lo0 port 9000`
  
  - a forceful close using the `SetLinger()` method

    socket 缓冲区信息可通过执行 `netstat -nt` 命令查看

    This method is changing the SO_LINGER socket option value using system calls against the Operating System.
    - If sec < 0 (the default), the operating system finishes sending the data in the background.
    - If sec == 0, the operating system discards any unsent or unacknowledged data.
    - If sec > 0, the data is sent in the background as with sec < 0. On some operating systems after sec seconds have elapsed any remaining unsent data may be discarded.

    ```go
    // Use SetLinger to force close the connection
    // When set to exactly 0, the Operating System will immediately close the connection and drop any outstanding packets.
    err := c.(*net.TCPConn).SetLinger(0)
    if err != nil {
        log.Printf("Error when setting linger: %s", err)
    }
    defer c.Close()
    ```
    A RST packet is a special type of packet used for “resetting” TCP connections. It is a way for the sender to tell the remote side that it will neither accept nor receive new data for this connection.

- [SO_REUSEPORT](https://douglasmakey.medium.com/socket-sharding-in-linux-example-with-go-b0514d6b5d08)
  
  Linux 3.9 内核引入了 SO_REUSEPORT选项（实际在此之前有一个类似的选项 SO_REUSEADDR，但它没有做到真正的端口复用，详细可见参考链接1）。

  SO_REUSEPORT 支持多个进程或者线程绑定到同一端口，用于提高服务器程序的性能。它的特性包含以下几点：

  - 允许多个套接字 bind 同一个TCP/UDP 端口
    - 每一个线程拥有自己的服务器套接字
    - 在服务器套接字上没有了锁的竞争
  - 内核层面实现负载均衡
  - 安全层面，监听同一个端口的套接字只能位于同一个用户下（same effective UID）
  - For TCP sockets, this option allows accept(2) load distribution in a multi-threaded server to be improved by using a distinct listener socket for each thread. This provides improved load distribution as compared to traditional techniques such using a single accept(2)ing thread that distributes connections, or having multiple threads that compete to accept(2) from the same socket.
  - For UDP sockets, the use of this option can provide better distribution of incoming datagrams to multiple processes (or threads) as compared to the traditional technique of having multiple processes compete to receive datagrams on the same socket.

  ```go
  var lc = net.ListenConfig{
      Control: func(network, address string, c syscall.RawConn) error {
          var opErr error
          if err := c.Control(func(fd uintptr) {
              opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
          }); err != nil {
              return err
          }
          return opErr
      },
  }
  l, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:8080")
  ```
  
  - Security
    
    To prevent this “port hijacking,” Linux has special protections or mechanisms to prevent these problems, such as:
    - Both sockets must have been created with the SO_REUSEPORT socket option. If there is a socket running without SO_REUSEPORT and we try to create another socket even with the SO_REUSEPORT socket option, it will fail with the error already in use.
    - All sockets that want to listen to the same IP and port combination must have the same effective userID. For example, if you want to hijack the Nginx port and it is running under the ownership of the user Pepito, a new process can listen to the same port only if it is also owned by the user Pepito. So one user cannot “steal” ports of other users.
- [TCP粘包](https://mp.weixin.qq.com/s/c3NYmCTf8LbakL5ZS1Fv-g)
  - MTU vs MSS
    - MTU: Maximum Transmit Unit，最大传输单元。 由网络接口层（数据链路层）提供给网络层最大一次传输数据的大小；一般 MTU=1500 Byte
    - MSS：Maximum Segment Size 。TCP 提交给 IP 层最大分段大小，不包含 TCP Header 和  TCP Option，只包含 TCP Payload ，MSS 是 TCP 用来限制应用层最大的发送字节数。
    ![img.png](network_mtu_mss.png)
  - 为什么会出现粘包
    - TCP，Transmission Control Protocol。传输控制协议，是一种面向连接的、可靠的、**基于字节流**的传输层通信协议。
    - 应用层传到 TCP 协议的数据，不是以消息报为单位向目的主机发送，而是以字节流的方式发送到下游，这些数据可能被切割和组装成各种数据包，接收端收到这些数据包后没有正确还原原来的消息，因此出现粘包现象。
  - 为什么要组装发送的数据
    - TCP的 Nagle 算法优化，目的是为了避免发送小的数据包。
    - 在 Nagle 算法开启的状态下，数据包在以下两个情况会被发送：
      - 如果包长度达到MSS（或含有Fin包），立刻发送，否则等待下一个包到来；如果下一包到来后两个包的总长度超过MSS的话，就会进行拆分发送；
      - 等待超时（一般为200ms），第一个包没到MSS长度，但是又迟迟等不到第二个包的到来，则立即发送。
  - 关掉Nagle算法就不会粘包了吗？
    - `TCP_NODELAY = 1`
    - 就算关闭 Nagle 算法，接收数据端的应用层没有及时读取 TCP Recv Buffer 中的数据，还是会发生粘包。
  - 怎么处理粘包
    - 只要在发送端每次发送消息的时候给消息带上识别消息边界的信息，接收端就可以根据这些信息识别出消息的边界，从而区分出每个消息
      - 加入特殊标志
      - 加入消息长度信息
  - UDP 会粘包吗
    - 在报头中有16bit用于指示 UDP 数据报文的长度，假设这个长度是 n ，以此作为数据边界。因此在接收端的应用层能清晰地将不同的数据报文区分开，从报头开始取 n 位，就是一个完整的数据报，从而避免粘包和拆包的问题
  - 为什么长度字段冗余还要加到 UDP 首部中
    - IP 层是网络层的，而 UDP 是传输层的，到了传输层，数据包就已经不存在IP头信息了，那么此时的UDP数据会被放在 UDP 的  Socket Buffer 中。当应用层来不及取这个 UDP 数据报，那么两个数据报在数据层面其实都是一堆 01 串。
  - IP 层有粘包问题吗
    - 先说结论，不会。首先前文提到了，粘包其实是由于使用者无法正确区分消息边界导致的一个问题。
    - IP 层从按长度切片到把切片组装成一个数据包的过程中，都只管运输，都不需要在意消息的边界和内容，都不在意消息内容了，那就不会有粘包一说了。
- [localhost 就一定是 localhost 么?](https://mp.weixin.qq.com/s/x0798dbodAxdyUIGYfEBjA)
  - Issue
    - 我们在本地测试或者本地通讯的时候经常使用 localhost 域名，但是访问 localhost 的对应的一定就是我们的本机地址么？
    - 我们明明是配置的 localhost，为什么会出现这个地址？localhost 不应该指向的是 127.0.0.1 么？我们使用 dig 和 nslookup 之后发现 localhost 的确是 127.0.0.1
    - 我们在机器上抓包之后发现 localhost 竟然走了域名解析! 并且 localhost 这个域名在我们内网还被注册了，解析出来的地址就是最开始发现的这个不知名的地址
    - 我们下意识认为的域名解析流程应该是这样的，先去找 /etc/hosts 文件，localhost 找到了（默认是 127.0.0.1）就返回了
    - 排查之后发现，实际上的流程是这样的，先做了 DNS 查询 DNS 没查到然后去查了 /etc/hosts 文件
    - 直到有一天，我们的内网域名解析中添加了一个 localhost 的域名解析，就直接查询成功返回了
  - 复现
    ```go
    func main() {
     client := &http.Client{}
     _, err := client.Get("http://localhost:8080")
     fmt.Println(err)
    }
    
    # GODEBUG="netdns=go+2" go run main.go 
    go package net: GODEBUG setting forcing use of Go's resolver
    go package net: hostLookupOrder(localhost) = files,dns
    Get "http://localhost:8080": dial tcp [::1]:8080: connect: connection refused
    上面显示的 files,dns 的意思就是先从 /etc/hosts 文件中查询，再去查询 dns 结果
    ```
    Docker模拟
    ```shell
    FROM golang:1.15 as builder
    
    WORKDIR /app
    
    COPY main.go main.go
    COPY run.sh run.sh
    
    ENV CGO_ENABLED=0
    ENV GOOS=linux
    
    RUN go build main.go
    
    FROM alpine:3
    
    WORKDIR /app
    
    COPY --from=builder /app /app
    COPY run.sh run.sh
    
    RUN chmod +x run.sh
    
    ENV GODEBUG="netdns=go+2"
    ENV CGO_ENABLED=0
    ENV GOOS=linux
    
    CMD /app/run.sh
    ```
    使用这个容器运行的结果如下，可以看到已经变成了 dns,files 为什么会这样呢？
  - 排查
    - src/net/dnsclient_unix.go Go 中定义了下面几种 DNS 解析顺序，其中 files 表示查询 /etc/hosts 文件，dns 表示执行 dns 查询
    - Go 会先根据一些初始条件判断查询的顺序，然后就查找 /etc/nsswitch.conf 文件中的 hosts 配置项，如果不存在就会走一些回退逻辑。这次的问题出现在这个回退逻辑上
    - 当前系统如果是 linux 并且不存在 /etc/nsswitch.conf 文件的时候，会直接返回 dns,files 的顺序
- [KUBERNETES/DOCKER网络排障](https://coolshell.cn/articles/18654.html)
  - 问题
    - 某个pod被重启了几百次甚至上千次。
    - 用 docker exec -it 命令直接到容器内启了一个 Python的 SimpleHttpServer来测试发现也是一样的问题
  - 排查
    - 用 telnet ip port 的命令手工测试网络连接时有很大的概率出现 connection refused 错误，大约 1/4的概率，而3/4的情况下是可以正常连接的
    - 抓个包看看，然后，用户抓到了有问题的TCP连接是收到了 SYN 后，立即返回了 RST, ACK (docker0 返回 RST ACK)
    - 在 telnet 上会显示 connection refused 的错误信息，对于我个人的经验，这种 SYN完直接返回 RST, ACK的情况只会有三种情况
      - TCP链接不能建立，不能建立连接的原因基本上是标识一条TCP链接的那五元组不能完成，绝大多数情况都是服务端没有相关的端口号。
      - TCP链接建错误，有可能是因为修改了一些TCP参数，尤其是那些默认是关闭的参数，因为这些参数会导致TCP协议不完整。
      - 有防火墙iptables的设置，其中有 REJECT 规则。
    - 有点像 NAT 的网络中服务端开启了 tcp_tw_recycle 和 tcp_tw_reuse 的症况 - 查看了一上TCP参数，发现用户一个TCP的参数都没有改，全是默认的，于是我们排除了TCP参数的问题
    - 也不觉得容器内还会设置上iptables，而且如果有那就是100%的问题，不会时好时坏。
    - 抓包这个事，在 docker0 上可以抓到，然而到了容器内抓不到容器返回 RST, ACK
    - 于是这个事把我们逼到了最后一种情况 —— IP地址冲突了
    - 我们发现用户的机器上有 arping 于是我们用这个命令来检测有没有冲突的IP地址。 -D 参数是检测IP地址冲突模式，如果这个命令的退状态是 0 那么就有冲突。结果返回了 1
      ```shell
      $ arping -D -I docker0 -c 2 10.233.14.145
      $ echo $?
      ```
    - 想看看所有的 network namespace 下的 veth 网卡上的IP
      - 首先，我们到 /var/run/netns目录下查看系统的network namespace，发现什么也没有。
      - 然后，我们到 /var/run/docker/netns 目录下查看Docker的namespace，发现有好些。
      - 于是，我们用指定位置的方式查看Docker的network namespace里的IP地址
        ```shell
        $ nsenter --net=/var/run/docker/netns/421bdb2accf1 ifconfig -a
        $ ls /var/run/docker/netns | xargs -I {} nsenter --net=/var/run/docker/netns/{} ip addr 
        $ lsns -t net | awk '{print $4}' | xargs -t -I {} nsenter -t {}&nbsp;-n ip addr | grep -C 4 "10.233.14.137"
        ```
  - Docker
    - Docker 1.11版以后，Docker进程组模型就改成上面这个样子了.
      - dockerd 是 Docker Engine守护进程，直接面向操作用户。dockerd 启动时会启动 containerd 子进程，他们之前通过RPC进行通信。
      - containerd 是dockerd和runc之间的一个中间交流组件。他与 dockerd 的解耦是为了让Docker变得更为的中立，而支持OCI 的标准 。
      - containerd-shim  是用来真正运行的容器的，每启动一个容器都会起一个新的shim进程， 它主要通过指定的三个参数：容器id，boundle目录（containerd的对应某个容器生成的目录，一般位于：/var/run/docker/libcontainerd/containerID）， 和运行命令（默认为 runc）来创建一个容器。
      - docker-proxy 你有可能还会在新版本的Docker中见到这个进程，这个进程是用户级的代理路由。只要你用 ps -elf 这样的命令把其命令行打出来，你就可以看到其就是做端口映射的。如果你不想要这个代理的话，你可以在 dockerd 启动命令行参数上加上：  --userland-proxy=false 这个参数。
- [线上一次大量 CLOSE_WAIT 复盘](https://ms2008.github.io/2019/07/04/golang-redis-deadlock/)
  - 出现 CLOSE_WAIT 本质上是因为服务端收到客户端的 FIN 后，仅仅回复了 ACK（由系统的 TCP 协议栈自动发出），并没有发 4 次断开的第二轮 FIN（由应用主动调用 Close() 或 Shutdown() 发出）
  - `ss -ta sport = :9000` 客户端关闭连接后，CLOSE_WAIT 依然不会消失，只能说明服务端 HANG 在了某处，没有调用 close
  - 打印出服务的调用栈信息, 出现了将近 140W 的 goroutine，有将近 100W 都 block 在了 redis 连接的获取上。顺手确认 redis 的连接情况：ss -tn dport = :6379 | sed 1d | wc -l 发现 redis 连接池已经占满。
  - redigo 初始化连接池的时候如果没有传入 timeout，那么在执行命令时将永远不会超时
    ```go
    connTimeout := redis.DialConnectTimeout(time.Duration(10) * time.Second)
    readTimeout := redis.DialReadTimeout(time.Duration(10) * time.Second)
    writeTimeout := redis.DialWriteTimeout(time.Duration(10) * time.Second)
    
    redisPool = &redis.Pool{
        MaxIdle:     conf.MaxIdle,
        MaxActive:   conf.MaxActive,
        Wait:        true,
        IdleTimeout: 240 * time.Second,
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", conf.Addr, connTimeout, readTimeout, writeTimeout)
            if err != nil {
                return nil, err
            }
            return c, err
        },
        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            if time.Since(t) < time.Minute {
                return nil
            }
            _, err := c.Do("PING")
            return err
        },
    }
    ```
  - redigo 也提供了一个更安全的获取连接的接口：GetContext()，通过显式传入一个 context 来控制 Get() 的超时：
- [CLOSE WAIT 分析](https://gocn.vip/topics/yQ0XrZH0Qd)
  - 背景
    - telnet这个ip port端口是通的，说明四层以下网络是通的。
    - 但是当我们curl到对应服务的时候，发现服务被连接重置 RST
  - 分析问题机器
    - 查看网络问题
      - `netstat -tunp | grep CLOSE_WAIT | wc -l`
      - 我们发现CLOSE WAIT的客户端的端口号，在服务端机器上有，但在客户端机器上早就消失。CLOSE WAIT就死在了这台机器上。
      - RECVQ这么大，为什么服务还能正常响应200
      - 为什么上面curl会超时，但是telnet可以成功
      - 为什么CLOSE WAIT没有pid
    - 线上调试
      - 我们使用了`sudo strace -p pid`，发现主进程卡住了，想当然的认为自己发现了问题。但实际过程中，这是正常现象，我们框架使用了 golang.org/x/net/netutil 里的 LimitListener ，当你的goroutine达到最大值，那么就会出现这个阻塞现象，因为没有可用的goroutine了，主进程就卡住了。
      - 查看全部线程trace指令为`sudo strace -f -p pid` ，当然也可以查看单个线程的 `ps -T -p pid` ，然后拿到spid，在执行 `sudo strace -p spid`
    - 线下模拟
      - 我们将服务端代码的 `server.Serve(netutil.LimitListener(listener, 1))` 里面的限制设置为1。然后客户端代码的http长连接开启，先做http请求，然后再做tcp请求。
      - `netstat -ntlpa | grep 8080 |grep LISTEN`
      - `netstat -ntlpa | grep 8080 | grep CLOSE | wc -l`
    - 分析
      - 当我们在服务端设置了limit为1的时候，意味这我们服务端的最大连接数只能为1。我们客户端在使用http keepalive，会将这个连接占满，如果这个时候又进行了tcp探活，那么这个探活的请求就会被堵到backlog里，也就是上面我们看到3.3.2中第一个图，里面的RECVQ为513。
      - [CLOSE WAIT不消失的情况](https://blog.cloudflare.com/this-is-strictly-a-violation-of-the-tcp-specification/)
  - 产生线上问题的可能原因
    - 线上的nginx到后端go配置的keepalive，当GO的HTTP连接数达到系统默认1024值，那么就会出现Goroutine无法让出，这个时候使用TCP的探活，将会堵在队列里，服务端出现CLOSE WAIT情况，然后由于一直没有释放goroutine，导致服务端无法发出fin包，一直保持CLOSE WAIT。
    - 而客户端的端口在此期间可能会被重用，一旦重用后，就造成了混乱。（如果在混乱后，Goroutine恢复后，服务端过了好久响应了fin包，客户端被重用的端口收到这个包是返回RST，还是丢弃？这个太不好验证）。个人猜测是丢弃了，导致服务端的CLOSE WAIT一直无法关闭，造成RECVQ的一直阻塞。
  - 其他问题
    - GO服务端的HTTP Keepalive是使用的客户端的Keepalive的时间，如果客户端的Keepalive存在问题，比如客户端的http keepalive泄露，也会导致服务端无法关闭Keepalive，从而无法回收goroutine，当然go前面挡了一层nginx，所以应该不会有这种泄露问题。但保险起见，go的服务端应该加一个keepalive的最大值。例如120s，避免这种问题。
    - GO服务端的HTTP chucked编码时候，如果客户端没有正确将response的body内容取走，会导致数据仍然在服务端的缓冲区，从而导致无法响应fin包，但这个理论上不会出现，并且客户端会自动的进行rst，不会影响业务。
    - 如过写的http服务存在业务问题，例如里面有个死循环，无法响应客户端，也会导致http服务出现CLOSE WAIT问题，但这个是有pid号的
    - 如果我们的业务调用某个服务的时候，由于没发心跳包给服务，会被服务关闭，但我们这个时候没正确处理这个场景，那么我们的业务处就会出现CLOSE WAIT。
- [线上大量CLOSE_WAIT原因排查](https://cloud.tencent.com/developer/article/1381359)
  - 现象
    - 告警出现了 504
  - 发现问题
    - `netstat -na | awk '/^tcp/ {++S[$NF]} END {for(a in S) print a, S[a]}'`发现绝大部份的链接处于 CLOSE_WAIT 状态
    - 为什么会出现大量的mysql连接是  CLOSE_WAIT
       - 确实没有调用close
       - 有耗时操作（火焰图可以非常明显看到），导致超时了
       - mysql的事务没有正确处理，例如：rollback 或者 commit
    - 使用 perf 把所有的调用关系使用火焰图给绘制出来。既然我们推断代码中没有释放mysql连接
    - 由于那一行代码没有对事务进行回滚，导致服务端没有主动发起close。因此 MySQL负载均衡器 在达到 60s 的时候主动触发了close操作，但是通过tcp抓包发现，服务端并没有进行回应，这是因为代码中的事务没有处理，因此从而导致大量的端口、连接资源被占用。
- [Keeping TCP Connections Alive in Golang](https://madflojo.medium.com/keeping-tcp-connections-alive-in-golang-801a78b7cf1)
  - With Go any net.TCPConn type can have keepalives enabled by running the net.TCPConn.SetKeepAlive() method with true as the value.
  - With Go, the frequency of keepalives can be changed from the default using the net.TCPConn.SetKeepAlivePeriod() method.
    - The above example sets the TCP_KEEPINTVL socket option to 30 seconds; overriding the tcp_keepalive_intvl system parameter for this specific connection
    - In addition to overriding the keepalive interval, this method also changes the idle time, setting the TCP_KEEPIDLE socket option to 30 seconds; overriding the tcp_keepalive_time system parameter.
- [记一次线上超时问题的发现、排查、定位以及解决过程](https://mp.weixin.qq.com/s/Yh2dOC8Pumeis-h1_gMLjg)
  - 根据引擎监控系统来查看三方的基础数据监控, 超时监控报警
  - 看到超时这么多，第一时间先ping下，看看网络间耗时咋样
  - 开始尝试使用curl来分析各个阶段的耗时 `curl  -o /dev/null -s -w %{time_namelookup}::%{time_connect}::%{time_starttransfer}::%{time_total}::%{speed_download}"\n" --data-binary @req.dat https://www.baidu.com`
    - 从上述结构可以看出，在time_starttransfe阶段，也就是说对方业务处理结果仍然会出现2s耗时，问题复现。
  - 代码： log是最直接也是最有效的方式
  - 线上抓包： 上面抓包结果来看，在序号为6的位置，对方返回了http 204，然后又重新手动发了一次curl请求，在序号10的时候，对方又返回了http 204(从发送请求到收到对方响应耗时38ms)，但是奇怪的时候，在后面50s后，client端(发送方)开发发送FIN请求关闭连接，从代码逻辑来分析，是因为超时导致的关闭连接。
  - 同类对比： 只能通过与其它家正常返回的做对比，看看能不能得到有用结论。
    - 通过对比上述两家的区别，发现超时的该家较正常的三方返回，多了Content-Length、Content-Type等字段
  - 分析原因
    - http 204 Content-Length hang
    - 从上节标准可以看出，在http 204、304的时候，不允许返回Content-Length，那么如果返回了，libcurl又是如何处理的呢 - libcurl bug
    
- [TCP丢包不重传问题](https://yesphet.github.io/posts/tcp%E4%B8%A2%E5%8C%85%E4%B8%8D%E9%87%8D%E4%BC%A0%E9%97%AE%E9%A2%98%E6%8E%92%E6%9F%A5/)
  - Issue
    - 从以下客户端抓包可以看到，客户端没有收到 [2711665,2719857] 这个包，因此一直在对 2711665进行ack，而服务端确实有收到大量的Duplicate ACK ack 2711665
  - Debug
    - 查看 /proc/sys/net/ipv4/tcp_reordering 的值为3，所以可以确认服务端是有开启快速重传的，在收到3次duplicate ACK后就应该进行快重传，但从抓包文件中并没有发现有触发快速重传
    - 再通过 nstat 发现TcpExtTcpWqueueTooBig值较为异常
    - [Adventures in the TCP stack: Uncovering performance regressions in the TCP SACKs vulnerability fixes](https://www.databricks.com/blog/2019/09/16/adventures-in-the-tcp-stack-performance-regressions-vulnerability-fixes.html)
  - Solve
    - 通过升级内核版本至 3.10.0-1062.1.1.el7.x86_64 后发现问题得到了解决
















