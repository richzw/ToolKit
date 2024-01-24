
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
- Hummock 
  - rust 自研了云原生的 LSM 存储引擎 Hummock ，并将之用来存储流计算中有状态算子的状态
  - 与一般的 LSM 存储引擎类似，新写入 Hummock 的数据将会写入 mutable mem-table 中，在特定条件下，mutable mem-table 会被 freeze 成 immutable mem-table
  - 最终immutable mem-table 会写成 SST(sorted string table)文件落入持久化的存储中，同时SST会在 LSM 的元数据中被加到 overlapping L0 中
  - 在处理 get 请求时，在经过 min-max 以及 bloom filter 的过滤后，Hummock 会从最上层的 mutable mem-table 到最下层的 non-overlapping levels 逐一查找，当查找到对应的 key 时，停止查找并返回对应的 value。
  - 在处理 iter 请求时，与处理 get 请求不同，经过前期的过滤后，给定范围内的数据可能存在于任意一层数据中
- 物化视图
  - 物化视图是数据库中用来存储操作、计算或查询结果的对象，它们相当于数据的本地快照，可以被重复使用，而无需重新计算或重新获取数据。使用物化视图可以节省处理时间和成本，并允许预先加载经常使用或最耗时的查询。
  - 当数据不断变化和更新但只偶尔查看时，通常会使用视图。而当数据需要经常查看，但保存的表不经常更新时，则会使用物化视图。
  - RisingWave 会将物化视图的数据存储在 Hummock 中，这样可以保证物化视图的数据与流计算的状态数据在同一个存储引擎中，从而可以在同一个事务中读取这两类数据。
  - 物化视图可以是只读的、可更新的或可写入的。在只读物化视图上不能执行 DML 语句（例如 INSERT 和 MERGE 语句）；但在可更新和可写的物化视图上也许可以执行这些语句
- 触发机制
  - Barrier aka window
    - DAG 上的 source 会在数据流中插入 Barrier，这些 Barrier 将输入流切分成很多段。
  - Watermark 与 Trigger
    - 在事件时间（Event Time）时间列上定义时间窗口（Time Window）对数据进行划分，使用水位线（Watermark）来描述事件的完整性，使得计算引擎可以在每个窗口内的事件完整且不会更改后，再使用类似批处理的方式进行处理。这种触发计算的方式也被称为完整性触发。
    - 在流计算系统中通过某些 timeout 或其他策略指定注入 Watermark。比如在 RisingWave 中可以通过如下 SQL 定义 timeout 为 5s 的 Watermark。
- [RisingWave 打造 Feature Store](https://mp.weixin.qq.com/s/KojIae28RGat-Wi_sVaqRA)
- [RisingWave 窗口函数](https://mp.weixin.qq.com/s/rgJTR6Ynn8FmkfvCAQZIwA)
- [多流 Join](https://mp.weixin.qq.com/s/YZzAqgHXsii3lBow_dE7ug)
  - 流处理 Join 使用的算法基本上都是 Symmetric Hash Join（需要有等值连接条件）
  - Unbounded state
    - 由于 Join 输入是 unbounded 的，可以推导出 Join 的状态也是 unbounded的。显然这会导致存储上的问题
    - RisingWave 通过存算分离的架构，可以把 Join 的状态存储到对象存储 (Object store) 当中。
    - 为了弥补对象存储访问延迟较高的问题，Risingwave 会利用内存和本地盘来缓存对象存储的文件，并通过 LRU 策略管理这些缓存
  - Watermark & Window
    - 我们不希望 Join 的状态大小也是 Unbounded 的。通过水位线和窗口 Join 技术可以将 Join 状态控制在一个有限的大小之内
    - 窗口在流处理上一般是通过 TUMBLE 或者 HOP Window 函数的方式为数据划分时间窗口，如果划分的时间字段带有 Watermark 信息，那么经过窗口函数后优化器也可以进一步推导出窗口时间列的 Watermark信息
  - Interval Join
    - 假如你正在处理的是用户的点击流数据，你可能想要连接用户的点击事件和他们的购买事件，但是这两个事件可能不会在严格的窗口期间内发生。在这种情况下，使用 Interval Join 就会更加合适
    - Interval Join 允许两个事件在一定的时间间隔内连接，而不是在严格的窗口期间内
  - Temporal Join
    - 传统数据库中的 Hash Join 只需要选择一边建立 Hash Table。为了提高性能，一个思考方向是打破对 Join 两边输入的对等关系
    - Temporal Join它可以将一边流 Join 一个 Risingwave 内部表或者物化视图，流的一边可以不再维护状态，当有数据过来时，可以直接查询另外一边的表，同时表则充当了 Join 状态本身
  - Join Ordering
    - 传统数据库很重要的一个思想是利用CBO（Cost Based Optimizer）来枚举执行计划搜索空间， 利用表的统计信息，估算每个算子的需要处理的数据量，计算出执行计划的代价，最后采用估算代价最低的执行计划
    - RisingWave 目前使用的 Join Ordering 算法是**尽量地将这棵树变成 Bushy Tree 并使得它的树高最低**
  - 子查询
    - RisingWave 的子查询 Unnesting 技术是按照 Paper：Unnesting arbitrary queries 来实现的。
    - 所有的子查询都会被转成 Apply 算子，并将 Apply算子不断往下推直到所有关联项都被改写成普通引用后就可以转成Join
  - Delta Join
    - 通过前文我们可以了解到 Join 是一个重状态的算子，每个 Join 有需要维护自身的 Join 状态。那如果有多条 SQL 都使用了同一个表输入，并且 Join Key 都一样，我们可以复用它的状态
  - Multi-Way Joins 
  - 快慢流





















