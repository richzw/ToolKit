
- HPA
  ![img.png](k8s_hpa.png)
- [VPC与三种K8s网络模型](https://mp.weixin.qq.com/s/W04uff4sHrPM_VtjzHPoJA)
  - 物理网卡和虚拟网卡数据接收完整数据处理流程
    - ksoftirqd扛起了网络包处理的重担。它从pull_list里找到要处理的网卡并从该网卡的queue里面拿到skb，然后沿着网络设备子系统 -> IP协议层 -> TCP层一路调用内核里面的函数来分析和处理这个skb，最终将其放置到位于TCP层的socket接收队列里。
    ![img.png](k8s_ksoftirqd.png)
    [Source](https://github.com/LanceHBZhang/LanceAndCloudnative/blob/master/%E9%AB%98%E6%B8%85%E5%A4%A7%E5%9B%BE/%E7%89%A9%E7%90%86%E7%BD%91%E5%8D%A1%E5%92%8C%E8%99%9A%E6%8B%9F%E7%BD%91%E5%8D%A1%E6%95%B0%E6%8D%AE%E6%8E%A5%E6%94%B6%E5%AE%8C%E6%95%B4%E6%95%B0%E6%8D%AE%E5%A4%84%E7%90%86%E6%B5%81%E7%A8%8B.png)
  - K8s“扁平网络”的三种典型实现方式是：Overlay、主机间路由（host-gw）以及Underlay
    - VPC and Overlay
      - K8s Overlay网络模型的实现有一个重要的技术点：veth pair。我们暂时将位于容器内的veth叫veth1，而插在bridge端口的那个对端veth称为veth-peer。当网络包从veth1流出时，其实是在图1的2.a处把skb所属的设备修改为veth-peer，先暂时放到了input_pkt_queue里，但这个时候skb还没有到设备veth-peer处
      - 在每台VM上，K8s CNI都会创建一个bridge和vtep。veth pair的一端安装在了Pod内部，而另一端则插到了网桥上
      - VTEP（VXLAN Tunnel Endpoints）是VXLAN网络中绝对的主角，它既可以是物理设备，也可以是虚拟化设备，主要负责 VXLAN 协议报文的封包和解包。
      - ![img.png](k8s_network_overlay.png)
    - VPC and UnderLay
      - 网络包从Pod流出后，像VM的eth0一样，直接进入了Open vSwith上
      - ![img.png](k8s_network_underlay.png)
    - VPC与host-gw
      - Host-gw简单来讲就是将每个Node当成Pod的网关。所谓网关就像城门，是网络包离开当前局部区域的卡口。
      - [VPC与host-gw（Flannel](https://github.com/LanceHBZhang/LanceAndCloudnative/blob/master/%E9%AB%98%E6%B8%85%E5%A4%A7%E5%9B%BE/vpc%E5%92%8CK8s%20host-gw%E7%BD%91%E7%BB%9C%E6%A8%A1%E5%9E%8B%EF%BC%88Flannel%E5%AE%9E%E7%8E%B0%E6%96%B9%E6%A1%88%EF%BC%89.png)
        - Flannel的实现方案里，由bridge来将离开Pod的网络包丢进宿主机TCP/IP协议栈进行路由的查询。最终网络包经由宿主机的eth0离开并进入对方宿主机的eth。当然这个过程中离不开OVS基于VXLAN所架设的隧道。
      - [VPC与host-gw（Calico](https://github.com/LanceHBZhang/LanceAndCloudnative/blob/master/%E9%AB%98%E6%B8%85%E5%A4%A7%E5%9B%BE/vpc%E5%92%8CK8s%20host-gw%E7%BD%91%E7%BB%9C%E6%A8%A1%E5%9E%8B%EF%BC%88Flannel%E5%AE%9E%E7%8E%B0%E6%96%B9%E6%A1%88%EF%BC%89.png)
        - BGP Client用于在集群里分发路由规则信息，而Felix则负责更新宿主机的路由表。
- [What Happen when K8S run](https://github.com/jamiehannaford/what-happens-when-k8s)
- [Demystifying kube-proxy](https://mayankshah.dev/blog/demystifying-kube-proxy/)
- [Container Networking Is Simple](https://iximiuz.com/en/posts/container-networking-is-simple/)
- [Life of a Packet in Kubernetes](https://dramasamy.medium.com/life-of-a-packet-in-kubernetes-part-1-f9bc0909e051)
- [如何调试Kubernetes集群中的网络延迟问题](https://mp.weixin.qq.com/s/78yVKmNP-huNAiHH7K_3aw)
  - [Kubernetes 平台上的服务零星延迟问题](https://github.blog/2019-11-21-debugging-network-stalls-on-kubernetes/)
  - 刚开始归结于网络链路抖动，一段时间后依然存在，虽然影响都是 P99.99 以后的数据，但是扰人心智，最后通过多方面定位，解决了该问题。
  - 通过排查，我们将问题缩小到与 Kubernetes 节点建立连接的这个环节，包括集群内部的请求或者是涉及到外部的资源和外部的访问者的请求。
  - 最简单的重现这个问题的方法是：在任意的内部节点使用 Vegeta 对一个以 NodePort 暴露的服务发起 HTTP 压测，我们就能观察到不时会产生一些高延迟请求
  - 拨开迷雾找到问题的关键
    - ![img.png](k8s_issue1.png)
    - Vegeta 客户端会向集群中的某个 Kube 节点发起 TCP 请求。在我们的数据中心的 Kubernetes 集群使用 Overlay 网络（运行在我们已有的数据中心网络之上），会把 Overlay 网络的 IP 包封装在数据中心的 IP 包内。当请求抵达第一个 kube 节点，它会进行 NAT 转换，从而把 kube 节点的 IP 和端口转换成 Overlay 的网络地址，具体来说就是运行着应用的 Pod 的 IP 和端口。在请求响应的时候，则会发生相应的逆变换（SNAT/DNAT）
    - 在最开始利用 Vegeta 进行进行压测的时候，我们发现在 TCP 握手的阶段（SYN 和 SYN-ACK 之间）存在延迟。为了简化 HTTP 和 Vegeta 带来的复杂度，我们使用 hping3 来发送 SYN 包，并观测响应的包是否存在延迟的情况，
      `$ sudo hping3 172.16.47.27 -S -p 30927 -i u10000 | egrep --line-buffered 'rtt=[0-9]{3}\.'`
    - 根据日志中的序列号以及时间，我们首先观察到的是这种延迟并不是单次偶发的，而是经常聚集出现，就好像把积压的请求最后一次性处理完似的。
    - 我们想要具体定位到是哪个组件有可能发生了异常。是 kube-proxy 的 NAT 规则吗，毕竟它们有几百行之多？还是 IPIP 隧道或类似的网络组件的性能比较差？排查的一种方式是去测试系统中的每一个步骤。如果我们把 NAT 规则和防火墙逻辑删除，仅仅使用 IPIP 隧道会发生什么？你同样也在一个 kube 节点上，那么 Linux 允许你直接和 Pod 进行通讯，非常简单
      `$ sudo hping3 10.125.20.64 -S -i u10000 | egrep --line-buffered 'rtt=[0-9]{3}\.'`
    - 从我们的结果看到，问题还是在那里！这排除了 iptables 以及 NAT 的问题。那是不是 TCP 出了问题？我们来看下如果我们用 ICMP 请求会发生什么。
      `$ sudo hping3 10.125.20.64 --icmp -i u10000 | egrep --line-buffered 'rtt=[0-9]{3}\.'`
    - 结果显示 ICMP 仍然能够复现问题。那是不是 IPIP 隧道导致了问题？让我们来进一步简化问题。那么有没有可能这些节点之间任意的通讯都会带来这个问题？
      `$ sudo hping3 172.16.47.27 --icmp -i u10000 | egrep --line-buffered 'rtt=[0-9]{3}\.'`
    - 在这个复杂性的背后，简单来说其实就是两个 kube 节点之间的任何网络通讯，包括 ICMP。如果这个目标节点是“异常的”.
    - 这次我们从 kube 节点发送请求到外部节点. 通过查看抓包中的延迟数据, 我们获得了更多的信息。具体来说，从发送端观察到了延迟，然而接收端的服务器没有看到延迟
    - 通过查看接收端的 TCP 以及 ICMP 网络包的顺序的区别（基于序列 ID）， 我们发现 ICMP 包总是按照他们发送的顺序抵达接收端，但是送达时间不规律，而 TCP 包的序列 ID 有时会交错，其中的一部分会停顿。尤其是，如果你去数 SYN 包发送/接收的端口，这些端口在接收端并不是顺序的，而他们在发送端是有序的。
    - 目前我们服务器所使用的网卡，比如我们在自己的数据中心里面使用的那些硬件，在处理 TCP 和 ICMP 网络报文时有一些微妙的区别。当一个数据报抵达的时候，网卡会对每个连接上传递的报文进行哈希，并且试图将不同的连接分配给不同的接收队列，并为每个队列（大概）分配一个 CPU 核心。对于 TCP 报文来说，这个哈希值同时包含了源 IP、端口和目标 IP、端口。换而言之，每个连接的哈希值都很有可能是不同的。对于 ICMP 包，哈希值仅包含源 IP 和目标 IP，因为没有端口之说。这也就解释了上面的那个发现。
    - 另一个新的发现是一段时间内两台主机之间的 ICMP 包都发现了停顿，然而在同一段时间内 TCP 包却没有问题。这似乎在告诉我们，是接收的网卡队列的哈希在“开玩笑”，我们几乎确定停顿是发生在接收端处理 RX 包的过程中，而不是发送端的问题。
  - 深入挖掘 Linux 内核的网络包处理过程
    - 在最简单原始的实现中，网卡接收到一个网络包以后会向 Linux 内核发送一个中断，告知有一个网络包需要被处理。内核会停下它当前正在进行的其他工作，将上下文切换到中断处理器，处理网络报文然后再切换回到之前的工作任务。
    - Linux 新增了一个 NAPI，Networking API 用于代替过去的传统方式，现代的网卡驱动使用这个新的 API 可以显著提升高速率下包处理的性能。在低速率下，内核仍然按照如前所述的方式从网卡接受中断。一旦有超过阈值的包抵达，内核便会禁用中断，然后开始轮询网卡，通过批处理的方式来抓取网络包。这个过程是在“softirq”中完成的，或者也可以称为软件中断上下文（software interrupt context）。这发生在系统调用的最后阶段，此时程序运行已经进入到内核空间，而不是在用户空间。
    - 这种方式比传统的方式快得多，但也会带来另一个问题。如果包的数量特别大，以至于我们将所有的 CPU 时间花费在处理从网卡中收到的包，但这样我们就无法让用户态的程序去实际处理这些处于队列中的网络请求（比如从 TCP 连接中获取数据等）。最终，队列会堆满，我们会开始丢弃包。
    - 为了权衡用户态和内核态运行的时间，内核会限制给定软件中断上下文处理包的数量，安排一个“预算”。一旦超过这个"预算"值，它会唤醒另一个线程，称为“ksoftiqrd”（或者你会在 ps 命令中看到过这个线程），它会在正常的系统调用路径之外继续处理这些软件中断上下文。这个线程会使用标准的进程调度器，从而能够实现公平的调度。
    - 通过整理 Linux 内核处理网络包的路径，我们发现这个处理过程确实有可能发生停顿。如果 softirq 处理调用之间的间隔变长，那么网络包就有可能处于网卡的 RX 队列中一段时间。这有可能是由于 CPU 核心死锁或是有一些处理较慢的任务阻塞了内核去处理 softirqs。
  - 将问题缩小到某个核心或者方法
    - 这些 ICMP 包会被散列到某一个特定的网卡 RX 队列，然后被某个 CPU 核心处理。如果我们想要理解内核正在做什么，那么我们首先要知道到底是哪一个 CPU 核心以及 softirq 和 ksoftiqrd 是如何处理这些包的，这对我们定位问题会十分有帮助。
    - 我们知道内核正在处理那些 IMCP 的 Ping 包，那么我们就来拦截一下内核的 icmp_echo 方法，这个方法会接受一个入站方向的 ICMP 的“echo 请求”包，并发起一个 ICMP 的回复“echo response”。我们可以通过 hping3 中显示的 icmp_seq 序列号来识别这些包。
    - 结果告诉我们一些事情。首先，这些数据包由 ksoftirqd/11 进程处理的，它很方便地告诉我们这对特定的机器将其 ICMP 数据包散列到接收方的 CPU 核心 11 上。我们还可以看到，每次看到停顿时，我们总是会看到在 cadvisor 的系统调用 softirq 上下文中处理了一些数据包，然后 ksoftirqd 接管并处理了积压，而这恰好就对应于我们发现的那些停顿的数据包。
  - cAdvisor 做了什么会导致停顿
    - 我们使用 cAdvisor 正是为了“分析正在运行的容器的资源使用情况和性能特征”，但它却引发了这一性能问题。
    - 为了让内核能够硬阻塞而不是提前调度 ksoftirqd，并且我们也看到了在 cAdvisor 的 softirq 上下文中处理的数据包，我们认为 cAdvisor 调用 syscall 可能非常慢，而在它完成之后其余的网络包才能够被正常处理
    ![img.png](k8s_cadvisor.png)
    - perf record 工具能以特定频率对指定的 CPU 内核进行采样，并且可以生成实时的调用图
      ```shell
      # record 999 times a second, or every 1ms with some offset so not to align exactly with timers
      sudo perf record -C 11 -g -F 999
      # take that recording and make a simpler stack trace.
      sudo perf script 2>/dev/null | ./FlameGraph/stackcollapse-perf-ordered.pl | grep ksoftir -B 100
      ```
    - 我们可以使用 strace 来查看 cAdvisor 到底在做什么，并找到那些超过 100ms 的系统调用。
      `sudo strace -p 10137 -T -ff 2>&1 | egrep '<0\.[1-9]'`
    - 我们非常确信 read()系统调用是很慢的。从 read 读取的内容和 mem_cgroup 这个上下文来看，那些 read()调用是在读取 memory.state 文件，这些文件用于描述系统的内存使用以及 cgroup 的限制。cAdvisor 通过轮询这个文件来获取容器所使用的资源的详情。
  - 是什么导致这个读取如此缓慢
    - 它主要是关于内存的 cgroup，它负责管理与统计命名空间（容器）内的内存使用情况。当该 cgroup 中的所有进程退出时，内存 cgroup 会被 Docker 释放。但是，“内存”不仅是进程的内存，而且虽然进程内存的使用量已经消失，但事实证明，内核还为缓存空间分配了内存，例如 dentries 和 inode（目录和文件元数据），这些内容被缓存到内存 cgroup 中。
    - “僵尸”cgroups：那些没有进程运行并被删除的 cgroups 仍然持有一定的内存空间（在我们的案例中，这些缓存对象是目录数据，但也有可能是页缓存或是 tmpfs）。
    - 与其在 cgroup 释放的时候遍历所有的缓存页，而这也可能很慢，内核会惰性地等待这些内存需要用到的时候再去回收它们，当所有的内存页被清理以后，相应的 cgroup 才会最后被回收。与此同时，这些 cgroup 仍然会被计入统计信息中。
    - 我们的节点具有大量的僵尸 cgroup，有些节点的读/停顿超过一秒钟












