
- [估算的一些方法](https://mp.weixin.qq.com/s/fH-AJpE99ulSLbC_1jxlqw)
  - 每个程序员都要了解的延迟数字
  ![img.png](sd_delay_number.png)
  - 一些数字积累
    - 某支付服务的支付峰值 60w QPS
    - Go GC 打开写屏障需要花费 10-30us
    - 内网中，一次网络请求的耗时是 ms 级别
    - 万兆网卡，1.25GB/s 打满
    - 4C8G 建 10w 到 20w 的连接没有问题
    - 因为机械硬盘的机械结构，随机 I/O 与顺序的 I/O 性能可能相差几百倍。固态硬盘则只有十几倍到几十倍之间
    - twitter 工程师认为，良好体验的网站平均响应时间应该在 500ms 左右，理想的时间是 200-300ms
    - 平均 QPS：日平均用户请求除以 4w。日平均用户请求，一般来自产品的评估。峰值 QPS：平均 QPS 的 2~4 倍
  - 本章最后有一个实战的例子：评估 twitter 的 QPS 和存储容量。
    - 先给出了一些预设：
      - 300 个 million 的月活跃用户
      - 50% 的用户每天都使用 twitter
      - 用户平均每天发表 2 条 tweets
      - 10% 的 tweets 包含多媒体
      - 多媒体数据保存 5 年
    - 下面是估算的过程：
      - 先预估 QPS：
        - DAU（每天的活跃用户数，Daily Active Users）为：300 million（总用户数） * 50% = 150 million
        - 发 tweets 的平均 QPS：150 million * 2 / 24 hour / 3600 second = ~3500
        - 高峰期 QPS 一般认为是平均 QPS 的 2 倍：2 * 3500 = 7000 QPS
      - 再来估算存储容量：
        - 假设多媒体的平均大小为 1MB，那么每天的存储容量为：150 million * 2 * 10% * 1MB = 30 TB。5 年的存储容量为 30 TB * 365 * 5 = 55 PB。
      - 最后这两个的估算过程是这样的：
        - 300 个 million * 10%* 1MB，1 MB 其实就是 6 个 0，相当于 million 要进化 2 次：million -> billion -> trillion，即从 M -> G -> T，于是结果等于 300 T * 10% = 30 T。
        - 30 TB * 365 * 5 = 30 TB * 1825 = 30 TB * 10^3 * 1.825，TB 进化一次变成 PB，于是等于 30 * 1.825 PB = 55 PB。
- [Go map[int64]int64 写入 redis 占用多少内存](https://mp.weixin.qq.com/s?__biz=MjM5MDUwNTQwMQ==&mid=2257487941&idx=1&sn=80cee0d0f88d73f57a25c496eef90393&scene=21#wechat_redirect)
  - 将内存中的一个超大的 map[int64]int64 写入到 redis，map 里的元素个数是千万级的。设计方案的时候，需要对 redis 的容量做一个估算。
  - 错误的示例
    - 如果不了解 redis 的话，可能你的答案是用元素个数直接乘以 16B（key 和 value 各占 8B）。我们假设元素个数是 5kw，那估算结果就是：5kw * 16B=50kk * 16B = 800MB
  - redis说起
    - Redis 中的一个 k-v 对用一个 entry 项表示，其中每个 entry 包含 key、value、next 三个指针，共 24 字节。由于 redis 使用 jemalloc 分配内存，因此一个 entry 需要申请 32 字节的内存。这里的 key, value 指针分别指向一个 RedisObject
       ```c++
       typedef struct redisObject {
           unsigned type:4;
           unsigned encoding:4;
           unsigned lru:LRU_BITS; 
           int refcount;
           void *ptr;
       } robj;
       ```
    - RedisObject 对应前面提到的各种数据类型，其中最简单的就是 redis 内部的字符串了。它有如下几种编码格式
      ![img.png](system_design_redis.png)
    - 当字符串是一个整型时，直接放在 ptr 位置，不用再分配新的内存了，非常高效
    - 我们要写入 redis 的 map 中的 key 和 value 都是整数，因此直接将值写入 ptr 处即可。
      于是 map 的一个 key 占用的内存大小为：32（entry）+16（key）+16（value）=64B。于是，5kw 个 key 占用的内存大小是 5kw*64B = 50 kk * 64B = 3200MB ≈ 3G
    - 假如我们在 key 前面加上了前缀，那就会生成 SDS，占用的内存会变大，访问效率也会变差。
- [高并发系统建设经验总结](https://mp.weixin.qq.com/s/TTn3YNwKKWn5IS8F6HJHIg)
  - 基础设施
    - 异地多活
  - Database
    - 读写分离; 大部分业务特点是读多写少，因此使用读写分离架构可以有效降低数据库的负载，提升系统容量和稳定性
      - 缺点也是同样明显的
        - 主从延迟
        - 从库的数量是有限的
        - 无法解决 TPS 高的问题
    - 分库分表; 当读写分离不能满足业务需要时，就需要考虑使用分库分表模式了。当确定要对数据库做优化时，应该优先考虑使用读写分离的模式，只有在读写分离的模式已经没办法承受业务的流量时，我们才考虑分库分表的模式
    - 由于是多 master 的架构，分库分表除了包含读写分离模式的所有优点外，还可以解决读写分离架构中无法解决的 TPS 过高的问题，同时分库分表理论上是可以无限横向扩展的，也解决了读写分离架构下从库数量有限的问题。
    - 缺点
      - 改造成本高
      - 事务问题
        - 在分库分表后应该要尽量避免这种跨 DB 实例的操作，如果一定要这么使用，优先考虑使用补偿等方式保证数据最终一致性，如果一定要强一致性，常用的方案是通过分布式事务的方式。
      - 无法支持多维度查询
        - 第一种是引入一张索引表，这张索引表是没有分库分表的，还是以按用户 ID 分库分表为例，索引表上记录各种维度与用户 ID 之间的映射关系，请求需要先通过其他维度查询索引表得到用户 ID，再通过用户 ID 查询分库分表后的表。
        - 通过引入NoSQL的方式，比较常见的组合是ES+MySQL，或者HBase+MySQL的组合等，这种方案本质上还是通过 NoSQL 来充当第一种方案中的索引表的角色
      - 数据迁移
        - 停机迁移
        - 双写，这主要是针对新增的增量数据，存量数据可以直接进行数据同步，关于如何进行双写迁移网上已经有很多分享了，这里也就不赘述，核心思想是同时写老库和新库
  - 架构
    - 缓存
      - [如何保证缓存与数据库的数据一致性](https://coolshell.cn/articles/17416.html)
        - Write through
          ```shell
          lock(运单ID) {
           //...
           
              // 删除缓存
             deleteCache();
              // 更新DB
             updateDB();
              // 重建缓存
             reloadCache()
          }
          ```
          防止并发问题，写请求都需要加分布式锁，锁的粒度是以运单 ID 为 key，在执行完业务逻辑后，先删除缓存，再更新 DB，最后再重建缓存，这些操作都是同步进行的，在读请求中先查询缓存，如果缓存命中则直接返回，如果缓存不命中则查询 DB，然后直接返回，也就是说在读请求中不会操作缓存，这种方式把缓存操作都收敛在写请求中，且写请求是加锁的，有效防止了读写并发导致的写入脏缓存数据的问题。
      - 缓存要避免大 key 和热 key 的问题
      - 读写性能
        - 写性能，影响写性能的主要因素是 key/value 的数据大小，比较简单的场景可以使用JSON的序列化方式存储，但是在高并发场景下使用 JSON 不能很好的满足性能要求，而且也比较占存储空间，比较常见的替代方案有protobuf、thrift等等
        - 读性能的主要影响因素是每次读取的数据包的大小。在实践中推荐使用redis pipeline+批量操作的方式
      - 适当冗余
        - 适当冗余的意思是说我们在设计对外的业务查询接口的时候，可以适当的做一些冗余。
        - 我们一开始设计对外查询接口的时候能做一些适当的冗余，区分不同的业务场景，虽然这样势必会造成有些接口的功能是类似的，但在加缓存的时候就能有的放矢，针对不同的业务场景设计不同的方案，比如关键的流程要注重数据一种的保证，而非关键场景则允许数据短暂的不一致来降低缓存实现的成本
    - 消息队列
      - 在高并发系统的架构中，消息队列（MQ）是必不可少的，当大流量来临时，我们通过消息队列的异步处理和削峰填谷的特性来增加系统的伸缩性，防止大流量打垮系统，此外，使用消息队列还能使系统间达到充分解耦的目的。
    - 服务治理
      - 超时
      - 熔断，限流
      - 降级
      - 注册 发现
      - 安全
      - 监控
  - 应用
    - 补偿
      - 定时任务模式
      - 消息队列模式
    - 幂等
    - 异步化
- [Websocket 百万长连接技术实践](https://mp.weixin.qq.com/s/MUourpb0IqqFo5XlxRLE0w)
  - GateWay
    - 网关拆分为网关功能部分和业务处理部分，网关功能部分为 WS-Gateway：集成用户鉴权、TLS 证书验证和 WebSocket 连接管理等；业务处理部分为 WS-API：组件服务直接与该服务进行 gRPC 通信。可针对具体的模块进行针对性扩容；服务重构加上 Nginx 移除，整体硬件消耗显著降低；服务整合到石墨监控体系
    ![img.png](system_design_gatewa.png)
    - 网关 客户端连接流程：
      - 客户端与 WS-Gateway 服务通过握手流程建立 WebSocket 连接；
      - 连接建立成功后，WS-Gateway 服务将会话进行节点存储，将连接信息映射关系缓存到 Redis 中，并通过 Kafka 向 WS-API 推送客户端上线消息；
      - WS-API 通过 Kafka 接收客户端上线消息及客户端上行消息；
      - WS-API 服务预处理及组装消息，包括从 Redis 获取消息推送的必要数据，并进行完成消息推送的过滤逻辑，然后 Pub 消息到 Kafka；
      - WS-Gateway 通过 Sub Kafka 来获取服务端需要返回的消息，逐个推送消息至客户端。
  - TLS 内存消耗优化
    - Go TLS 握手过程中消耗的内存 [issue](https://github.com/golang/go/issues/43563)
    - 采用七层负载均衡，在七层负载上进行 TLS 证书挂载，将 TLS 握手过程移交给性能更好的工具完成
  - Socket ID 设计
    - K8S 场景中，采用注册下发的方式返回编号，WS-Gateway 所有副本启动后向数据库写入服务的启动信息，获取副本编号，以此作为参数作为 SnowFlake 算法的副本编号进行 Socket ID 生产，服务重启会继承之前已有的副本编号，有新版本下发时会根据自增 ID 下发新的副本编号
  - 心跳机制
    - 避免大量客户端同时进行心跳上报对 Redis 产生压力。
      - 客户端建立 WebSocket 连接成功后，服务端下发心跳上报参数；
      - 客户端依据以上参数进行心跳包传输，服务端收到心跳后会更新会话时间戳；
      - 客户端其他上行数据都会触发对应会话时间戳更新；
      - 服务端定时清理超时会话，执行主动关闭流程；
    - 实现心跳正常情况下的动态间隔，每 x 次正常心跳上报，心跳间隔增加 a，增加上限为 y
  - 消息接收与发送
    - 发现每个 WebSocket 连接都会占用 3 个 goroutine，每个 goroutine 都需要内存栈，单机承载连十分有限，主要受制于大量的内存占用，而且大部分时间 c.writer() 是闲置状态，于是考虑，是否只启用 2 个 goroutine 来完成交互
    - 保留 c.reader() 的 goroutine，如果使用轮询方式从缓冲区读取数据，可能会产生读取延迟或者锁的问题，c.writer() 操作调整为主动调用，不采用启动 goroutine 持续监听，降低内存消耗
    - 调研了 gev 和 gnet 等基于事件驱动的轻量级高性能网络库，实测发现在大量连接场景下可能产生的消息延迟的问题，所以没有在生产环境下使用
  - 核心对象缓存
    - 网关部分的核心对象为 Connection 对象，围绕 Connection 进行了 run、read、write、close 等函数的开发
    - 使用 sync.pool 来缓存该对象，减轻 GC 压力，创建连接时，通过对象资源池获取 Connection 对象，生命周期结束之后，重置 Connection 对象后 Put 回资源池
      ```go
      var ConnectionPool = sync.Pool{
         New: func() interface{} {
            return &Connection{}
         },
      }
      
      func GetConn() *Connection {
         cli := ConnectionPool.Get().(*Connection)
         return cli
      }
      
      func PutConn(cli *Connection) {
         cli.Reset()
         ConnectionPool.Put(cli) // 放回连接池
      }
      ```
  - 数据传输过程优化
    - 消息流转过程中，需要考虑消息体的传输效率优化，采用 MessagePack 对消息体进行序列化，压缩消息体大小
- [动手实现一个localcache](https://mp.weixin.qq.com/s/ZtSA3J8HK4QarhrJwBQtXw)
  - 数据结构 - HashTable
  - 并发安全 
    - 在读操作远多于写操作的时候，使用sync.map的性能是远高于map+sync.RWMutex的组合的
    - 我们本地缓存不仅支持进行数据存储的时候要使用锁，进行过期清除等操作时也需要加锁，所以使用map+sync.RWMutex的方式更灵活
  - 高性能并发访问
    - 我们可以使用djb2哈希算法把key打散进行分桶，然后在对每一个桶进行加锁，也就是锁细化，减少竞争
  - 淘汰策略
    - LFU
      - 根据数据的历史访问频率来淘汰数据，这种算法核心思想认为最近使用频率低的数据,很大概率不会再使用，把使用频率最小的数据置换出去
      - 问题：某些数据在短时间内被高频访问，在之后的很长一段时间不再被访问，因为之前的访问频率急剧增加，那么在之后不会在短时间内被淘汰，占据着队列前头的位置，会导致更频繁使用的块更容易被清除掉，刚进入的缓存新数据也可能会很快的被删除
    - LRU
      - 根据数据的历史访问记录来淘汰数据，这种算法核心思想认为最近使用的数据很大概率会再次使用，最近一段时间没有使用的数据，很大概率不会再次使用，把最长时间未被访问的数据置换出去
      - 问题：当某个客户端访问了大量的历史数据时，可能会使缓存中的数据被历史数据替换，降低缓存命中率
    - FIFO
      - 这种算法的核心思想是最近刚访问的，将来访问的可能性比较大，先进入缓存的数据最先被淘汰掉。
      - 问题：这种算法采用绝对公平的方式进行数据置换，很容易发生缺页中断问题
    - Two Queue
      - 是FIFO + LRU的结合，其核心思想是当数据第一次访问时，将数据缓存在FIFO队列中，当数据第二次被访问时将数据从FIFO队列移到LRU队列里面，这两个队列按照自己的方法淘汰数据。
      - 问题：这种算法和LRU-2一致，适应性差，存在LRU中的数据需要大量的访问才会将历史记录清除掉
    - ARU
      - 即自适应缓存替换算法，是LFU和LRU算法的结合使用，其核心思想是根据被淘汰数据的访问情况，而增加对应 LRU 还是 LFU链表的大小，ARU主要包含了四个链表，LRU 和 LRU Ghost ，LFU 和LFU Ghost， Ghost 链表为对应淘汰的数据记录链表，不记录数据，只记录 ID 等信息
      - 当数据被访问时加入LRU队列，如果该数据再次被访问，则同时被放到 LFU 链表中；如果该数据在LRU队列中淘汰了，那么该数据进入LRU Ghost队列，如果之后该数据在之后被再次访问了，就增加LRU队列的大小，同时缩减LFU队列的大小。
      - 问题：因为要维护四个队列，会占用更多的内存空间
  - 过期清除
    - 除了使用缓存淘汰策略清除数据外，还可以添加一个过期时间做双重保证，避免不经常访问的数据一直占用内存。可以有两种做法：
      - 数据过期了直接删除   
      - 数据过期了不删除，异步更新数据
  - 缓存监控
    - 可以使用Prometheus进行监控上报，我们自测可以简单写一个小组件，定时打印缓存数、缓存命中率等指标
  - GC调优
  - 缓存穿透
    - 使用缓存就要考虑缓存穿透的问题，不过这个一般不在本地缓存中实现，基本交给使用者来实现，当在缓存中找不到元素时,它设置对缓存键的锁定;这样其他线程将等待此元素被填充,而不是命中数据库（外部使用singleflight封装一下）
- [缓存淘汰算法](https://zhuanlan.zhihu.com/p/352910565)
  - LFU
    - 最近使用频率高的数据很大概率将会再次被使用,而最近使用频率低的数据,很大概率不会再使用
    - 把使用频率最小的数据置换出去。这种算法是完全从使用频率的角度去考虑的
    - Issue: 某些数据短时间内被重复引用，并且在很长一段时间内不再被访问。由于它的访问频率计数急剧增加，即使它在相当长的一段时间内不会被再次使用，也不会在短时间内被淘汰。这使得其他可能更频繁使用的块更容易被清除，此外，刚进入缓存的新项可能很快就会再次被删除，因为它们的计数器较低，即使之后可能会频繁使用。
    - 参见Redis处理，加入冷却时间以及指数增长
  - LRU
    - 最近使用的数据很大概率将会再次被使用。而最近一段时间都没有使用的数据，很大概率不会再使用。把最长时间未被访问的数据置换出去。这种算法是完全从最近使用的时间角度去考虑的
    - Issue： 如果某个客户端访问大量历史数据时，可能使缓存中的数据被这些历史数据替换，其他客户端访问数据的命中率大大降低。
  - ARC 自适应缓存替换算法,它结合了LRU与LFU,来获得可用缓存的最佳使用。
    - 当时访问的数据趋向于访问最近的内容，会更多地命中LRU list，这样会增大LRU的空间； 当系统趋向于访问最频繁的内容，会更多地命中LFU list，这样会增加LFU的空间.
    - 执行过程
      - 整个Cache分成两部分，起始LRU和LFU各占一半，后续会动态适应调整partion的位置（记为p）除此，LRU和LFU各自有一个ghost list(因此，一共4个list)
      - 在缓存中查找客户端需要访问的数据， 如果没有命中，表示缓存穿透，将需要访问的数据 从磁盘中取出，放到LRU链表的头部。
      - 如果命中，且LFU链表中没有，则将数据放入LFU链表的头部，所有LRU链表中的数据都必须至少被访问两次才会进入LFU链表。
      - 如果命中，且LFU链表中存在，则将数据重新放到LFU链表的头部。这么做，那些真正被频繁访问的页面将永远呆在缓存中，不经常访问的页面会向链表尾部移动，最终被淘汰出去。
      - 如果此时缓存满了，则从LRU链表中淘汰链表尾部的数据，将数据的key放入LRU链表对应的ghost list。然后再在链表头部加入新数据。如果ghost list中的元素满了，先按照先进先出的方式来淘汰ghost list中的一个元素，然后再加入新的元素
      - 如果没有命中的数据key处于ghost list中，则表示是一次幽灵（phantom）命中，系统知道，这是一个刚刚淘汰的页面，而不是第一次读取或者说很久之前读取的一个页面。ARC用这个信息来调整它自己，以适应当前的I/O模式（workload）。 这个迹象说明我们的LRU缓存太小了。在这种情况下，LRU链表的长度将会被增加1，并将命中的数据key从ghost list中移除，放入LRU链表的头部。显然，LFU链表的长度将会被减少1。
        同样，如果一次命中发生在LFU ghost 链表中，它会将LRU链表的长度减一，以此在LFU 链表中加一个可用空间。
      ![img.png](system_design_arc1.png)
      ![img.png](system_design_arc2.png)
  - 2Q
    - 有两个缓存队列，一个是FIFO队列，一个是LRU队列。当数据第一次访问时，2Q算法将数据缓存在FIFO队列里面，当数据第二次被访问时，则将数据从FIFO队列移到LRU队列里面，两个队列各自按照自己的方法淘汰数据
    - 执行过程
      - 新访问的数据插入到FIFO队列；
      - 如果数据在FIFO队列中一直没有被再次访问，则最终按照FIFO规则淘汰；
      - 如果数据在FIFO队列中被再次访问，则将数据移到LRU队列头部；
      - 如果数据在LRU队列再次被访问，则将数据移到LRU队列头部；
      - LRU队列淘汰末尾的数据。
- [动手实现一个localcache - 欣赏优秀的开源设计](https://mp.weixin.qq.com/s/KfxfRqTrFvt9K5dEVU94yw)
  - 高效的并发访问
    - 本地缓存的简单实现可以使用map[string]interface{} + sync.RWMutex的组合，使用sync.RWMutex对读进行了优化
    - 当并发量上来以后，还是变成了串行读，等待锁的goroutine就会block住。为了解决这个问题我们可以进行分桶，每个桶使用一把锁，减少竞争。分桶也可以理解为分片，每一个缓存对象都根据他的key做hash(key)，然后在进行分片：hash(key)%N，N就是要分片的数量
      - 分片的实现主要考虑两个点：
        - hash算法的选择，哈希算法的选择要具有如下几个特点：
          - 哈希结果离散率高，也就是随机性高
          - 避免产生多余的内存分配，避免垃圾回收造成的压力
          - 哈希算法运算效率高
        - 分片的数量选择，分片并不是越多越好，根据经验，我们的分片数可以选择N的2次幂，分片时为了提高效率还可以使用位运算代替取余。
    - 开源实现
      - 开源的本地缓存库中 bigcache、go-cache、freecache都实现了分片功能
        - bigcache的hash选择的是fnv64a算法、
        - go-cache的hash选择的是djb2算法、
        - freechache选择的是xxhash算法。
        - 通过对比结果我们可以观察出来Fnv64a算法的运行效率还是很高
  - 减少GC
    - [bigcache](https://pengrl.com/p/35302/)做到避免高额GC的设计是基于Go语言垃圾回收时对map的特殊处理
      - 如果map对象中的key和value不包含指针，那么垃圾回收器就会无视他，针对这个点们的key、value都不使用指针，就可以避免gc。bigcache使用哈希值作为key，然后把缓存数据序列化后放到一个预先分配好的字节数组中，使用offset作为value，
    - freecache中的做法是自己实现了一个ringbuffer结构，通过减少指针的数量以零GC开销实现map
      - key、value都保存在ringbuffer中，使用索引查找对象。freecache与传统的哈希表实现不一样，实现上有一个slot的概念
- [优先级队列](https://mp.weixin.qq.com/s/eXJcjPnXiy733k79Y1vbBg)
  - 三个重要的角色，分别是优先级队列、工作单元Job、消费者worker (队列-消费者模式)
    - 队列
      ```go
      type JobQueue struct {
        mu sync.Mutex //队列的操作需要并发安全
        jobList *list.List //List是golang库的双向队列实现，每个元素都是一个job
        noticeChan chan struct{} //入队一个job就往该channel中放入一个消息，以供消费者消费
      }
      func (queue *JobQueue) PushJob(job Job) {
        queue.jobList.PushBack(job) //将job加到队尾
        queue.noticeChan <- struct{}{}
      }
      func (queue *JobQueue) PopJob() Job {
        queue.mu.Lock()
        defer queue.mu.Unlock()
      
        if queue.jobList.Len() == 0 {
        return nil
        }
      
        elements := queue.jobList.Front() //获取队列的第一个元素
        return queue.jobList.Remove(elements).(Job) //将元素从队列中移除并返回
      }
      
      func (queue *JobQueue) WaitJob() <-chan struct{} {
        return queue.noticeChan
      }
      ```
    - 工作单元Job
      ```go
      type BaseJob struct {
        Err error
        DoneChan chan struct{} //当作业完成时，或者作业被取消时，通知调用者
        Ctx context.Context
        cancelFunc context.CancelFunc
      }
      ```
      ![img.png](system_design_queue.png)
    - 
  - 消费者Worker
    ```go
    type WorkerManager struct {
        queue *JobQueue
        closeChan chan struct{}
    }
    func (m *WorkerManager) StartWork() error {
        fmt.Println("Start to Work")
        for {
            select {
            case <-m.closeChan:
              return nil
    
            case <-m.queue.noticeChan:
              job := m.queue.PopJob()
              m.ConsumeJob(job)
            }
        }
    
        return nil
    }
    
    func (m *WorkerManager) ConsumeJob(job Job) {
      defer func() {
          job.Done()
      }()
    
      job.Execute()
    }
    ```
- [HTTP Server](https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html)
  - Return the handler
    ```go
    func (s *server) handleSomething() http.HandlerFunc {
        thing := prepareThing()
        return func(w http.ResponseWriter, r *http.Request) {
            // use thing        
        }
    }
    ```
  - HandlerFunc over Handler
    ```go
    func (s *server) handleSomething() http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            ...
        }
    }
    ```
- [高并发下如何设计秒杀系统](https://mp.weixin.qq.com/s/Qoo5yPa5Mmzgb3_kIoTXbw)
  - 瞬时高并发
    - 页面静态化
    - CDN加速
    - 缓存
    - mq异步处理
    - 限流
    - 分布式锁
- [图解淘宝10年后台架构演进](https://mp.weixin.qq.com/s/liMJO48daWzkzh37vxMbPg)
  - 引入本地缓存和分布式缓存
  - 引入反向代理实现负载均衡
  - 数据库读写分离
  - 数据库按业务分库
  - 把大表拆分为小表
  - 使用LVS或F5来使多个Nginx负载均衡
    - 由于瓶颈在Nginx，因此无法通过两层的Nginx来实现多个Nginx的负载均衡。图中的LVS和F5是工作在网络第四层的负载均衡解决方案，其中LVS是软件，运行在操作系统内核态，可对TCP请求或更高层级的网络协议进行转发，因此支持的协议更丰富，并且性能也远高于Nginx，可假设单机的LVS可支持几十万个并发的请求转发；F5是一种负载均衡硬件，与LVS提供的能力类似，性能比LVS更高，但价格昂贵。
    - 由于LVS是单机版的软件，若LVS所在服务器宕机则会导致整个后端系统都无法访问，因此需要有备用节点。可使用keepalived软件模拟出虚拟IP，然后把虚拟IP绑定到多台LVS服务器上，浏览器访问虚拟IP时，会被路由器重定向到真实的LVS服务器，当主LVS服务器宕机时，keepalived软件会自动更新路由器中的路由表，把虚拟IP重定向到另外一台正常的LVS服务器，从而达到LVS服务器高可用的效果。
  - 通过DNS轮询实现机房间的负载均衡
  - 引入NoSQL数据库和搜索引擎等技术
  - 大应用拆分为小应用
  - 复用的功能抽离成微服务
  - 引入企业服务总线ESB屏蔽服务接口的访问差异
  - 引入容器化技术实现运行环境隔离与动态服务管理
  - 以云平台承载系统
- [用 MQ 解耦其实是骗你的](https://xargin.com/mq-is-not-savior/)
  - MQ解耦的故事
    - 核心系统依赖下游系统是因为调用关系，下游系统依赖核心系统是因为下游系统要使用核心系统的数据
    - 我们使用 MQ 只是解开了单个方向上的依赖，核心系统没有对下游系统的调用了。
    - 隐式依赖导致事故
  - 解决方法?
    - 接受有些复杂场景下，上下游就是要有耦合的事实
    - 使用 schema registry 之类的方案，让上下游使用一份 domain event 的 schema，也可以用 pb 之类的文件来描述，让队列里的消息也能类似 API 定义一样，有“软件契约”
    - 增加专门的 data validation 服务，类似 Google 在应对机器学习数据问题时的一个方案：https://blog.acolyer.org/2019/06/05/data-validation-for-machine-learning/
    - 如果公司内有 mono repo，那么在有 mono repo + schema 描述的前提下，适当开发一些静态分析工具，来阻止上游程序员犯错
    - 耦合实在解除不了的情况下，让下游的计算和查询模块做好分离，重点保障查询模块的稳定性，不要拖累主流程
- [高可用的 11 个关键技巧](https://mp.weixin.qq.com/s/frHzdZRZaab8u8kSkB954Q)
  - 大型互联网架构设计，讲究一个四件套组合拳玩法，高并发、高性能、高可用、高扩展
  - 高可用都有哪些设计技巧
    - 系统拆分
      - 一个复杂的业务域按DDD的思想拆分成若干子系统，每个子系统负责专属的业务功能，做好垂直化建设，各个子系统之间做好边界隔离，降低风险蔓延
    - 解耦
      - 高内聚、低耦合 - 小到接口抽象、MVC 分层，大到 SOLID 原则、23种设计模式。核心都是降低不同模块间的耦合度，避免一处错误改动影响到整个系统
      - AOP - 核心就是采用动态代理技术，通过对字节码进行增强，在方法调用的时候进行拦截，以便于在方法调用前后，增加我们需要的额外处理逻辑
      - 事件机制，通过发布订阅模式，新增的需求，只需要订阅对应的事件通知，针对性消费即可
    - 异步
      - 线程池（ThreadPoolExecutor）, 消息队列
    - 重试
      - 重试主要是体现在远程的RPC调用，受 网络抖动、线程资源阻塞 等因素影响，请求无法及时响应。
      - 重试通常跟幂等组合使用，如果一个接口支持了 幂等，那你就可以随便重试
      - 幂等 的解决方案
        - 插入前先执行查询操作，看是否存在，再决定是否插入
        - 增加唯一索引
        - 建防重表
        - 引入状态机，比如付款后，订单状态调整为已付款，SQL 更新记录前 增加条件判断
        - 增加分布式锁
        - 采用 Token 机制，服务端增加 token 校验，只有第一次请求是合法的
    - 补偿
      - 我们知道不是所有的请求都能收到成功响应。除了上面的 重试 机制外，我们还可以采用补偿玩法，实现数据最终一致性
      - 补偿操作有个重要前提，业务能接受短时间内的数据不一致。
      - 补偿有很多的实现方式：
        - 本地建表方式，存储相关数据，然后通过定时任务扫描提取，并借助反射机制触发执行
        - 也可以采用简单的消息中间件，构建业务消息体，由下游的的消费任务执行。如果失败，可以借助MQ的重试机制，多次重试
    - 备份
    - 多活策略
      - 同城双活、两地三中心、三地五中心、异地双活、异地多活
    - 隔离
    - 限流
      - 计数器限流
      - 滑动窗口限流
      - 漏桶限流
      - 令牌桶限流
    - 熔断
    - 降级
      - 降级是通过暂时关闭某些非核心服务或者组件从而保护核心系统的可用性
- [高并发设计，都有哪些技术方案](https://mp.weixin.qq.com/s/89o-GHFNyIKrrIjkQMquFA)
  - 负载均衡
    - 它的职责是将网络请求 “均摊”到不同的机器上。避免集群中部分服务器压力过大，而另一些服务器比较空闲的情况
    - 随机算法
    - 轮询算法
    - 轮询权重算法
    - 一致性哈希算法
    - 最小连接
    - 自适应算法
    - 常用负载均衡工具：
      - LVS
      - Nginx
      - HAProxy
    - 对于一些大型系统，一般会采用 DNS+四层负载+七层负载的方式进行多层次负载均衡。
  - 分布式微服务
    - 采用分而治之的思想，通过SOA架构，将一个大的系统拆分成若干个微服务，粒度越来越小，称之为微服务架构
  - 缓存机制
    - 性能不够，缓存来凑
    - 缓存更新常用策略
      - Cache aside，通常会先更新数据库，然后再删除缓存，为了兜底还会设置缓存时间。
      - Read/Write through， 一般是由一个 Cache Provider 对外提供读写操作，应用程序不用感知操作的是缓存还是数据库。
      - Write behind，延迟写入，Cache Provider 每隔一段时间会批量写入数据库，大大提升写的效率。像操作系统的page cache也是类似机制。
  - 分布式关系型数据库
    - 分表又可以细分为 垂直分表 和 水平分表 两种形式
      - 垂直分表
        - 表由“宽”变“窄”，简单来讲，就是将大表拆成多张小表
        - 冷热分离，把常用的列放在一个表，不常用的放在一个表。
        - 字段更新、查询频次拆分
        - 大字段列独立存放
        - 关系紧密的列放在一起
      - 水平分表
      - 数据量大，就分表；并发高，就分库
  - 分布式消息队列
    - 异步处理。将一个请求链路中的非核心流程，拆分出来，异步处理，减少主流程链路的处理逻辑，缩短RT，提升吞吐量。如：注册新用户发短信通知。
    - 削峰填谷。避免流量暴涨，打垮下游系统，前面会加个消息队列，平滑流量冲击。比如：秒杀活动。生活中像电源适配器也是这个原理。
    - 应用解耦。两个应用，通过消息系统间接建立关系，避免一个系统宕机后对另一个系统的影响，提升系统的可用性。如：下单异步扣减库存
    - 消息通讯。内置了高效的通信机制，可用于消息通讯。如：点对点消息队列、聊天室。
  - CDN 全称 （Content Delivery Network），内容分发网络
    - CDN = 镜像（Mirror）+缓存（Cache）+整体负载均衡（GSLB）
    - 本地Cache加速
    - 镜像服务
    - 远程加速
    - 带宽优化
    - 集群抗攻击
- [超大规模分布式存储系统架构设计](https://mp.weixin.qq.com/s/IJhXKZSa5SBiBgXg0dJkHg)
- [超大规模分布式存储系统架构设计——浅谈B站对象存储](https://mp.weixin.qq.com/s/0FL8WbsBSg9hsqUil3C6KA)


