
- [RisingWave 流处理 Join](https://mp.weixin.qq.com/s/u_9Y_cuOiuRGT17XU_34zA)
  - Symmetric Hash Join
  - Interval Join
    - Interval Join允许两个事件在一定的时间间隔内连接，而不是在严格的窗口期间内
  - Temporal Join
    - 传统数据库中的Hash Join只需要选择一边建立Hash Table。为了提高性能，一个思考方向是打破对Join两边输入的对等关系
    - Risingwave提供Temporal Join，它可以将一边流Join一个Risingwave内部表或者物化视图，流的一边可以不再维护状态，当有数据过来时，可以直接查询另外一边的表，同时表则充当了Join状态本身
  - Join Ordering
    - 传统数据库中针对Join Ordering的优化文献浩如烟海，其中很重要的一个思想是利用CBO（Cost Based Optimizer）来枚举执行计划搜索空间， 利用表的统计信息，估算每个算子的需要处理的数据量，计算出执行计划的代价，最后采用估算代价最低的执行计划。
    - Risingwave目前使用的Join Ordering算法是**尽量地将这棵树变成Bushy Tree并使得它的树高最低**
  - NestedLoop Join
  - Delta Join
  - 快慢流 / Multi-Way Joins
- 对象存储优化
  - 流式写入
    - 在 LSM compaction 或算子状态持久化期间，我们需要将已排序的键值对写入多个 SST 文件中。一种直接的方法是在内存中缓冲整个SST，并在其大小达到容量时启动后台任务进行上传。
    - 我们使用 Multipart上传来解决这个问题。它允许将单个对象上传为一组parts；每个 part 是待上传对象的一个连续部分
  - 流式读取
    - LSM compaction 期间，我们还需要读取和遍历 SST 文件
    - AWS SDK 提供了一个函数into_async_read 来把返回的 HTTP 流封装成一个 tokio::io::AsyncRead 对象。然后我们可以在这个对象上调用 read_exact ，以字节流的形式读取数据。
    - 我们在 SST 字节流上创建了一个 BlockStream 抽象。其可以根据 SST 元数据以块为粒度进行读取，并解析其中的键值对，做归并排序。
  - S3 存储桶支持多个前缀
    - 我们采用了一种简单的方法来设置集群的最大前缀数量，并基于其 ID 的哈希值将 SST 文件分配到不同的前缀中。例如，我们可以将最大前缀数配置为128，理论上可以支持每秒 3,500 * 128 的写入或 5,500 * 128 的读取请求。
- 利用随机化的 SQL 测试来帮助检测错误
  - SQLSmith 是一个用于自动生成和测试 SQL 查询的工具。它旨在通过生成随机的有效 SQL 查询并在目标数据库上执行这些查询来探索数据库系统的功能和限制。
- [存储引擎 Hummock 及其存储架构](https://mp.weixin.qq.com/s/PXkkOaikx0h54Msm0HEFSw)
  - 一致性快照
    - Hummock 提供一致性读取（Snapshot Read），这对于增量数据的 Join 计算是十分有必要的
    - Hummock 使用与 Barrier 绑定的 Epoch 作为所有写入数据的 MVCC 版本号，因此我们在查询时可以利用当前算子所流过的 Barrier 来指定读取 Hummock 中对应的版本，对于指定的查询 Epoch ，如果某个目标 key 存在比 Epoch 更大的版本号，则忽略该版本数据，并且查询定位到小于等于 Epoch 的最大（新）的那个版本。
  - Schema-aware Bloom Filter
    - LSM Tree 架构的存储引擎的数据文件都按照写入顺序或者其他规则分割成多层，这意味着即便只读取某一极小范围的数据，依然避免不了要查询多个文件，这将带来额外的 IO 与计算操作。
    - 一种通用的做法是为同一个文件的所有 key 建立 Bloom Filter，遇到查询时先通过 Bloom Filter 过滤掉不必要的文件，然后再对剩下的文件进行查询。
    - Hummock 为每个文件建立了 Bloom Filter，但是 Hummock 的 Bloom Filter 与传统的 Bloom Filter 不同，它是 Schema-aware 的，即 Hummock 的 Bloom Filter 会根据文件中的数据类型来选择不同的哈希函数，这样可以大大降低 Bloom Filter 的误判率。
  - Sub Level
    - Hummock 为了解决这个问题，引入了 Sub Level 的概念，即在每个 Level 中，Hummock 会将数据按照 key 的前缀进行分组，每个分组称为一个 Sub Level，每个 Sub Level 都会有一个 Bloom Filter，这样在查询时，我们可以先通过 Bloom Filter 过滤掉不必要的 Sub Level，然后再对剩下的 Sub Level 进行查询。
    - 为了提高 L0 文件的 Compact 速度，我们参考了 CockroachDB 存储引擎pebble的设计
  - Fast Checkpoint
    - RisingWave 默认每10秒执行一次 Checkpoint ，如果用户的集群节点因为各种原因而宕机，那么在集群恢复之后，RisingWave 仅需要重新处理最近十秒的历史数据，便可以跟上用户的最新输入，极大地减轻了故障对业务造成的影响。
    - 刷新任务会持有内存数据的引用，在后台将内存数据序列化为文件格式并且上传到 S3 中，因此，频繁的 Checkpoint 并不会阻塞数据流的计算。
    - RisingWave 将整个集群的数据划分为了多个 group （初始时只有2个，一个存储状态，另一个存储 Matieriallize View），同一个计算节点上的同一个 group 的所有算子的状态变更，会合并写入到同一个文件中，对单节点集群来说，一次 Checkpoint 只会生成2个文件。
    - 但是某些场景下，这样产生的文件仍然是非常小的。为了避免加重写放大的负担，我们增加了多种策略来判断是否应当先将当前 Level 0 的多个小文件合并为更大的文件后，再合并到更下层，见 L0 Intra Compaction 中的介绍。
    - 如果某个 group 的数据量过大，为了降低单个 LSM Tree 执行 Compaction 时的写放大开销，RisingWave 会自动分裂数据量过大的 group。




