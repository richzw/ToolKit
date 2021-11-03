
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

- Go 语言的网络编程模型
  - Go 语言的网络编程模型是同步网络编程。它基于 协程 + I/O 多路复用 （linux 下 epoll，darwin 下 kqueue，windows 下 iocp，通过网络轮询器 netpoller 进行封装），结合网络轮询器与调度器实现。
  - 用户层 goroutine 中的 block socket，实际上是通过 netpoller 模拟出来的。runtime 拦截了底层 socket 系统调用的错误码，并通过 netpoller 和 goroutine 调度让 goroutine 阻塞在用户层得到的 socket fd 上
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



