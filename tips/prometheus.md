
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







