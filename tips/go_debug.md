
- [PProf 工具](https://mp.weixin.qq.com/s?__biz=MzUxMDI4MDc1NA==&mid=2247488702&idx=1&sn=b941ddb5473e8f6b85cd970e81225347&chksm=f90401e3ce7388f50f390eb4dfd887481a7866cb50011802d1916ec644c3ba5485ea0e423036&scene=178&cur_album_id=1628210764123521025#rd)
  - PProf 是用于可视化和分析性能分析数据的工具，PProf 以 profile.proto 读取分析样本的集合，并生成报告以可视化并帮助分析数据
  - 哪几种采样方式
    - runtime/pprof：采集程序（非 Server）的指定区块的运行数据进行分析。
    - net/http/pprof：基于HTTP Server运行，并且可以采集运行时数据进行分析。
    - go test：通过运行测试用例，并指定所需标识来进行采集
  - 可以做什么
    - CPU Profiling：CPU 分析，按照一定的频率采集所监听的应用程序 CPU（含寄存器）的使用情况，可确定应用程序在主动消耗 CPU 周期时花费时间的位置。
    - Memory Profiling：内存分析，在应用程序进行堆分配时记录堆栈跟踪，用于监视当前和历史内存使用情况，以及检查内存泄漏。
    - Block Profiling：阻塞分析，记录Goroutine阻塞等待同步（包括定时器通道）的位置，默认不开启，需要调用runtime.SetBlockProfileRate进行设置。
    - Mutex Profiling：互斥锁分析，报告互斥锁的竞争情况，默认不开启，需要调用runtime.SetMutexProfileFraction进行设置。
    - Goroutine Profiling：Goroutine 分析，可以对当前应用程序正在运行的 Goroutine 进行堆栈跟踪和分析。
- [跟踪剖析 trace](https://mp.weixin.qq.com/s/OY-w05uJIgjov9qGmJ-Wwg)
  - 有时候单单使用 pprof 还不一定足够完整观查并解决问题，因为在真实的程序中还包含许多的隐藏动作，例如：
    - Goroutine 在执行时会做哪些操作？
    - Goroutine 执行/阻塞了多长时间？
    - Syscall 在什么时候被阻止？在哪里被阻止的？
    - 谁又锁/解锁了 Goroutine ？
    - GC 是怎么影响到 Goroutine 的执行的？
  - 功能
    - View trace：查看跟踪
    - Goroutine analysis：Goroutine 分析
    - Network blocking profile：网络阻塞概况
    - Synchronization blocking profile：同步阻塞概况
    - Syscall blocking profile：系统调用阻塞概况
    - Scheduler latency profile：调度延迟概况
    - User defined tasks：用户自定义任务
    - User defined regions：用户自定义区域
    - Minimum mutator utilization：最低 Mutator 利用率
- [性能调优利器--火焰图](https://zhuanlan.zhihu.com/p/147875569)
  - 火焰图类型
  ![img.png](go_debug_frame.png)
  - 如何绘制火焰图
    - perf 相对更常用，多数 Linux 都包含了 perf 这个工具，可以直接使用；
    - SystemTap 则功能更为强大，监控也更为灵活
      - SystemTap 是动态追踪工具，它通过探针机制，来采集内核或者应用程序的运行信息，从而可以不用修改内核和应用程序的代码
      - SystemTap 定义了一种类似的 DSL 脚本语言，方便用户根据需要自由扩展
  - [Blazing Performance with Flame Graphs](https://www.usenix.org/conference/lisa13/technical-sessions/plenary/gregg)
- [如何使用 Kubernetes 监测定位慢调用](https://mp.weixin.qq.com/s/mOdn5eE0QtLfHuotpgacwg)
  - 定位慢调用一般来说有什么样的步骤 - 黄金信号 + 资源指标 + 全局架构
    - 黄金信号
      - 延时--用来描述系统执行请求花费的时间。常见指标包括平均响应时间，P90/P95/P99 这些分位数，这些指标能够很好的表征这个系统对外响应的快还是慢，是比较直观的。
      - 流量--用来表征服务繁忙程度，典型的指标有 QPS、TPS。
      - 错误--也就是我们常见的类似于协议里 HTTP 协议里面的 500、400 这些，通常如果错误很多的话，说明可能已经出现问题了。
      - 饱和度--就是资源水位，通常来说接近饱和的服务比较容易出现问题，比如说磁盘满了，导致日志没办法写入，进而导致服务响应。典型的那些资源有 CPU、 内存、磁盘、队列长度、连接数等等。
    - 资源指标 - 对于每一个资源去检查 utilization（使用率），saturation （饱和度），error（错误） ，合起来就是 USE 了
  - Case
    - 网络性能差
      - 指标 - 速率跟带宽，第二个是吞吐量，第三个是延时，第四个是 RTT。
- [是什么影响了我的接口延迟](https://mp.weixin.qq.com/s/k69-rs64XSkOFOpvUwq9sw)
  - 接口延迟大幅上升时
    - 先去看看 pprof 里的 goroutine 页面，看看 goroutine 是不是阻塞在什么地方了(比如锁)
  - USE 方法论，其中提到了一个 Saturation (饱和度)的概念，这个和 Utilization 有啥区别
    - Util 一般指的是繁忙程度，繁忙程度指的是你的资源有多少正在被利用
    - Sat 一般指的是饱和程度，而饱和度则指的是等待利用这些资源的队列有多长。
    - 如果只有一个核，那么我们就可以通过 util 和 sat 指标推断出这样的结论：sat 越高，接口延迟越高。util 高，影响不是特别大。
    - 现代 CPU 支持超线程(hyper thread)，你可以理解成一个窗口要排两个队，所以有时 CPU 的总 util 过了 50%，API 的延迟就比较高了
  - Case
    - 在 Go 的服务中，阻塞的 goroutine 数量变多，本质上还是这些 goroutine 发生了排队，了解底层的读者应该一想就知道 goroutine 是在哪里排队了。所以 goroutine 数量越多，说明队列也越拥挤
    - 网络应用中的 send buffer，receiver buffer 本质上也是队列
    - CPU 调度器本身也有执行队列，可以用 bcc 中的 runqlen 工具来查看
    - 磁盘的读写也有相应的队列
  - [Controlling Queue Delay](https://queue.acm.org/detail.cfm?id=2209336)
- [pprof快速定位Go程序内存泄露](https://mp.weixin.qq.com/s/PEpvCqpi9TPhVuPdn3nyAg)
- [Analyze Current Goroutines in Go](https://trstringer.com/analyze-goroutines/)
- [Advent of Go Profiling](https://felixge.de/2021/12/01/advent-of-go-profiling-2021-day-1-1/)







