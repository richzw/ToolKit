
- [Prometheus 数据存储]()
  - tsdb 数据
    - 基于相对稳定频率持续产生的一系列指标监测数据，那么存储就是一些标签键加上一个时间序列作为一个大 key，值就是一个数字
  - 对于数据的存储 Prometheus 按冷热数据进行分离，最近的数据肯定是看的最多的，所以缓存在内存里面，为了防止宕机而导致数据丢失因而引入 wal 来做故障恢复
  - 数据超过一定量之后会从内存里面剥离出来以 chunk 的形式存放在磁盘上这就是 head chunk。
  - 对于更早的数据会进行压缩持久化变成 block 存放到磁盘中。
  - 对于 block 中的数据由于是不会变的，数据较为固定，所以每个 block 通过 index 来索引其中的数据，并且为了加快数据的查询引用倒排索引，便于快速定位到对应的 chunk。
- [Prometheus TSDB](https://ganeshvernekar.com/blog/prometheus-tsdb-the-head-block/)
- [时序数据高基问题](https://mp.weixin.qq.com/s/baTvpUXuA594JDUVU0wL5Q)
  - 第一个有效的解法是垂直切分，大部分业界主流时序数据库或多或少都采用了类似方法，按照时间来切分索引，因为如果不做这个切分的话，随着时间的推进，索引会越来越膨胀，最后到内存放不下，如果按照时间切分，可以把旧的 index chunk 交换到磁盘甚至远程存储，起码写入是不会被影响到了。
  - 与垂直切分相对的，就是水平切分，用一个 sharding key，一般可以是查询谓词使用频率最高的一个或者几个 tag，按照这些 tag 的 value 来进行 range 或者 hash 切分，这样就相当于使用分布式的分而治之思想解决了单机上的瓶颈，代价就是如果查询条件不带 sharding key 的话通常是无法将算子下推，只能把数据捞到最上层去计算。
- [监控降噪](https://mp.weixin.qq.com/s/rEn25SejnU0rOWFb39QuUw)
  - 如何衡量监控效果
    - 衡量效果最有效的指标为召回率——即衡量能够正确识别出正样本的百分比。召回率的计算公式为：召回率 = 正确识别的正样本数 / 所有正样本数
    - 在保证召回率，识别线上问题的同时，提升准确率降低噪音，便是监控治理要做的事。
  - 监控规则
    - 避免维度单一
    - 利用黑白名单
    - 利用环比和同比
- [Sidecar 的资源和性能管理]
```shell
record: "container_cpu_usage_against_request:pod:rate1m"
   expr: |
    (   
      count(kube_pod_container_resource_requests{resource="cpu", container!=""}) by (container, pod, namespace)
      *   
      avg(
        irate(
          container_cpu_usage_seconds_total{container!=""}[1m]
        )   
      ) by (container, pod, namespace)
    )   
    /   
    avg(
      avg_over_time(
        kube_pod_container_resource_requests{resource="cpu", container!=""}[1m]
      )   
    ) by (container, pod, namespace) * 100 
    *   
    on(pod) group_left(workload) (
      avg by (pod, workload) (
        label_replace(kube_pod_info{created_by_kind=~"ReplicaSet|Job"}, "workload", "$1", "created_by_name", "^(.*)-([^-]+)$")
        or  
        label_replace(kube_pod_info{created_by_kind=~"DaemonSet|StatefulSet"}, "workload", "$1", "created_by_name", "(.*)")
        or  
        label_replace(kube_pod_info{created_by_kind="Node"}, "workload", "node", "", "") 
        or  
        label_replace(kube_pod_info{created_by_kind=""}, "workload", "none", "", "") 
      )   
    )
```
- [Metrics 系统架构演进](https://mp.weixin.qq.com/s/ezG3VQLgE2e0AWSxsoBHRg)
  -  Thanos
    - 可以从多个 Prometheus 集群查询数据，统一了查询入口，提高了用户的体验。同时提供长期数据，另外 Thanos 可以通过 Prometheus-Operator 来管理，所以大大降低了整体管理成本和入侵性
  - 优化：
    - 升级了 Thanos 的版本，为 query-frontend 和 storegateway 服务增加了 Redis 缓存，从而提升查询的性能。
    - 为 store gateway 做了基于时间的分片
  - 面临以下几个问题：
    - 超 100+ 倍数据点增长导致查询缓慢
    - 架构复杂，参数调优困难
    - 频繁 OOM
  - VictoriaMetrics 
    - 根据容器可用的 CPU 数量计算协程数量
    - 区分 IO 协程和计算协程，同时提供了协程优先级策略
    - 使用 ZSTD 压缩传输内容降低磁盘性能要求
    - 根据可用物理内存限制对象的总量，避免 OOM
    - 区分 fast path 和 slow path，优化 fast path 避免 GC 压力过大
- metric 
  - database
    - PostgreSQL 是多进程模式，所以需要十分关注链接数和页表大小，虽然使用 Hugepage 方案可以降低页表的负担，但是 Hugepage 本身还是有比较多的副作用，利用 pgBouncer 之类的 proxy 做链接复用是一种更好的解法；
      - 当开启 full page 时，PostgreSQL 对 I/O 带宽的需求非常强烈，此时的瓶颈为 I/O 带宽；当 I/O 和链接数都不是瓶颈时，PostgreSQL 在更高的并发下瓶颈来自内部的锁实现机制。
    - MongoDB 整体表现比较稳定，主要的问题一般来自 Disk I/O 和链接数，WiredTiger 在 cache 到 I/O 的流控上做得比较出色，虽然有 I/O 争抢，但是 IO hang 的概率比较小，
      - 当然 OLTP 数据库的 workload 会比 MongoDB 更复杂一些，也更难达到一种均衡。
    - Redis 的瓶颈主要在网络，所以需要特别关注应用和 Redis 服务之间的网络延迟，这部分延迟由网络链路决定，
      - Redis 满载时 70%+ 的 CPU 消耗在网络栈上，所以为了解决网络性能的扩展性问题，Redis 6.0 版本引入了网络多线程功能，真正的 worker thread 还是单线程，这个功能在大幅提升 Redis 性能的同时也保持了 Redis 简单优雅的特性。
  - Pod
    - CPU
      - `sum(rate(container_cpu_usage_seconds_total{namespace=~"alpha|beta|prod", image!="", container_name!="POD", pod=~".*$project.*"}[1m])) by (pod) /
         sum(container_spec_cpu_quota{namespace=~"alpha|beta|prod", image!="", container_name!="POD", pod=~".*$project.*"}/container_spec_cpu_period{namespace=~"alpha|beta|prod", image!="", container_name!="POD", pod=~".*$project.*"}) by (pod) * 100`
    - Memory
      - `avg(container_memory_working_set_bytes{namespace=~"alpha|beta|prod", pod=~".*$project.*", container!="POD"} > 0) by (pod) /
        avg(kube_pod_container_resource_requests_memory_bytes{namespace=~"alpha|beta|prod", pod=~".*$project.*", container!="POD"} > 0) by (pod) * 100`
    - Network
    - 
- [Prometheus 指标值为何不准](https://mp.weixin.qq.com/s/A3W3hSCpQi8DQYJxOS1ZGA)
  - Overview
    - Prometheus 指标值不准的“怪现象”，其实是在下面的“不可能三角”中，做出了取舍——为保全效率和可用性，舍弃了精度
    - 其手段通常是对原始数据先采样、再聚合，利用有限的信息，分析变化趋势
    - Prometheus 毕竟处在一个条件有限的真实世界，它还要随时面临以下困难: 自身硬件有限, 采样统计的局限性, 分布式的局限性
    - 交出 not perfect、但是 good enough 的指标。于是就有了下述设计：
      - 单次采样不重要，多次采样组成的时间序列才重要。所以，单次采样受阻，是可以无需重试、直接丢弃的。
      - 单点数值不重要，多点数值汇聚的变化趋势才重要。所以，单点数值是可以“无中生有”、"脑补"估算的。
  - Case
    - 失真的 rate/increase
      - 在使用 rate 或者 increase 观测 counter 类型的指标增量时，经常碰到
        - 每分钟新增的请求数，竟然是个小数？ ，不仅是个小数，还比真实增量更大？
      - 最常见的原因，就是线性外推（linear extrapolation）
        - 线性外推算法：取窗口覆盖范围内的第一个点和最后一个点，计算斜率，并按照该斜率将直线延伸至窗口边界，无中生有地“脑补”出虚拟的两个“样本点”，即可相减计算 increase 了：
      - rate/increase[时间范围] 在计算该时间范围内的增量时，第一步要拿到该时间范围边界上（开始时刻和结束时刻）的样本点，相减得到差值
      - Prometheus 的选择是：naive 地假设所有样本点在该时间范围内是均匀分布的，然后按照这个均匀分布的线性规律，“脑补”估算出边界上的采样点
      - 要计算 [1m] 的时间范围/取样窗口内的 increase，在最理想的情况下，Prometheus 根本不想关心这个窗口内的其他数据，而只需从窗口左边界取第一个点，右边界取最后一个点，相减即可：
    - 离谱的 histogram
      - histogram 百分位（percentile）不准，这是为啥呢？这就不得不提线性插值（linear interpolation） 了
      - PTS 搜集了响应时间的平均值、P50、P90、P95——但就是没有 P99
      - 所求分位值 = bucket 段左边界值 + (bucket 段右边界值 - bucket 段左边界值) * (目标样本在本 bucket 段的排行 / 本 bucket 段的样本总数)
      - 若想用 histogram 获得较为准确的分位值，则需对样本分布有一定的了解，再根据这个分布，设置合理的 bucket 边界
    - 薛定谔的 range
      - 以上述 rate(errors_total[时间范围]) 为例，若我们分别选时间范围 [30s]、[1m]、[5m]，看一眼三者的 Grafana 图表，这不能说一模一样，只能说是毫不相关：随着时间范围扩大，主打一个逐渐平滑、失去尖峰
      - 曲线随 rate 窗口而峰值和形态大变的原因：
        - 窗口小则更加敏感，能够捕捉到更短时间内的变化。这意味着如果有突发事件或者短期波动，它会在曲线上表现得更明显。
        - 窗口大会更加平滑，因为它平均了更长时间内的数据。这样可以减少短期波动的影响，但也可能掩盖掉短时间内的突发事件。
      - 在选择合适的时间范围时，应考虑以下因素：
        - 指标的特性：对于波动较大的指标，可能需要一个较短的时间范围来快速发现问题。对于相对平稳的指标，较长的时间范围可以提供更清晰的趋势。
        - 监控目标：如果你需要实时监控和快速响应，短时间范围可能更合适。如果关注长期趋势，那么长时间范围会更有帮助。
        - Prometheus 抓取间隔：时间范围应该至少是 Prometheus 抓取间隔的两倍，这样才能确保有足够的数据点来计算速率。
    - 在一个分布式的世界，网络抖动、对端延迟等引起的数据丢失问题，会给本就不精确的 Prometheus 指标值雪上加霜
      - 虽则 rate 计算斜率需要至少两个点，但最佳实践建议将 rate 的时间范围至少设为 Prometheus scrape interval（抓取周期/间隔）的 4 倍
      - 网络抖动可能导致丢点，也可能导致点的延迟。那么当延迟的点到达时，它就出现在了本不属于它的统计周期内。这可能导致 rate 出现波动
    - 对 Prometheus 使用范围查询（range query），就必然涉及 step（步长）
      - Grafana 需要渲染整条曲线，可以理解为 Grafana 在时间轴上按 step 每走一步，就要做一次查询/evaluation，得到一个值，生成曲线上的一个
      - 当 step 的步长，叠加 Prometheus scrape interval，再叠加 PromQL 里的 range 时间范围窗口
  - Summary
    - 在一个分布式的世界，网络抖动、对端延迟等引起的数据丢失问题，会给本就不精确的 Prometheus 指标值雪上加霜。
      - 例如：虽则 rate 计算斜率需要至少两个点，但最佳实践建议将 rate 的时间范围至少设为 Prometheus scrape interval（抓取周期/间隔）的 4 倍。这将确保即使抓取速度缓慢、且发生了一次抓取故障，也始终可以使用两个样本。
      - 再例如：网络抖动可能导致丢点，也可能导致点的延迟。那么当延迟的点到达时，它就出现在了本不属于它的统计周期内。这可能导致 rate 出现波动，尤其是在监控较短时间范围的 rate 时。
    - 文章里只关注了对 PromQL 的一次查询/evaluation。而在现实中对 Prometheus 使用范围查询（range query），就必然涉及 step（步长）。
      - 比如 Grafana 需要渲染整条曲线，可以理解为 Grafana 在时间轴上按 step 每走一步，就要做一次查询/evaluation，得到一个值，生成曲线上的一个点。那么当 step 的步长，叠加 Prometheus scrape interval，再叠加 PromQL 里的 range 时间范围窗口……可以设想，这几个参数不同的排列组合，会导致曲线更加充满惊喜意外……
    - Prometheus 的增量外推（extrapolation），其实也不是纯粹地无脑外推；它有时还会考虑到距离窗口边界的距离，而做一些其他微调。
    - 本文未涉及 Prometheus counter 重置（reset）对 increase/rate 准确度的影响。也即：counter 如遇归零（如服务器重启导致），Prometheus 会有应对机制自动来处理，正常情况下不用担心。但若好巧不巧，数据点存在乱序，则可能因为数值下降而误触 Prometheus 重置后的补偿机制，被“脑补”计算出一个极大的异常 increase/rate。
- [Flame graph AI](https://grafana.com/docs/grafana-cloud/monitor-applications/profiles/flamegraph-ai/?pg=blog&plcmt=body-txt)
  - Performance bottlenecks: What’s causing the slowdown?
  - Root causes: Why is it happening?
  - Recommended fixes: How would you resolve it?














