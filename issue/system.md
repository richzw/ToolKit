
- [Redis请求毛刺](https://mp.weixin.qq.com/s/8l2Qf2vozhCcb9AvvJSBYA)
  
  步骤
  - 首先排查是不是网络问题，查一段时间的 redis slowlog（slowlog 最直接简单）
  - 本地抓包，看日志中 redis 的 get key 网络耗时跟日志的时间是否对的上
  - 查机器负载，是否对的上毛刺时间（弹性云机器，宿主机情况比较复杂）
  - 查 redis sdk，看源码，看实时栈，看是否有阻塞（sdk 用了pool，pool 逻辑是否可能造成阻塞）
    - 源码追踪
    - trace实时栈，看是否有 lock，wait 之类的逻辑
  - 查看 runtime 监控，看是否有协程暴增，看 gc stw 时间是否影响 redis（go 版本有点低，同时内存占用大）；
    - 抓了下线上heap 图，查看历史的gc stw 信息:
      `curl http://localhost:8003/debug/pprof/heap?debug=1`
  - trace ，看调度时间和调度时机是否有问题（并发协程数，GOMAXPROCS cpu负载都会影响调度）
    ```shell
    curl http://127.0.0.1:8080/debug/pprof/trace?seconds=300 > trace.out
    go tool trace trace.out
    ```
    查看goroutine analysis
    - 按 scheduler wait 排序后数据，影响调度的，主要是协程数量和线程数量
    - 线程数查看 `ps -T -p pid`，线程数是200+
    - 将 GOMAXPROC 设置成8，然后重新上线。然后抖动立刻降下来了
    
    原因： 怎么解决获取正确核数的问题？
    - 是设置环境变量 GOMAXPROCS 启动
    - 显式调用 uber/automaxprocs。

- Request无响应
  - 网络正常
    
  - 问题触发：进程要向容器标准输出打印日志 容器引擎重启
    

- 95% percentile request < 5 ms
  - 存储：
    
    问题：一次调用一个用户的三百个特征原方案是用 redis hash 做表，每个 field 为用户的一个特征。由于用户单个请求会获取几百个特征，即使用hmget做合并，存储也需要去多个 slot 中获取数据，效率较低
    
    改进1：把 hash 表的所有 filed 打包成一个 json 格式的 string
    
    改进的问题：若 hash filed 过多，string 的 value 值会很大，（redis 大key）
    
    改进2：
      - 按照类型将特征做细分，比如原来一个 string 里面有 300 的字段，拆分成 3 个有 100 个值的 string 类型。 
      - 对 string val 进行压缩
      - 再加一层cache
  - 代码
    
    分析：pprof 可以看 CPU、内存、协程等信息在压测流量进来时系统调用的各部分耗时情况。而 trace 可以查看 runtime 的情况

    改进：
      - 采用优化的json库
      - string <--> []byte 优化
      - json.Unmarshal 的结果cache，防止多次冗余unmarshal
      - prealloc， slice 与map 预分配大小
    
- [GC pause over 100ms排查](https://mp.weixin.qq.com/s/Lk1EbiT7WprVOyX_dXYMyg)
  
  - 复现：用ab 50并发构造些请求看看. 网络来回延时60ms, 但是平均处理耗时200多ms, 99%耗时到了679ms
  - GC以及trace： 该进程的runtime信息, 发现内存很少，gc-pause很大，GOMAXPROCS为76，是机器的核数
    `export GODEBUG=gctrace=1`, 重启进程看看. 可以看出gc停顿的确很严重
    ```shell
        curl -o trace.out 'http://ip:port/debug/pprof/trace?seconds=20'
        sz ./trace.out
    ```
  - 原因： 容器中Go进程没有正确的设置GOMAXPROCS的个数, 导致可运行的线程过多, 可能出现调度延迟的问题. 正好出现进入gc发起stw的线程把其他线程停止后, 其被调度器切换出去, 很久没有调度该线程, 实质上造成了stw时间变得很长  
  - Solution: go.uber.org/automaxprocs, 容器中go进程启动时, 会正确设置GOMAXPROCS. 
  - 总结
    - 容器中进程看到的核数为母机CPU核数，一般这个值比较大>32, 导致go进程把P设置成较大的数，开启了很多P及线程
    - 一般容器的quota都不大，0.5-4，linux调度器以该容器为一个组，里面的线程的调度是公平，且每个可运行的线程会保证一定的运行时间，因为线程多, 配额小, 虽然请求量很小, 但上下文切换多, 也可能导致发起stw的线程的调度延迟, 引起stw时间升到100ms的级别，极大的影响了请求
    - 通过使用automaxprocs库, 可根据分配给容器的cpu quota, 正确设置GOMAXPROCS以及P的数量, 减少线程数，使得GC停顿稳定在<1ms了. 且同等CPU消耗情况下, QPS可增大一倍，平均响应时间由200ms减少到100ms. 线程上下文切换减少为原来的1/6
    - 同时还简单分析了该库的原理. 找到容器的cgroup目录, 计算cpuacct,cpu下cpu.cfs_quota_us/cpu.cfs_period_us, 即为分配的cpu核数.





