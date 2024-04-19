
- [Index](https://www.mongodb.com/blog/post/performance-best-practices-indexing)
  - Use Compound Indexes
    
    For compound indexes, follow [ESR rule](https://www.alexbevi.com/blog/2020/05/16/optimizing-mongodb-compound-indexes-the-equality-sort-range-esr-rule/)
    - First, add those fields against which **Equality** queries are run.
    - The next fields to be indexed should reflect the **Sort** order of the query.
    - The last fields represent the **Range** of data to be accessed.
    
    Operator Type Check 
    - Inequality - _$ne $nin_ belong to **Range**
    - Regex - _/car/_ belong to **Range**
    - _$in_ 
      - Alone: a series of **Equality** matches
      - Combined: possible a **Range**
        - May optimize **blocking sort** with **Merge Sort**
    
    Exception
      - check **totalDocsExamined**
  - Use Covered Queries When Possible
    - Covered queries return results from an index directly without having to access the source documents, and are therefore very efficient.
    - If the _explain()_ output displays **totalDocsExamined** as _0_, this shows the query is covered by an index.
    - A common gotcha when trying to achieve covered queries is that the __id_ field is always returned by default. You need to explicitly exclude it from query results, or add it to the index.
  - Use Caution When Considering Indexes on Low-Cardinality Fields
    - Queries on fields with a small number of unique values (low cardinality) can return large result sets. Compound indexes may include fields with low cardinality, but the value of the combined fields should exhibit high cardinality.
  - Wildcard Indexes Are Not a Replacement for Workload-Based Index Planning
    - If your application’s query patterns are known in advance, then you should use more selective indexes on the specific fields accessed by the queries
  - Use text search to match words inside a field
    - If you only want to match on a specific word in a field with a lot of text, then use a text index.
    - If you are running MongoDB in the Atlas service, consider using Atlas Full Text Search which provides a fully-managed Lucene index integrated with the MongoDB database.
  - Use Partial Indexes
    - Reduce the size and performance overhead of indexes by only including documents that will be accessed through the index
  - Take Advantage of Multi-Key Indexes for Querying Arrays
    - If your query patterns require accessing individual array elements, use a multi-key index.
  - Avoid Regular Expressions That Are Not Left Anchored or Rooted
    - Indexes are ordered by value. Leading wildcards are inefficient and may result in full index scans. Trailing wildcards can be efficient if there are sufficient case-sensitive leading characters in the expression.
  - Avoid Case Insensitive Regular Expressions
    - If the sole reason for using a regex is case insensitivity, use a case insensitive index instead, as those are faster.
  - Use Index Optimizations Available in the WiredTiger Storage Engine
    - If you are self-managing MongoDB, you can optionally place indexes on their own separate volume, allowing for faster disk paging and lower contention. See wiredTiger options for more information.
- [Tips and Tricks++ for Querying and Indexing MongoDB](https://www.youtube.com/watch?v=5mBY27wVau0&list=PL4RCxklHWZ9u_xtprouvxCvzq2m6q_0_E&index=10)
  - ESR rule
    - A good starting place applicable to most user cases
    - Place keys in the following order
      - Equality first
      - Sort next
      - Range last
    - Remember
      - Some operators may be range instead of equality
      - Having consecutive keys used in the index is important
  - Operator Type Check
    - Inequality: $ne, $nin -> Range
    - Regex Operator: /car/, /^car/i -> Range
    - $in
      - it depends with respect to key ordering
      - Alone: a series of Equality matches
      - Combines: possible a Range
      - Mongo optimize it as Merge Sort instead of Blocking Sort
  - Consecutive Index keys
  - Is the ESR rule always optimal? Nope
    - Check total keys examined from execute plan
- [MongoDB中的锁](https://mp.weixin.qq.com/s/FxaUhtRho5YOpCFgmk1Mdg)
  - 慢日志
    - 一般将执行时间大于 100ms 的请求称为慢请求，内核会在执行时统计请求的执行时间，并记录下执行时间大于 100ms 的请求相关信息，打印至内核运行日志中，记录为慢日志
    - MongoDB 中，可以通过以下语句设定 Database Profiler 用于过滤、采集请求，用于慢操作的分析。`db.getProfilingStatus()`
    - 在设置 Profiler 后，满足条件的慢请求将会被记录在 system.profile 表中，该表为一个 capped collection，可以通过 db.system.profile.find() 来过滤与查询慢请求的记录
  - 慢请求的产生无非以下几点原因：
    - CPU负载高：如频繁的认证/建链接会使大量CPU消耗，导致请求执行慢；
    - 等锁/锁冲突：一些请求需要获取锁，而如果有其他请求拿到锁未释放，则会导致请求执行慢；
    - 全表扫描：查询未走索引，导致全表扫描，会导致请求执行慢；
    - 内存排序：与上述情况类似，未走索引的情况下内存排序导致请求执行慢；
    - 但开启分析器 Profiler 是需要一些代价的（如影响内核性能），且一般来说默认关闭，故在处理线上问题时，我们往往只能拿到内核日志中记录的慢日志信息
  - 锁
    - 从5.0开始，将 RESOURCE_PBWM、RESOURCE_RSTL、RESOURCE_GLOBAL 全部归为了 RESOURCE_GLOBAL，且使用一个 enum 对其进行划分
    - 从5.0开始，还新增了一批 lock-free read 操作，这些操作在其他操作持有同 collection 的排他写锁时也不会被阻塞，如 find、count、distinct、aggregate、listCollections、listIndexes 等
    - 在 MongoDB 中为了提高并发效率，提供了类似读写锁的模式，即共享锁（Shared, S）（读锁）以及排他锁（Exclusive, X）（写锁）
    - 为了解决多层级资源之间的互斥关系，提高多层级资源请求的效率，还在此基础上提供了意向锁（Intent Lock）
    ```
    enum LockMode {
    MODE_NONE = 0,
    MODE_IS = 1, //意向共享锁，意向读锁，r
    MODE_IX = 2, // 意向排他锁，意向写锁，w
    MODE_S = 3, // 共享锁，读锁，R
    MODE_X = 4, // 排它锁，写锁，W
    ```
    - 意向锁有什么用呢
      - 如果另一个任务企图在某表级别上应用共享或排他锁，则受由第一个任务控制的表级别意向锁的阻塞，第二个任务在锁定该表前不需要检查各个页或行锁，而只需检查表上的意向锁。
  - MongoDB 的锁矩阵
    ```
     /**
     * MongoDB锁矩阵，可以根据锁矩阵快速查询当前想要加的锁与已经加锁的类型是否冲突
     *
     * | Requested Mode |                      Granted Mode                     |
     * |----------------|:------------:|:-------:|:--------:|:------:|:--------:|
     * |                |  MODE_NONE   | MODE_IS |  MODE_IX | MODE_S |  MODE_X  |
     * | MODE_IS        |      +       |    +    |     +    |    +   |          |
     * | MODE_IX        |      +       |    +    |     +    |        |          |
     * | MODE_S         |      +       |    +    |          |    +   |          |
     * | MODE_X         |      +       |         |          |        |          |
       */
    ```
-  [硬件和操作系统配置](https://mp.weixin.qq.com/s/FKfbg1qw0XAvK_xc_G6u6A)
  - WiredTiger存储引擎的内部缓存大小可以通过storage.wiredTiger.engineConfig.cacheSizeGB进行设置，其大小应足以容纳整个工作集
- Mongo Index
  - MongoDB 底层是如何存储数据的
    - 一个 collection 对应到底层存储引擎就是一个文件，另外每个索引也是单独的文件，每个数据和索引文件的默认结构是 b 树
  - 底层格式存储
    - 在 MongoDB 中设计了 KeyString 结构，将所有类型可以归一化为 string， 然后使用 memcmp 进行二进制比较。
    - 转换成二进制， 优秀的比较性能 / 可以实现不同类型的快速比较；/ 针对数值类型进行细化，解决了整数类型和浮点数类型转换的兼容性问题， 以及节省存储成本。
  - MongoDB 中使用索引查询数据会有 2 个阶段：
    - 查索引，通过索引字段的 KeyString 找到对应的 RecordId；
    - 查数据, 根据 RecordId 找到 BSON 文档；
    - 普通索引的 key 包含 RecordId
  - 索引查询过程
    - IXSCAN 和 FETCH 阶段
    - IXSCAN 通过扫描索引 b 树，返回 RecordId； FETCH 得到 RecordId 后从数据 b 树取出对应的 BSON 文档，直接提交给上层
  





















