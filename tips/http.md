
- [RSA 握手的过程](https://mp.weixin.qq.com/s?__biz=MzUxODAzNDg4NQ==&mid=2247487650&idx=1&sn=dfee83f6773a589c775ccd6f40491289&scene=21#wechat_redirect)
- [ECDHE 算法](https://www.cnblogs.com/xiaolincoding/p/14318338.html)
- [网络知识汇总](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247563566&idx=1&sn=26156d79dffb3f0f10b6a26931f993cc&chksm=c1850e7ff6f28769b6ff3358366e917d3d54fc0f0563131422da4bed201768c958262b5d5a99&scene=21#wechat_redirect)
- [C10K到C10M高性能网络的探索与实践](https://mp.weixin.qq.com/s/Jap26lIutqpYSwbFfhb8eQ)
  - 优化
    - IO模型的优化 
      - epoll、kqueue、iocp，[io_ring](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247562787&idx=1&sn=471a0956249ca789afad774978522717&chksm=c1850172f6f28864474f9832bfc61f723b5f54e174417d570a6b1e3f9f04bda7b539662c0bed&scene=21#wechat_redirect) 就是IO模型优化的一些最佳实践
      - 以epoll为例，在它的基础上抽象了一些开发框架和库. libevent、libev
    - CPU亲和性&内存局域性
      - 当前x86服务器以NUMA架构为主，这种平台架构下，每个CPU有属于自己的内存，如果当前CPU需要的数据需要到另外一颗CPU管理的内存获取，必然增加一些延时。
      - Linux提供了sched_set_affinity函数，我们可以在代码中，将我们的任务绑定在指定的CPU核心上。
      - 一些Linux发行版也在用户态中提供了numactl和taskset工具，通过它们也很容易让我们的程序运行在指定的节点上。
    - [RSS、RPS、RFS、XPS](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247561718&idx=1&sn=f93ad69bff3ab80665e4b9d67265e6bd&chksm=c18506a7f6f28fb12341c3e439f998d09c4b1d93f8bf59af6b1c6f4427cea0c48b51244a3e53&scene=21#wechat_redirect)
      - RSS需要硬件的支持，目前主流的网卡都已支持，即俗称的多队列网卡，充分利用多个CPU核心，让数据处理的压力分布到多个CPU核心上去。
      - RPS和RFS在linux2.6.35的版本被加入，一般是成对使用的，在不支持RSS特性的网卡上，用软件来模拟类似的功能，并且将相同的数据流绑定到指定的核心上，尽可能提升网络方面处理的性能。
      - XPS特性在linux2.6.38的版本中被加入，主要针对多队列网卡在发送数据时的优化，当你发送数据包时，可以根据CPU MAP来选择对应的网卡队列，低于指定的kernel版本可能无法使用相关的特性，但是发行版已经backport这些特性。
    - IRQ 优化
      - 中断合并. 一次中断触发后，接下来用轮循的方式读取后续的数据包，以降低中断产生的数量，进而也提升了处理的效率
      - IRQ亲和性. 将不同的网卡队列中断处理绑定到指定的CPU核心上去，适用于拥有RSS特性的网卡。
    - 网络卸载的优化
      - TSO，以太网MTU一般为1500，减掉TCP/IP的包头，TCP的MaxSegment Size为1460，通常情况下协议栈会对超过1460的TCP Payload进行分段，保证最后生成的IP包不超过MTU的大小，对于支持TSO/GSO的网卡来说，协议栈就不再需要这样了，可以将更大的TCPPayload发送给网卡驱动，然后由网卡进行封包操作。通过这个手段，将需要在CPU上的计算offload到网卡上，进一步提升整体的性能
      - GSO为TSO的升级版，不在局限于TCP协议。
      - LRO和TSO的工作路径正好相反，在频繁收到小包时，每次一个小包都要向协议栈传递，对多个TCPPayload包进行合并，然后再传递给协议栈，以此来提升协议栈处理的效率。
      - GRO为LRO的升级版本，解决了LRO存在的一些问题。
    - Kernel 优化
      - 内核网络参数的调整在以下两处：`net.ipv4.*`参数和`net.core.*`参数
    - cache 优化
      - 对于在一条网络连接上的数据处理，尽可能保持在一个CPU核心上，以此换取CPUCache的利用率
      - 无锁数据结构，其是更多的都是编程上的一些技巧。
      - 保证你的数据结构尽可能在相同的CPU核心上进行处理，对于一个数据包相关的数据结构或者哈希表，在相同的类型实例上都保存一份，虽然增加了一些内存占用，但降低了资源冲突的概率
    - 内存优化
      - 尽可能的不要使用多级的指针嵌套
        - 多级的指针检索，有很大的机率触发你的CacheMiss。对于网络数据处理，尽可能将你的数据结构以及数据业务层面尽可能抽象更加扁平化一些，这是一个取舍的问题，也是在高性能面前寻求一个平衡点
      - Hugepage主要解决的问题就是TLB Miss的问题
        - 我们的内存几十G，上百G都已经是常态了。这种不对称的存在，造成了大内存在4k页面的时候产生了大量的TLB Miss
      - 内存预分配的问题
        - 对内存进行更精细化的管理，避免在内存分配上引入一些性能损失或容量损失
  - 探索和实践
    - 内核协议栈问题
      - 全局的队列，在我们在写用户态网络程序中，对同一个网络端口，仅允许一个监听实例，接收的数据包由一个队列来维护，并发的短连接请求较大时，会对这个队列造成较大的竞争压力，成为一个很大瓶颈点
        - 3.9的版本合并了一个很关键的特性SO_REUSEPORT，支持多个进程或线程监听相同的端口，每个实例分配一个独立的队列，一定程度上缓解这个问题。用更容易理解的角度来描述，就是支持了我们在用户态上对一个网络端口，可以有多个进程或线程去监听它。正是因为有这样一个特性，我们可以根据CPU的核心数量来进行端口监听实例的选择，进一步优化网络连接处理的性能。
      - 在linux kernel中有一个全局的连接表，用于维护TCP连接状态，这个表在维护大量的TCP连接时，会造成相当严重的资源竞争。总的来说，有锁的地方，有资源占用的地方都可能会成为瓶颈点。
    - 内核协议栈优化 - 网络数据包处理
      - 在收到数据包之后不进协议栈，把数据包的内存直接映射到用户态，让我们的程序在用户态直接可以看到这些数据。这样就绕过了kernel的处理
      - 利用了linuxUIO，这个特性叫UIO，比如当前流行的DPDK，通过这个模块框架我们可以在驱动程序收到数据包之后，直接放到用户态的内存空间中，也同样达到了绕过协议栈的目的。
- [Http协议各版本的对比](https://mp.weixin.qq.com/s/1TWisgy0wdN2dFMFSRZMUg)
  - http 1.1 - 增加长连接, 管道化, host字段: Host字段用来指定服务器的域名，这样就可以将多种请求发往同一台服务器上的不同网站，提高了机器的复用
  - http 2.0 - 服务端推送, 头部压缩, 多路复用, 二进制格式
    - 二进制分帧层binary framing layer
      - 二进制编码机制使得通信可以在单个TCP连接上进行
      - 二进制协议将通信数据分解为更小的帧，数据帧充斥在C/S之间的双向数据流中
      - 链接Link: 就是指一条C/S之间的TCP链接，这是个基础的链路数据的高速公路
      - 数据流Stream: 已建立的TCP连接内的双向字节流，TCP链接中可以承载一条或多条消息
      - 消息Message: 消息属于一个数据流，消息就是逻辑请求或响应消息对应的完整的一系列帧，也就是帧组成了消息
      - 帧Frame: 帧是通信的最小单位，每个帧都包含帧头和消息体，标识出当前帧所属的数据流
    - 多路复用
      - 由于2.0版本中使用新的二进制分帧协议突破了1.0的诸多限制，从根本上实现了真正的请求和响应多路复用
    - 首部压缩: HPACK算法
    - 服务端推送
    - HTTP2.0虽然性能已经不错了，还有什么不足吗？
      - 建立连接时间长(本质上是TCP的问题)
      - 队头阻塞问题
      - 移动互联网领域表现不佳(弱网环境
  - HTTP3.0又称为HTTP Over QUIC基于UDP协议的QUIC协
    - QUIC 
      - 队头阻塞问题可能存在于HTTP层和TCP层，在HTTP1.x时两个层次都存在该问题。
      - HTTP2.0协议的多路复用机制解决了HTTP层的队头阻塞问题，但是在TCP层仍然存在队头阻塞问题
      - TCP协议如果其中某个包丢失了，就必须等待重传，从而出现某个丢包数据阻塞整个连接的数据使用。
    - 0RTT 建链
      - RTT包括三部分：往返传播时延、网络设备内排队时延、应用程序数据处理时延。
      - 一般来说HTTPS协议要建立完整链接包括:TCP握手和TLS握手，总计需要至少2-3个RTT，普通的HTTP协议也需要至少1个RTT才可以完成握手。
      - QUIC协议可以实现在第一个包就可以包含有效的应用数据，从而实现0RTT0RTT也是需要条件的，对于第一次交互的客户端和服务端0RTT也是做不到的，毕竟双方完全陌生。
      - 连接迁移
        - QUIC协议基于UDP实现摒弃了五元组的概念，使用64位的随机数作为连接的ID，并使用该ID表示连接
- [体验 http3: 基于 nginx quic 分支](https://mp.weixin.qq.com/s/ynqqlSOIOSyp6USynEtrEg)
- [HTTP服务优雅关闭中出现的小插曲](https://mp.weixin.qq.com/s/HY82uS2eYzt7cHqvdHKF-Q)
  - ListenAndServe 总是返回一个非空错误，在调用Shutdown或者Close方法后，返回的错误是ErrServerClosed。
    ```go
    go func() {
      if err := h.server.ListenAndServe(); nil != err {
       if err == http.ErrServerClosed {
        h.logger.Infof("[HTTPServer] http server has been close, cause:[%v]", err)
       }else {
        h.logger.Fatalf("[HTTPServer] http server start fail, cause:[%v]", err)
       }
      }
    }()
    ```
- [QUIC 是如何解决TCP性能瓶颈的](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247541229&idx=1&sn=1179980ab7817614ebd57a741b2326d2&chksm=c184d6bcf6f35faa7d4682f7e69c55e5bec603b3f2d2fa0dd3f0ce932019f7b3d336a7ddb6a4&scene=178&cur_album_id=1843108380194750466#rd)
  - ![img.png](http_http2_quic.png)
  - 一、QUIC 如何解决TCP的队头阻塞问题？
    ![img.png](http_https_http2_quic.png)
    - 1.1 TCP 为何会有队头阻塞问题
      - TCP 采用ACK确认和超时重传机制来保证数据包的可靠交付
      - 逐个发送数据包，等待确认应答到来后再发送下一个数据包，效率太低了，TCP 采用滑动窗口机制来提高数据传输效率
      - TCP 因为超时确认或丢包引起的滑动窗口阻塞问题
      - HTTP/2 在应用协议层通过多路复用解决了队头阻塞问题(使用Stream ID 来标识当前数据流属于哪个资源请求，这同时也是数据包多路复用传输到接收端后能正常组装的依据)，但TCP 在传输层依然存在队头阻塞问题，这是TCP 协议的一个主要性能瓶颈
    - 1.2 QUIC 如何解决队头阻塞问题
      - TCP 队头阻塞的主要原因是数据包超时确认或丢失阻塞了当前窗口向右滑动，我们最容易想到的解决队头阻塞的方案是不让超时确认或丢失的数据包将当前窗口阻塞在原地
      - QUIC 同样是一个可靠的协议，它使用 Packet Number 代替了 TCP 的 Sequence Number，并且每个 Packet Number 都严格递增，也就是说就算 Packet N 丢失了，重传的 Packet N 的 Packet Number 已经不是 N，而是一个比 N 大的值，比如Packet N+
      - QUIC 支持乱序确认，当数据包Packet N 丢失后，只要有新的已接收数据包确认，当前窗口就会继续向右滑动
      - 重传的数据包Packet N+M 和丢失的数据包Packet N 单靠Stream ID 的比对一致仍然不能判断两个数据包内容一致，还需要再新增一个字段Stream Offset，标识当前数据包在当前Stream ID 中的字节偏移量。有了Stream Offset 字段信息，属于同一个Stream ID 的数据包也可以乱序传输了
      - ![img.png](http_quic_packet_offset.png)
      - QUIC 通过单向递增的Packet Number，配合Stream ID 与 Offset 字段信息，可以支持非连续确认应答Ack而不影响数据包的正确组装，摆脱了TCP 必须按顺序确认应答Ack 的限制（也即不能出现非连续的空位），解决了TCP 因某个数据包重传而阻塞后续所有待发送数据包的问题（也即队头阻塞问题）
    - 1.3 QUIC 没有队头阻塞的多路复用
      - ![img.png](http_quic_frame.png)
      - 同一个Connection ID 可以同时传输多个Stream ID，由于QUIC 支持非连续的Packet Number 确认，某个Packet N 超时确认或丢失，不会影响其它未包含在该数据包中的Stream Frame 的正常传输。
      - HTTP/2 中如果TCP 出现丢包，TLS 也会因接收到的数据不完整而无法对其进行处理，也即HTTP/2 中的TLS 协议层也存在队头阻塞问题，该问题如何解决呢？
      - 既然TLS 协议是因为接收数据不完整引起的阻塞，我们只需要让TLS 加密认证过程基于一个独立的Packet，不对多个Packet 同时进行加密认证，就能解决TLS 协议层出现的队头阻塞问题
  - 二、QUIC 如何优化TCP 的连接管理机制？
    - 2.1 TCP连接的本质是什么
      - TCP 连接主要是双方记录并同步维护的状态组成的。一般来说，建立连接是为了维护前后分组数据的承继关系，维护前后承继关系最常用的方法就是对其进行状态记录和管理。
      - TCP 的状态管理可以分为连接状态管理和分组数据状态管理两种，连接状态管理用于双方同步数据发送与接收状态，分组数据状态管理用于保证数据的可靠传输。
    - 2.2 QUIC 如何减少TCP 建立连接的开销
      - ![img.png](http_tfo.png)
      - TCP 建立连接需要三次握手过程，第三次握手报文发出后不需要等待应答回复就可以发送数据报文了，所以TCP 建立连接的开销为 1-RTT
      - TLS 简短握手过程是将之前完整握手过程协商的信息记录下来，以Session Ticket 的形式传输给客户端，如果想恢复之前的会话连接，可以将Session Ticket 发送给服务器，就能通过简短的握手过程重建或者恢复之前的连接，通过复用之前的握手信息可以节省 1-RTT 的连接建立开销
      - TCP 也提供了快速建立连接的方案 TFO (TCP Fast Open)，原理跟TLS 类似，也是将首次建立连接的状态信息记录下来，以Cookie 的形式传输给客户端，如果想复用之前的连接，可以将Cookie 发送给服务器，如果服务器通过验证就能快速恢复之前的连接，TFO 技术可以通过复用之前的连接将连接建立开销缩短为 0-RTT
      - QUIC 可以理解为”TCP + TLS 1.3“（QUIC 是基于UDP的，可能使用的是DTLS 1.3），QUIC 自然也实现了首次建立连接的开销为 1-RTT，快速恢复先前连接的开销为 0-RTT 的效率。
    - 2.3 QUIC 如何实现连接的无感迁移
      - 每个网络连接都应该有一个唯一的标识，用来辨识并区分特定的连接。TCP 连接使用<Source IP, Source Port, Target IP, Target Port> 这四个信息共同标识
      - QUIC 数据包结构中有一个Connection ID 字段专门标识连接，Connection ID 是一个64位的通用唯一标识UUID (Universally Unique IDentifier)。借助Connection ID，QUIC 的连接不再绑定IP 与 Port 信息，即便因为网络迁移或切换导致Source IP 和Source Port 发生变化，只要Connection ID 不变就仍是同一个连接
  - 三、QUIC 如何改进TCP 的拥塞控制机制？
    - 3.1 TCP 拥塞控制机制的瓶颈在哪？
      - ![img.png](http_cwnd.png)
      - 一个是目前接收窗口的大小，通过接收端的实际接收能力来控制发送速率的机制称为流量控制机制；
      - 另一个是目前拥塞窗口的大小，通过慢启动和拥塞避免算法来控制发送速率的机制称为拥塞控制机制，TCP 发送窗口大小被限制为不超过接收窗口和拥塞窗口的较小值。
    - 3.2 QUIC 如何降低重传概率
      - QUIC 采用单向递增的Packet Number 来标识数据包，原始请求的数据包与重传请求的数据包编号并不一样，自然也就不会引起重传的歧义性，采样RTT 的测量更准确
      - ![img.png](http_quic_rtt.png)
      - QUIC 计算RTT 时除去了接收端的应答延迟时间，更准确的反映了网络往返时间，进一步提高了RTT 测量的准确性，降低了数据包超时重传的概率
      - QUIC 引入了前向冗余纠错码（FEC: Fowrard Error Correcting），如果接收端出现少量（不超过FEC的纠错能力）的丢包或错包，可以借助冗余纠错码恢复丢失或损坏的数据包，这就不需要再重传该数据包了，降低了丢包重传概率，自然就减少了拥塞控制机制的触发次数，可以维持较高的网络利用效率。
    - 3.3 QUIC 如何改进拥塞控制机制
      - TCP 的拥塞控制实际上包含了四个算法：慢启动、拥塞避免、快速重传、快速恢复. 由于TCP 内置于操作系统，拥塞控制算法的更新速度太过缓慢，跟不上网络环境改善速度，TCP 落后的拥塞控制算法自然会降低网络利用效率
      - QUIC 协议当前默认使用了 TCP 的 Cubic 拥塞控制算法，同时也支持 CubicBytes、Reno、RenoBytes、[BBR](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247540557&idx=1&sn=958c13feee4aa384cdac8435eff55c5e&chksm=c184a81cf6f3210a7c3f309d054592e99a2845e9cf93f58aeefa58001304c7c0268283da684d&scene=21#wechat_redirect)、PCC 等拥塞控制算法
      - QUIC 是处于应用层的，可以随浏览器更新，QUIC 的拥塞控制算法就可以有较快的迭代速度，在TCP 的拥塞控制算法基础上快速迭代，可以跟上网络环境改善的速度，尽快提高拥塞恢复的效率。













