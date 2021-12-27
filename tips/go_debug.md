
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








