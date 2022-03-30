
- [ptmalloc,tcmalloc和jemalloc内存分配策略](https://cloud.tencent.com/developer/article/1173720)
  - 操作系统内存布局
    ![img.png](os_32bit_memor.png)
    32位默认内存布局
    ![img_1.png](os_64bit_memory.png)
    64位内存布局
  - Ptmalloc
    - Ptmalloc采用主-从分配区的模式，当一个线程需要分配资源的时候，从链表中找到一个没加锁的分配区，在进行内存分配。
    - 小内存分配
      - [获取分配区(arena)并加锁] -> fast bin -> unsorted bin -> small bin -> large bin -> top chunk -> 扩展堆
    - 大内存分配
      - top chunk -> mmap
    - 注意事项：
      - Ptmalloc默认后分配内存先释放，因为内存回收是从top chunk开始的。
      - 避免多线程频繁分配和释放内存，会造成频繁加解锁。
      - 不要分配长生命周期的内存块，容易造成内碎片，影响内存回收。
    - Ptmalloc在性能上还是存在一些问题
      - 不同分配区（arena）的内存不能交替使用
      - 每个内存块分配都要浪费8字节内存
  - Tcmalloc
    - 内存管理分为线程内存和中央堆两部分
    - 小内存分配
      - 在线程缓存内的60个分配器（_文档上说60个，但是我在2.0的代码里看到得是86个_）分别维护了一个大小固定的自由空间链表，直接由这些链表分配内存的时候是不加锁的。但是中央堆是所有线程共享的，在由其分配内存的时候会加自旋锁(spin lock)。
    - 大内存分配
      - 对于大内存分配(大于8个分页, 即32K)，tcmalloc直接在中央堆里分配。中央堆的内存管理是以分页为单位的，同样按大小维护了256个空闲空间链表，前255个分别是1个分页、2个分页到255个分页的空闲空间，最后一个是更多分页的小的空间。这里的空间如果不够用，就会直接从系统申请了。
    - 资源释放
      - 首先计算其分页编号，然后再查找出它对应的span，如果它是一个小对象，则直接归入小对象分配器的空闲链表。等到空闲空间足够大以后划入中央堆。如果是大对象，则会把物理地址连续的前后的span也找出来，如果空闲则合并，并归入中央堆中。
  - Jemalloc
    - Jemalloc 把内存分配分为了三个部分
      - 第一部分类似tcmalloc，是分别以8字节、16字节、64字节等分隔开的small class；
      - 第二部分以分页为单位，等差间隔开的large class；
      - 然后就是huge class。
      - 内存块的管理也通过一种chunk进行，一个chunk的大小是2^k (默认4 MB)。通过这种分配实现常数时间地分配small和large对象，对数时间地查询huge对象的meta（使用红黑树）。
    - 小内存（small class）： 线程缓存bin -> 分配区bin(bin加锁) -> 问系统要
    - 中型内存（large class）：分配区bin(bin加锁) -> 问系统要
    - 大内存（huge class）： 直接mmap组织成N个chunk+全局huge红黑树维护(带缓存)
  - Jemalloc vs Tcmalloc
    - Jemalloc设计上比前两个复杂地多，其内部使用了红黑树管理分页和内存块。
    - Jemalloc对内存分配粒度分类地更细。这导致一方面比ptmalloc的锁争用要少，另一方面很多索引和查找都能回归到指数级别，方便了很多复杂功能的实现。
    - 在大内存分配上，内存碎片也会比tcmalloc少。
    - 因为他的结构比较复杂，记录了很多meta，所以在分配很多小内存的时候记录meta数据的空间会略微多于tcmalloc。但是又不像ptmalloc那样每一个内存块都有一个header，而采用全局的bitmap记录状态，所以大量小内存的时候，会比ptmalloc消耗的额外内存小。
- [ptmalloc、tcmalloc与jemalloc](https://www.cyningsun.com/07-07-2018/memory-allocator-contrasts.html)
  - Issue
    - 为了进行耗时优化，基础库这层按照惯例使用tcmalloc替代glibc标配的ptmalloc做优化，CPU消耗和耗时确实有所降低。但在晚上高峰时期，在CPU刚刚超过50%之后却出现了指数上升，服务在几分钟之内不可用。最终定位到是tcmalloc在内存分配的时候使用自旋锁，在锁冲突严重的时候导致CPU飙升
  - ptmalloc
    - 当某一线程需要调用malloc()分配内存空间时，该线程先查看线程私有变量中是否已经存在一个分配区，如果存在，尝试对该分配区加锁，如果加锁成功，使用该分配区分配内存，如果失败，该线程搜索循环链表试图获得一个没有加锁的分配区。如果所有的分配区都已经加锁，那么malloc()会开辟一个新的分配区，把该分配区加入到全局分配区循环链表并加锁，然后使用该分配区进行分配内存操作。在释放操作中，线程同样试图获得待释放内存块所在分配区的锁，如果该分配区正在被别的线程使用，则需要等待直到其他线程释放该分配区的互斥锁之后才可以进行释放操作
    - 从工作原理来看：
      - Fast bins是小内存块的高速缓存，当一些大小小于64字节的chunk被回收时，首先会放入fast bins中，在分配小内存时，首先会查看fast bins中是否有合适的内存块，如果存在，则直接返回fast bins中的内存块，以加快分配速度。
      - Usorted bin只有一个，回收的chunk块必须先放到unsorted bin中，分配内存时会查看unsorted bin中是否有合适的chunk，如果找到满足条件的chunk，则直接返回给用户，否则将unsorted bin的所有chunk放入small bins或是large bins中。
      - Small bins用于存放固定大小的chunk，共64个bin，最小的chunk大小为16字节或32字节，每个bin的大小相差8字节或是16字节，当分配小内存块时，采用精确匹配的方式从small bins中查找合适的chunk。
      - Large bins用于存储大于等于512B或1024B的空闲chunk，这些chunk使用双向链表的形式按大小顺序排序，分配内存时按最近匹配方式从large bins中分配chunk。
    ![img.png](os_ptmalloc.png)
    - 问题
      - 如果后分配的内存先释放，无法及时归还系统。因为 ptmalloc 收缩内存是从 top chunk 开始,如果与 top chunk 相邻的 chunk 不能释放, top chunk 以下的 chunk 都无法释放。
      - 内存不能在线程间移动，多线程使用内存不均衡将导致内存浪费
      - 每个chunk至少8字节的开销很大
      - 不定期分配长生命周期的内存容易造成内存碎片，不利于回收。
      - 加锁耗时，无论当前分区有无耗时，在内存分配和释放时，会首先加锁。
  - tcmalloc
    - TCMalloc是专门对多线并发的内存管理而设计的，TCMalloc主要是在线程级实现了缓存，使得用户在申请内存时大多情况下是无锁内存分配。整个 TCMalloc 实现了三级缓存，分别是ThreadCache(线程级缓存)，Central Cache(中央缓存：CentralFreeeList)，PageHeap(页缓存)，最后两级需要加锁访问
      ![img.png](os_tcmalloc.png)
    - tcmalloc的优势
      - 小内存可以在ThreadCache中不加锁分配(加锁的代价大约100ns)
      - 大内存可以直接按照大小分配不需要再像ptmalloc一样进行查找
      - 大内存加锁使用更高效的自旋锁
      - 减少了内存碎片
    - 问题
      - 使用自旋锁虽然减少了加锁效率，但是如果使用大内存较多的情况下，内存在Central Cache或者Page Heap加锁分配。而tcmalloc对大小内存的分配过于保守，在一些内存需求较大的服务（如推荐系统），小内存上限过低，当请求量上来，锁冲突严重，CPU使用率将指数暴增
  - jemalloc
    - jemalloc最大的优势还是其强大的多核/多线程分配能力. CPU的核心数量越多, 程序线程数越多, jemalloc的分配速度越快
    - jemalloc 按照内存分配请求的尺寸，分了 small object (例如 1 – 57344B)、 large object (例如 57345 – 4MB )、 huge object (例如 4MB以上)。jemalloc同样有一层线程缓存的内存名字叫tcache，当分配的内存大小小于tcache_maxclass时，jemalloc会首先在tcache的small object以及large object中查找分配，tcache不中则从arena中申请run，并将剩余的区域缓存到tcache。若arena找不到合适大小的内存块， 则向系统申请内存。当申请大小大于tcache_maxclass且大小小于huge大小的内存块时，则直接从arena开始分配。而huge object的内存不归arena管理， 直接采用mmap从system memory中申请，并由一棵与arena独立的红黑树进行管理。
    ![img.png](os_gemalloc.png)
- [Tunes EC2](https://www.brendangregg.com/blog/2017-12-31/reinvent-netflix-ec2-tuning.html)
  - for Ubuntu Xenial instances on EC2.
    - CPU
      `schedtool –B PID`
    - Virtual Memory
      ```shell
      vm.swappiness = 0       # from 60
      ```
    - Huge Pages
      ```shell
      # echo madvise > /sys/kernel/mm/transparent_hugepage/enabled
      ```
    - NUMA
      ```shell
      kernel.numa_balancing = 0
      ```
    - File System
      ```shell
      vm.dirty_ratio = 80                     # from 40
      vm.dirty_background_ratio = 5           # from 10
      vm.dirty_expire_centisecs = 12000       # from 3000
      mount -o defaults,noatime,discard,nobarrier …
      ```
    - Storage I/O
      ```shell
      /sys/block/*/queue/rq_affinity  2
      /sys/block/*/queue/scheduler        noop
      /sys/block/*/queue/nr_requests  256
      /sys/block/*/queue/read_ahead_kb    256
      mdadm –chunk=64 ...
      ```
    - Networking
      ```shell
      net.core.somaxconn = 1000
      net.core.netdev_max_backlog = 5000
      net.core.rmem_max = 16777216
      net.core.wmem_max = 16777216
      net.ipv4.tcp_wmem = 4096 12582912 16777216
      net.ipv4.tcp_rmem = 4096 12582912 16777216
      net.ipv4.tcp_max_syn_backlog = 8096
      net.ipv4.tcp_slow_start_after_idle = 0
      net.ipv4.tcp_tw_reuse = 1
      net.ipv4.ip_local_port_range = 10240 65535
      net.ipv4.tcp_abort_on_overflow = 1    # maybe
      ```
- [Network Overview](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247563566&idx=1&sn=26156d79dffb3f0f10b6a26931f993cc&chksm=c1850e7ff6f28769b6ff3358366e917d3d54fc0f0563131422da4bed201768c958262b5d5a99&scene=21#wechat_redirect)
  - 协议
    - TCP UDP QUIC
  - 网络编程
    - Reactor是基于同步IO,事件驱动机制，reactor实现了一个被动事件接收和分发的模型，同步等待事件到来，并作出响应，Reactor实现相对简单，对于耗时短的处理场景处理高效，同时接收多个服务请求，并且依次同步的处理它们的事件驱动程序
      - redis - level trigger
      - nginx - edge trigger
    - Proactor基于异步IO，异步接收和同时处理多个服务请求的事件驱动程序，处理速度快，Proactor性能更高，能够处理耗时长的并发场景
  - Linux内核协议栈
    - [TCP/IP协议栈](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247510568&idx=1&sn=79f335aaab5c0a36c0a66c5bfb1619ae&chksm=c1845d79f6f3d46f81b6fd24335eb8994c9daf21b6846d80af2cad73d9f638c5dda48b02892c&scene=21#wechat_redirect)
    - [Linux网络子系统](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247532046&idx=2&sn=04ffe282ce1278297d124f0c382ba665&chksm=c184895ff6f300497eb2bcc63d352b6d6b374606399cb7dd5b5bb59a773e674a368f9f4c9169&scene=21#wechat_redirect)
  - [新技术基石 | eBPF and XDP](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247511570&idx=1&sn=18d5f1045b7ed6e27e1c8fb36302c3f9&chksm=c1845943f6f3d055c2533d580acb2d4258daf2a84ffba30e2cc6376b3ffdb1687773ed4c19a3&scene=21#wechat_redirect)
    - 传统Linux网络驱动的问题
      - 中断开销突出，大量数据到来会触发频繁的中断（softirq）开销导致系统无法承受，
      - 需要把包从内核缓冲区拷贝到用户缓冲区，带来系统调用和数据包复制的开销，
      - 对于很多网络功能节点来说，TCP/IP协议并非是数据转发环节所必需的，
      - NAPI/Netmap等虽然减少了内核到用户空间的数据拷贝，但操作系统调度带来的cache替换也会对性能产生负面影响。
    - 改善iptables/netfilter的规模瓶颈，提高Linux内核协议栈IO性能，内核需要提供新解决方案，那就是eBPF/XDP框架
    - BPF 是 Linux内核中高度灵活和高效的类似虚拟机的技术，允许以安全的方式在各个挂钩点执行字节码。它用于许多Linux内核子系统，最突出的是网络、跟踪和安全（例如沙箱）。
    - XDP的全称是：eXpress DataPath，XDP 是Linux内核中提供高性能、可编程的网络数据包处理框架。
      - 直接接管网卡的RX数据包（类似DPDK用户态驱动）处理
      - 通过运行BPF指令快速处理报文；
      - 和Linux协议栈无缝对接；
  - 容器网络
    - Kubernetes本身并没有自己实现容器网络，而是通过插件化的方式自由接入进来。在容器网络接入进来需要满足如下基本原则：
      - Pod无论运行在任何节点都可以互相直接通信，而不需要借助NAT地址转换实现。
      - Node与Pod可以互相通信，在不限制的前提下，Pod可以访问任意网络。
      - Pod拥有独立的网络栈，Pod看到自己的地址和外部看见的地址应该是一样的，并且同个Pod内所有的容器共享同个网络栈。
    - 目前流行插件：Flannel、Calico、Weave、Contiv
    - Overlay模式：Flannel（UDP、vxlan）、Weave、Calico（IPIP）
    - 三层路由模式：Flannel（host-gw）、Calico（BGP）
    - Underlay网络：Calico（BGP）
  - [网络性能优化](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247505264&idx=1&sn=0735027d6e20ca7bd2bdc7a7c04477f1&chksm=c1842221f6f3ab371d64032b5bae6e5cc65d8f0265745150f76794932c4d3ac8bf4058d60090&scene=21#wechat_redirect)
    - APP性能优化：空间局部性和时间局部性
    - 协议栈调优：numa亲和，中断亲和，软中断亲和，软中断队列大小，邻居表大小，连接跟踪表大小，TCP队列调整，协议栈内存分配，拥塞控制算法优化
    - Bypass内核协议栈：DPDK,RDMA.
    - 网卡优化：驱动配置，offload，网卡队列，智能网卡，高性能网卡。
    - 硬件加速器：ASIC、FPGA,网络处理器,多核处理器
  - 高性能网络
    - DPDK网络优化
      - PMD用户态驱动•CPU亲缘性和独占•内存大页和降低内存访问开销•避免False Sharing•内存对齐•cache对齐•NUMA•减少进程上下文切换•分组预测机制•利用流水线并发•为了利用空间局部性•充分挖掘网卡的潜能
    - RDMA 作为一种旁路内核的远程内存直接访问技术，被广泛应用于数据密集型和计算密集型场景中，是高性能计算、机器学习、数据中心、海量存储等领域的重要解决方案。
      - RDMA 具有零拷贝、协议栈卸载的特点。RDMA 将协议栈的实现下沉至RDMA网卡(RNIC)，绕过内核直接访问远程内存中的数据
  - [排障](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247505250&idx=1&sn=a854ee9a456e27e3bd4202380e4782c8&chksm=c1842233f6f3ab251f1a686e4f4bbeaa305a73ff3b09d5846b3ae536153ed22ba716c99f0ae5&scene=21#wechat_redirect)

- [Linux问题分析与性能优化](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247529939&idx=2&sn=f193d1ea112482945d8be3d97d1eba85&chksm=c1848282f6f30b94a2b1851d2a18a8ce0e026cc79b5ace4076909389704aaab185ccdc9c5bef&scene=21#wechat_redirect) 
  - 整体情况：
    - top/htop/atop命令查看进程/线程、CPU、内存使用情况，CPU使用情况；
    - dstat 2查看CPU、磁盘IO、网络IO、换页、中断、切换，系统I/O状态;
    - vmstat 2查看内存使用情况，内存状态；
    - iostat -d -x 2查看所有磁盘的IO情况，系统I/O状态；
    - iotop查看IO靠前的进程，系统的I/O状态；
    - perf top查看占用CPU最多的函数，CPU使用情况；
    - perf record -ag -- sleep 15;perf report查看CPU事件占比，调用栈，CPU使用情况；
    - sar -n DEV 2查看网卡的吞吐，网卡状态；
    - /usr/share/bcc/tools/filetop -C查看每个文件的读写情况，系统的I/O状态；
    - /usr/share/bcc/tools/opensnoop显示正在被打开的文件，系统的I/O状态；
    - mpstat -P ALL 1 单核CPU是否被打爆；
    - ps aux --sort=-%cpu 按CPU使用率排序，找出CPU消耗最多进程；
    - ps -eo pid,comm,rss | awk '{m=$3/1e6;s["*"]+=m;s[$2]+=m} END{for (n in s) printf"%10.3f GB  %s\n",s[n],n}' | sort -nr | head -20 统计前20内存占用；
    - awk 'NF>3{s["*"]+=s[$1]=$3*$4/1e6} END{for (n in s) printf"%10.1f MB  %s\n",s[n],n}' /proc/slabinfo | sort -nr | head -20  统计内核前20slab的占用；
  - 进程分析，进程占用的资源：
    - pidstat 2 -p 进程号查看可疑进程CPU使用率变化情况；
    - pidstat -w -p 进程号 2查看可疑进程的上下文切换情况；
    - pidstat -d -p 进程号 2查看可疑进程的IO情况；
    - lsof -p 进程号查看进程打开的文件；
    - strace -f -T -tt -p 进程号显示进程发起的系统调用；
  - 协议栈分析，连接/协议栈状态：
    - ethtool -S 查看网卡硬件情况；
    - cat /proc/net/softnet_stat/ifconfig eth1 查看网卡驱动情况；
    - netstat -nat|awk '{print awk $NF}'|sort|uniq -c|sort -n查看连接状态分布；
    - ss -ntp或者netstat -ntp查看连接队列；
    - netstat -s 查看协议栈情况；
  - 方法论
    - RED方法：监控服务的请求数（Rate）、错误数（Errors）、响应时间（Duration）
    - USE方法：监控系统资源的使用率（Utilization）、饱和度（Saturation）、错误数（Errors）。
  - Tools
    - ![img.png](os_cpu_cmd.png)
    - ![img_1.png](os_memory_cmd.png)
    - ![img.png](os_file_cmd.png)
    - ![img_2.png](img_2.png)
- [服务器性能优化之网络性能优化](https://mp.weixin.qq.com/s?__biz=MzkyMTIzMTkzNA==&mid=2247561718&idx=1&sn=f93ad69bff3ab80665e4b9d67265e6bd&chksm=c18506a7f6f28fb12341c3e439f998d09c4b1d93f8bf59af6b1c6f4427cea0c48b51244a3e53&scene=21#wechat_redirect)
  - 以前
    - 网卡很慢，只有一个队列。当数据包到达时，网卡通过DMA复制数据包并发送中断，Linux内核收集这些数据包并完成中断处理。随着网卡越来越快，基于中断的模型可能会因大量传入数据包而导致 IRQ 风暴。
    - 为了解决这个问题，NAPI(中断+轮询)被提议。当内核收到来自网卡的中断时，它开始轮询设备并尽快收集队列中的数据包
    - ![img_1.png](os_performance_network.png)
  - RSS：接收端缩放 Receive Side Scaling
    - 具有多个RX / TX队列过程的数据包。当带有RSS 的网卡接收到数据包时，它会对数据包应用过滤器并将数据包分发到RX 队列。过滤器通常是一个哈希函数，可以通过“ethtool -X”进行配置
    - CPU 亲和性也很重要。最佳设置是分配一个 CPU 专用于一个队列。首先通过检查/proc/interrupt找出IRQ号，然后将CPU位掩码设置为/proc/irq/<IRQ_NUMBER>/smp_affinity来分配专用CPU。为避免设置被覆盖，必须禁用守护进程irqbalance。
  - RPS：接收数据包控制
    - RSS提供硬件队列，一个称为软件队列机制Receive Packet Steering （RPS）在Linux内核实现
    - 当驱动程序接收到数据包时，它会将数据包包装在套接字缓冲区 ( sk_buff ) 中，其中包含数据包的u32哈希值。散列是所谓的第 4 层散列（l4 散列），它基于源 IP、源端口、目的 IP 和目的端口，由网卡或__skb_set_sw_hash() 计算。由于相同 TCP/UDP 连接（流）的每个数据包共享相同的哈希值，因此使用相同的 CPU 处理它们是合理的。
    - RPS 的基本思想是根据每个队列的 rps_map 将同一流的数据包发送到特定的 CPU。这是 rps_map 的结构：映射根据 CPU 位掩码动态更改为`/sys/class/net/<dev>/queues/rx-<n>/rps_cpus`。
    - ![img.png](os_network_rps.png)
  - RFS: Receive Flow Steering
    - 尽管 RPS 基于流分发数据包，但它没有考虑用户空间应用程序。应用程序可能在 CPU A 上运行，而内核将数据包放入 CPU B 的队列中。由于 CPU A 只能使用自己的缓存，因此 CPU B 中缓存的数据包变得无用。
    - 代替每个队列的哈希至CPU地图，RFS维护全局flow-to-CPU的表，rps_sock_flow_table：该掩模用于将散列值映射成所述表的索引。
    - ![img.png](os_network_rfs.png)
  - aRFS: Accelerated Receive Flow Steering
    - aRFS 进一步延伸RFS为RX队列硬件过滤。要启用 aRFS，它需要具有可编程元组过滤器和驱动程序支持的网卡。要启用ntuple 过滤器。
  - SO_REUSEPORT
    - SO_REUSEPORT支持多个进程或者线程绑定到同一端口，用以提高服务器程序的性能
      - 允许多个套接字 bind()/listen() 同一个TCP/UDP端口
        - 1.每一个线程拥有自己的服务器套接字。
        - 2.在服务器套接字上没有了锁的竞争。
      - 内核层面实现负载均衡。
      - 安全层面，监听同一个端口的套接字只能位于同一个用户下面。
    - 其核心的实现主要有三点：
      - 扩展socket option，增加 SO_REUSEPORT选项，用来设置 reuseport。
      - 修改 bind 系统调用实现，以便支持可以绑定到相同的 IP 和端口。
      - 修改处理新建连接的实现，查找 listener 的时候，能够支持在监听相同 IP 和端口的多个 sock 之间均衡选择
    - 带来意义
      - CPU之间平衡处理，水平扩展，模型简单，维护方便了，进程的管理和应用逻辑解耦，进程的管理水平扩展权限下放给程序员/管理员，可以根据实际进行控制进程启动/关闭，增加了灵活性。
      - 针对对客户端而言，表面上感受不到其变动，因为这些工作完全在服务器端进行。
      - 服务器无缝重启/切换，热更新，提供新的可能性
    - 已知问题
      - SO_REUSEPORT分为两种模式，即热备份模式和负载均衡模式，在早期的内核版本中，即便是加入对reuseport选项的支持，也仅仅为热备份模式，而在3.9内核之后，则全部改为了负载均衡模式
      - SO_REUSEPORT根据数据包的四元组{src ip, src port, dst ip, dst port}和当前绑定同一个端口的服务器套接字数量进行数据包分发。若服务器套接字数量产生变化，内核会把本该上一个服务器套接字所处理的客户端连接所发送的数据包（比如三次握手期间的半连接，以及已经完成握手但在队列中排队的连接）分发到其它的服务器套接字上面，可能会导致客户端请求失败。
    - 如何预防以上已知问题，一般解决思路：
      - 使用固定的服务器套接字数量，不要在负载繁忙期间轻易变化。
      - 允许多个服务器套接字共享TCP请求表(Tcp request table)。
      - 不使用四元组作为Hash值进行选择本地套接字处理，比如选择 会话ID或者进程ID，挑选隶属于同一个CPU的套接字。
      - 使用一致性hash算法。
    - 演进
      - 3.9之前内核，能够让多个socket同时绑定完全相同的ip+port，但不能实现负载均衡，实现是热备。
      - Linux 3.9之后，能够让多个socket同时绑定完全相同的ip+port，可以实现负载均衡。
      - Linux4.5版本后，内核引入了reuseport groups，它将绑定到同一个IP和Port，并且设置了SO_REUSEPORT选项的socket组织到一个group内部。目的是加快socket查询
    - 与其他特性关系
      - SO_REUSEADDR：主要是地址复用
        - 让处于time_wait状态的socket可以快速复用原ip+port
        - 使得0.0.0.0（ipv4通配符地址）与其他地址（127.0.0.1和10.0.0.x）不冲突
        - SO_REUSEADDR 的缺点在于，没有安全限制，而且无法保证所有连接均匀分配。
- [文件的 io 栈](https://mp.weixin.qq.com/s/IrZF9lWweEs1rhxuvMUCKA)
  - IO 从用户态走系统调用进到内核，内核的路径：`VFS → 文件系统 → 块层 → SCSI 层 `
    - VFS 负责通用的文件抽象语义，管理并切换文件系统；
    - 文件系统负责抽象出“文件的概念”，维护“文件”数据到块层的位置映射，怎么摆放数据，怎么抽象文件都是文件系统说了算；
    - 块层对底层硬件设备做一层统一的抽象，最重要的是做一些 IO 调度的策略。比如，尽可能收集批量 IO 聚合下发，让 IO 尽可能的顺序，合并 IO 请求减少 IO 次数等等；
    - SCSI 层则是负责最后对硬件磁盘的对接，驱动层，本质就是个翻译器
  - page cache 是发生在文件系统这里。通常我们确保数据落盘有两种方式：
    - Writeback 回刷数据的方式：write 调用 + sync 调用；
    - Direct IO 直刷数据的方式；
- [零拷贝](https://mp.weixin.qq.com/s/K-5HJCxDzjZuHhWk1SPEQQ)
  - 零拷贝是指计算机执行IO操作时，CPU不需要将数据从一个存储区域复制到另一个存储区域，从而可以减少上下文切换以及CPU的拷贝时间。它是一种I/O操作优化技术。
  - 传统的IO流程，包括read和write的过程。4次数据拷贝（两次CPU拷贝以及两次的DMA拷贝)
    - read：把数据从磁盘读取到内核缓冲区，再拷贝到用户缓冲区
    - write：先把数据写入到socket缓冲区，最后写入网卡设备。
    ![img.png](os_write_read.png)
    - DMA，英文全称是Direct Memory Access，即直接内存访问。DMA本质上是一块主板上独立的芯片，允许外设设备和内存存储器之间直接进行IO数据传输，其过程不需要CPU的参与。
  - 零拷贝并不是没有拷贝数据，而是减少用户态/内核态的切换次数以及CPU拷贝的次数。零拷贝实现有多种方式，分别是
    - mmap+write：2次DMA拷贝和1次CPU拷贝
      ![img.png](os_mmap_write.png)
    - sendfile： 2次DMA拷贝和1次CPU拷贝
      ![img.png](os_sendfile.png)
    - 带有DMA收集拷贝功能的sendfile: 2次数据拷贝都是包DMA拷贝
      - linux 2.4版本之后，对sendfile做了优化升级，引入SG-DMA技术，其实就是对DMA拷贝加入了scatter/gather操作，它可以直接从内核空间缓冲区中将数据读取到网卡。使用这个特点搞零拷贝，即还可以多省去一次CPU拷贝
      ![img.png](os_sendfile_scattergatter.png)
- [进程和线程19个问题](https://mp.weixin.qq.com/s/NCl17jrOwP_A017nUqOkJQ)
- [进程调度](https://mp.weixin.qq.com/s/uBa65Vd3WsZsIv2uQy3cHQ)
  - O(n)调度器采用全局runqueue，导致多cpu加锁问题和cache利用率低的问题
  - O(1)调度器为每个cpu设计了一个runqueue，并且采用MLFQ算法思想设置140个优先级链表和active/expire两个双指针结构
  - CFS调度器采用红黑树来实现O(logn)复杂度的pick-next算法，摒弃固定时间片机制，采用调度周期内的动态时间机制
  - O(1)和O(n)都在交互进程的识别算法上下了功夫，但是无法做的100%准确
  - CFS另辟蹊径采用完全公平思想以及虚拟运行时间来实现进行的调度
  - CFS调度器也并非银弹，在某些方面可能不如O(1)
- [Linux 性能优化全景指南](https://mp.weixin.qq.com/s/6_utyj1kCyC5ZWpveDZQIQ)
  - 性能优化
    - 高并发和响应快对应着性能优化的两个核心指标：吞吐和延时
    - 平均负载：单位时间内，系统处于可运行状态和不可中断状态的平均进程数，也就是平均活跃进程数。
    - 平均负载高时可能是CPU密集型进程导致，也可能是I/O繁忙导致。具体分析时可以结合`mpstat/pidstat`工具辅助分析负载来源
  - CPU
    - CPU上下文切换分为：
      - 进程上下文切换: 一次系统调用过程其实进行了两次CPU上下文切换
      - 线程上下文切换
      - 中断上下文切换: 中断上下文只包括内核态中断服务程序执行所必须的状态
    - 通过vmstat可以查看系统总体的上下文切换情况
    - 使用pidstat来查看每个进程上下文切换情况
      ```shell
      vmstat 1 1    #首先获取空闲系统的上下文切换次数
      sysbench --threads=10 --max-time=300 threads run #模拟多线程切换问题
      
      vmstat 1 1    #新终端观察上下文切换情况
      此时发现cs数据明显升高，同时观察其他指标：
      r列： 远超系统CPU个数，说明存在大量CPU竞争
      us和sy列：sy列占比80%，说明CPU主要被内核占用
      in列： 中断次数明显上升，说明中断处理也是潜在问题
      
      说明运行/等待CPU的进程过多，导致大量的上下文切换，上下文切换导致系统的CPU占用率高
      
      pidstat -w -u 1  #查看到底哪个进程导致的问题, 分析sysbench模拟的是线程的切换，因此需要在pidstat后加-t参数查看线程指标。
      
      另外对于中断次数过多，我们可以通过/proc/interrupts文件读取 `watch -d cat /proc/interrupts`
      ```
    - 某个应用的CPU使用率达到100%，怎么办？
      - CPU使用率 - 除了空闲时间以外的其他时间占总CPU时间的百分比。可以通过/proc/stat中的数据来计算出CPU使用率。
      - 分析进程的CPU问题可以通过perf，它以性能事件采样为基础 `perf top / perf record / perf report `
        ```shell
        sudo docker run --name nginx -p 10000:80 -itd feisky/nginx
        sudo docker run --name phpfpm -itd --network container:nginx feisky/php-fpm
        
        ab -c 10 -n 100 http://XXX.XXX.XXX.XXX:10000/ #测试Nginx服务性能
        发现此时每秒可承受请求给长少，此时将测试的请求数从100增加到10000。在另外一个终端运行top查看每个CPU的使用率。发现系统中几个php-fpm进程导致CPU使用率骤升。
        
        接着用perf来分析具体是php-fpm中哪个函数导致该问题。`perf top -g -p XXXX #对某一个php-fpm进程进行分析`
        ```
    - 系统的CPU使用率很高，为什么找不到高CPU的应用？
       ```shell
       sudo docker run --name nginx -p 10000:80 -itd feisky/nginx:sp
       sudo docker run --name phpfpm -itd --network container:nginx feisky/php-fpm:sp
       ab -c 100 -n 1000 http://XXX.XXX.XXX.XXX:10000/ #并发100个请求测试
       
       此时用top和pidstat发现系统CPU使用率过高，但是并没有发现CPU使用率高的进程。
       出现这种情况一般时我们分析时遗漏的什么信息，重新运行top命令并观察一会。发现就绪队列中处于Running状态的进行过多，超过了我们的并发请求次数5. 再仔细查看进程运行数据，发现nginx和php-fpm都处于sleep状态，真正处于运行的却是几个stress进程。
       ```
      - 此时有可能时以下两种原因导致：
        - 进程不停的崩溃重启（如段错误/配置错误等），此时进程退出后可能又被监控系统重启；
        - 短时进程导致，即其他应用内部通过exec调用的外面命令，这些命令一般只运行很短时间就结束，很难用top这种间隔较长的工具来发现
      - 可以通过pstree来查找 stress的父进程，找出调用关系。 `pstree | grep stress`
    - 系统中出现大量不可中断进程和僵尸进程怎么办？
      - 不可中断状态
        - 进程状态 D Disk Sleep，不可中断状态睡眠，一般表示进程正在跟硬件交互，并且交互过程中不允许被其他进程中断
        - 对于不可中断状态，一般都是在很短时间内结束，可忽略。但是如果系统或硬件发生故障，进程可能会保持不可中断状态很久，甚至系统中出现大量不可中断状态，此时需注意是否出现了I/O性能问题。
      - 大量的僵尸进程会用尽PID进程号，导致新进程无法建立
      - 磁盘O_DIRECT问题
        ```shell
        sudo docker run --privileged --name=app -itd feisky/app:iowait
        ps aux | grep '/app'
        ```
        - 可以看到此时有多个app进程运行，状态分别时Ss+和D+。其中后面s表示进程是一个会话的领导进程，+号表示前台进程组。
        - 进程组表示一组相互关联的进程，子进程是父进程所在组的组员。会话指共享同一个控制终端的一个或多个进程组
        - top查看系统资源发现：
          - 1）平均负载在逐渐增加，且1分钟内平均负载达到了CPU个数，说明系统可能已经有了性能瓶颈；
          - 2）僵尸进程比较多且在不停增加；
          - 3）us和sys CPU使用率都不高，iowait却比较高；
          - 4）每个进程CPU使用率也不高，但有两个进程处于D状态，可能在等待IO。
        - iowait过高导致系统平均负载升高，僵尸进程不断增长说明有程序没能正确清理子进程资源。
        - 用dstat`dstat 1 10    #间隔1秒输出10组数据`来分析，因为它可以同时查看CPU和I/O两种资源的使用情况，便于对比分析
        - 之前top查看的处于D状态的进程号，用pidstat -d -p XXX 展示进程的I/O统计数据。发现处于D状态的进程都没有任何读写操作。
        - 在用pidstat -d 查看所有进程的I/O统计数据，看到app进程在进行磁盘读操作，每秒读取32MB的数据。进程访问磁盘必须使用系统调用处于内核态，接下来重点就是找到app进程的系统调用。
        - 用perf record -d和perf report进行分析，查看app进程调用栈。
        - 看到app确实在通过系统调用sys_read()读取数据，并且从new_sync_read和blkdev_direct_IO看出进程时进行直接读操作，请求直接从磁盘读，没有通过缓存导致iowait升高。
           通过层层分析后，root cause是app内部进行了磁盘的直接I/O
      - 僵尸进程
        - 首先要定位僵尸进程的父进程，通过pstree -aps XXX，打印出该僵尸进程的调用树，发现父进程就是app进程。
        - 查看app代码，看看子进程结束的处理是否正确（是否调用wait()/waitpid(),有没有注册SIGCHILD信号的处理函数等）
      - Summary
        - 碰到iowait升高时，先用dstat pidstat等工具确认是否存在磁盘I/O问题，再找是哪些进程导致I/O，不能用strace直接分析进程调用时可以通过perf工具分析。
        - 对于僵尸问题，用pstree找到父进程，然后看源码检查子进程结束的处理逻辑即可。
    - CPU性能指标
      - CPU使用率
        - 用户CPU使用率, 包括用户态(user)和低优先级用户态(nice). 该指标过高说明应用程序比较繁忙.
        - 系统CPU使用率, CPU在内核态运行的时间百分比(不含中断). 该指标高说明内核比较繁忙.
        - 等待I/O的CPU使用率, iowait, 该指标高说明系统与硬件设备I/O交互时间比较长.
        - 软/硬中断CPU使用率, 该指标高说明系统中发生大量中断.
        - steal CPU / guest CPU, 表示虚拟机占用的CPU百分比.
      - 平均负载理想情况下平均负载等于逻辑CPU个数,表示每个CPU都被充分利用. 若大于则说明系统负载较重.
      - 进程上下文切换包括无法获取资源的自愿切换和系统强制调度时的非自愿切换. 上下文切换本身是保证Linux正常运行的一项核心功能. 过多的切换则会将原本运行进程的CPU时间消耗在寄存器,内核占及虚拟内存等数据保存和恢复上
      - CPU缓存命中率CPU缓存的复用情况,命中率越高性能越好. 其中L1/L2常用在单核,L3则用在多核中
    - 性能工具
      - 平均负载案例
        - 先用uptime查看系统平均负载
        - 判断负载在升高后再用mpstat和pidstat分别查看每个CPU和每个进程CPU使用情况.找出导致平均负载较高的进程.
      - 上下文切换案例
        - 先用vmstat查看系统上下文切换和中断次数
        - 再用pidstat观察进程的自愿和非自愿上下文切换情况
        - 最后通过pidstat观察线程的上下文切换情况
      - 进程CPU使用率高案例
        - 先用top查看系统和进程的CPU使用情况,定位到进程
        - 再用perf top观察进程调用链,定位到具体函数
      - 系统CPU使用率高案例
        - 先用top查看系统和进程的CPU使用情况,top/pidstat都无法找到CPU使用率高的进程
        - 重新审视top输出
        - 从CPU使用率不高,但是处于Running状态的进程入手
        - perf record/report发现短时进程导致 (execsnoop工具)
      - 不可中断和僵尸进程案例
        - 先用top观察iowait升高,发现大量不可中断和僵尸进程
        - strace无法跟踪进程系统调用
        - perf分析调用链发现根源来自磁盘直接I/O
      - 软中断案例
        - top观察系统软中断CPU使用率高
        - 查看/proc/softirqs找到变化速率较快的几种软中断
        - sar命令发现是网络小包问题
        - tcpdump找出网络帧的类型和来源, 确定SYN FLOOD攻击导致
      ![img.png](os_cpu_tool.png)
      ![img.png](os_cpu_tool1.png)
      - 先运行几个支持指标较多的工具, 如top/vmstat/pidstat,根据它们的输出可以得出是哪种类型的性能问题. 
      - 定位到进程后再用strace/perf分析调用情况进一步分析. 如果是软中断导致用/proc/softirqs
      ![img.png](os_cpu_tool2.png)
    - CPU优化
      - 应用程序优化
        - 编译器优化: 编译阶段开启优化选项, 如gcc -O2
        - 算法优化
        - 异步处理: 避免程序因为等待某个资源而一直阻塞,提升程序的并发处理能力. (将轮询替换为事件通知)
        - 多线程代替多进程: 减少上下文切换成本
        - 善用缓存: 加快程序处理速度
      - 系统优化
        - CPU绑定: 将进程绑定要1个/多个CPU上,提高CPU缓存命中率,减少CPU调度带来的上下文切换
        - CPU独占: CPU亲和性机制来分配进程
        - 优先级调整:使用nice适当降低非核心应用的优先级
        - 为进程设置资源显示: cgroups设置使用上限,防止由某个应用自身问题耗尽系统资源
        - NUMA优化: CPU尽可能访问本地内存
        - 中断负载均衡: irpbalance,将中断处理过程自动负载均衡到各个CPU上
  - 内存
    - Linux内存是怎么工作的
      - 内存映射
        - 大多数计算机用的主存都是动态随机访问内存(DRAM)，只有内核才可以直接访问物理内存。Linux内核给每个进程提供了一个独立的虚拟地址空间，并且这个地址空间是连续的。这样进程就可以很方便的访问内存(虚拟内存)。
        - 为了完成内存映射，内核为每个进程都维护了一个页表，记录虚拟地址和物理地址的映射关系。页表实际存储在CPU的内存管理单元MMU中，处理器可以直接通过硬件找出要访问的内存。
        - MMU以页为单位管理内存，页大小4KB。为了解决页表项过多问题Linux提供了多级页表和HugePage的机制。
      - 内存分配与回收
        - 分配
          - **brk()** 针对小块内存(<128K)，通过移动堆顶位置来分配。内存释放后不立即归还内存，而是被缓存起来。
          - **mmap()** 针对大块内存(>128K)，直接用内存映射来分配，即在文件映射段找一块空闲内存分配。
          - 上述两种调用并没有真正分配内存，这些内存只有在首次访问时，才通过缺页异常进入内核中，由内核来分配
        - 回收
          - 回收缓存：LRU算法回收最近最少使用的内存页面；
          - 回收不常访问内存：把不常用的内存通过交换分区写入磁盘
          - 杀死进程：OOM内核保护机制 `echo -16 > /proc/$(pidof XXX)/oom_adj`
      - 如何查看内存使用情况 - free
      - 怎样理解内存中的Buffer和Cache
        - buffer是对磁盘数据的缓存，cache是对文件数据的缓存，它们既会用在读请求也会用在写请求中
      - 如何利用系统缓存优化程序的运行效率
        - 安装bcc包后可以通过cachestat和cachetop来监测缓存的读写命中情况。
        - 安装pcstat后可以查看文件在内存中的缓存大小以及缓存比例
    - 内存泄漏，如何定位和处理？
      - 可以通过memleak工具来跟踪系统或进程的内存分配/释放请求
      - `/usr/share/bcc/tools/memleak -a -p $(pidof app)`
    - 为什么系统的Swap变高
      - Swap本质就是把一块磁盘空间或者一个本地文件当作内存来使用，包括换入和换出两个过程
      - NUMA 与 SWAP
        - 很多情况下系统剩余内存较多，但SWAP依旧升高，这是由于处理器的NUMA架构
        - 在NUMA架构下多个处理器划分到不同的Node，每个Node都拥有自己的本地内存空间。在分析内存的使用时应该针对每个Node单独分析 `numactl --hardware`
      - swappiness
        - 在实际回收过程中Linux根据/proc/sys/vm/swapiness选项来调整使用Swap的积极程度，从0-100，数值越大越积极使用Swap，即更倾向于回收匿名页；数值越小越消极使用Swap，即更倾向于回收文件页
      - Swap升高时如何定位分析
        ```shell
        free #首先通过free查看swap使用情况，若swap=0表示未配置Swap
        #先创建并开启swap
        fallocate -l 8G /mnt/swapfile
        chmod 600 /mnt/swapfile
        mkswap /mnt/swapfile
        swapon /mnt/swapfile
        
        free #再次执行free确保Swap配置成功
        
        dd if=/dev/sda1 of=/dev/null bs=1G count=2048 #模拟大文件读取
        sar -r -S 1  #查看内存各个指标变化 -r内存 -S swap
        #根据结果可以看出，%memused在不断增长，剩余内存kbmemfress不断减少，缓冲区kbbuffers不断增大，由此可知剩余内存不断分配给了缓冲区
        #一段时间之后，剩余内存很小，而缓冲区占用了大部分内存。此时Swap使用之间增大，缓冲区和剩余内存只在小范围波动
        
        停下sar命令
        cachetop5 #观察缓存
        #可以看到dd进程读写只有50%的命中率，未命中数为4w+页，说明正式dd进程导致缓冲区使用升高
        watch -d grep -A 15 ‘Normal’ /proc/zoneinfo #观察内存指标变化
        #发现升级内存在一个小范围不停的波动，低于页低阈值时会突然增大到一个大于页高阈值的值
        ```
        - 说明剩余内存和缓冲区的波动变化正是由于内存回收和缓存再次分配的循环往复。有时候Swap用的多，有时候缓冲区波动更多。此时查看swappiness值为60，是一个相对中和的配置，系统会根据实际运行情况来选去合适的回收类型.
    - 如何“快准狠”找到系统内存存在的问题
      - ![img.png](os_memory_metric_tool.png)
      - ![img.png](os_memory_metric_tool2.png)
    - 如何迅速分析内存的性能瓶颈
      - 通常先运行几个覆盖面比较大的性能工具，如free，top，vmstat，pidstat等
        - 先用free和top查看系统整体内存使用情况
        - 再用vmstat和pidstat，查看一段时间的趋势，从而判断内存问题的类型
        - 最后进行详细分析，比如内存分配分析，缓存/缓冲区分析，具体进程的内存使用分析等
      - 常见的优化思路：
         - 最好禁止Swap，若必须开启则尽量降低swappiness的值
         - 减少内存的动态分配，如可以用内存池，HugePage等
         - 尽量使用缓存和缓冲区来访问数据。如用堆栈明确声明内存空间来存储需要缓存的数据，或者用Redis外部缓存组件来优化数据的访问
         - cgroups等方式来限制进程的内存使用情况，确保系统内存不被异常进程耗尽
         - /proc/pid/oom_adj调整核心应用的oom_score，保证即使内存紧张核心应用也不会被OOM杀死
  - [Tools](https://www.ctq6.cn/linux%E6%80%A7%E8%83%BD%E4%BC%98%E5%8C%96/)
    - vmstat使用详解
      - vmstat命令是最常见的Linux/Unix监控工具，可以展现给定时间间隔的服务器的状态值,包括服务器的CPU使用率，内存使用，虚拟内存交换情况,IO读写情况。
       ```shell
       # 结果说明
       - r 表示运行队列(就是说多少个进程真的分配到CPU)，我测试的服务器目前CPU比较空闲，没什么程序在跑，当这个值超过了CPU数目，就会出现CPU瓶颈了。这个也和top的负载有关系，一般负载超过了3就比较高，超过了5就高，超过了10就不正常了，服务器的状态很危险。top的负载类似每秒的运行队列。如果运行队列过大，表示你的CPU很繁忙，一般会造成CPU使用率很高。
       - b 表示阻塞的进程,这个不多说，进程阻塞，大家懂的。
       - swpd 虚拟内存已使用的大小，如果大于0，表示你的机器物理内存不足了，如果不是程序内存泄露的原因，那么你该升级内存了或者把耗内存的任务迁移到其他机器。
       
       - free   空闲的物理内存的大小，我的机器内存总共8G，剩余3415M。
       
       - buff   Linux/Unix系统是用来存储，目录里面有什么内容，权限等的缓存，我本机大概占用300多M
       
       - cache cache直接用来记忆我们打开的文件,给文件做缓冲，我本机大概占用300多M(这里是Linux/Unix的聪明之处，把空闲的物理内存的一部分拿来做文件和目录的缓存，是为了提高 程序执行的性能，当程序使用内存时，buffer/cached会很快地被使用。)
       
       - si  每秒从磁盘读入虚拟内存的大小，如果这个值大于0，表示物理内存不够用或者内存泄露了，要查找耗内存进程解决掉。我的机器内存充裕，一切正常。
       
       - so  每秒虚拟内存写入磁盘的大小，如果这个值大于0，同上。
       
       - bi  块设备每秒接收的块数量，这里的块设备是指系统上所有的磁盘和其他块设备，默认块大小是1024byte，我本机上没什么IO操作，所以一直是0，但是我曾在处理拷贝大量数据(2-3T)的机器上看过可以达到140000/s，磁盘写入速度差不多140M每秒
       
       - bo 块设备每秒发送的块数量，例如我们读取文件，bo就要大于0。bi和bo一般都要接近0，不然就是IO过于频繁，需要调整。
       
       - in 每秒CPU的中断次数，包括时间中断
       
       - cs 每秒上下文切换次数，例如我们调用系统函数，就要进行上下文切换，线程的切换，也要进程上下文切换，这个值要越小越好，太大了，要考虑调低线程或者进程的数目,例如在apache和nginx这种web服务器中，我们一般做性能测试时会进行几千并发甚至几万并发的测试，选择web服务器的进程可以由进程或者线程的峰值一直下调，压测，直到cs到一个比较小的值，这个进程和线程数就是比较合适的值了。系统调用也是，每次调用系统函数，我们的代码就会进入内核空间，导致上下文切换，这个是很耗资源，也要尽量避免频繁调用系统函数。上下文切换次数过多表示你的CPU大部分浪费在上下文切换，导致CPU干正经事的时间少了，CPU没有充分利用，是不可取的。
       
       - us 用户CPU时间，我曾经在一个做加密解密很频繁的服务器上，可以看到us接近100,r运行队列达到80(机器在做压力测试，性能表现不佳)。
       
       - sy 系统CPU时间，如果太高，表示系统调用时间长，例如是IO操作频繁。
       
       - id 空闲CPU时间，一般来说，id + us + sy = 100,一般我认为id是空闲CPU使用率，us是用户CPU使用率，sy是系统CPU使用率。
       
       - wt 等待IO CPU时间
       ```
    - pidstat 使用详解
      - pidstat主要用于监控全部或指定进程占用系统资源的情况,如CPU,内存、设备IO、任务切换、线程等。使用方法：
        ````shell
        pidstat –d interval times 统计各个进程的IO使用情况
        pidstat –u interval times 统计各个进程的CPU统计信息
        pidstat –r interval times 统计各个进程的内存使用信息
        pidstat -w interval times 统计各个进程的上下文切换
        ````








