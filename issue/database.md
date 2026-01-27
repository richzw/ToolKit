- [The Internals of PostgreSQL](https://www.interdb.jp/pg/index.html)
  - [PostgreSQL Mistakes and How to Avoid Them](https://www.manning.com/books/postgresql-mistakes-and-how-to-avoid-them)
  - [Postgres Internals - Indexes, WAL, MVCC, Locks and Queries](https://gitlab.com/-/snippets/4918687)
  - [Postgres archs ](https://implnotes.pages.dev/postgres/logical_structure)
  - [Postgres SQL Query Roadtrip: Overview](https://internals-for-interns.com/posts/sql-query-roadtrip-in-postgres/)
    - PostgreSQL 执行一条 SQL 从文本到结果的全流程路线图：Parsing → Analysis → Rewriting → Planning → Execution
- Debug high CPU of Postgres
  - We recommend running an `ANALYZE VERBOSE` operation to refresh the `pg_statistic` table.
  - [How can I troubleshoot high CPU utilization for Amazon RDS or Amazon Aurora PostgreSQL](https://aws.amazon.com/premiumsupport/knowledge-center/rds-aurora-postgresql-high-cpu/)
  - There can be many executions from same query which caused the CPU utilization at the time. In order to check that you can follow below steps.
    - Create the pg_stat_statements extension and check the number of calls at the time.
  - if SELECT query having any sorting operations we have to tuning up the work_mem. Below is one of the example how to do that.
    - [Tune up the work_mem and sorting operations](https://aws.amazon.com/blogs/database/tune-sorting-operations-in-postgresql-with-work_mem/)
  - Check query plan
    - `#EXPLAIN (analyze,buffers) SELECT * FROM test ORDER BY name`
    - Note
      - DO NOT set work_mem at the instance level in parameter group
        - The work_mem allocated is not only session specific but actually based on how many sorts are required to complete the request submitted by that session.
        - For example, a complex query could comprise of multiple sorts in a single request and at times, parallel processing could occur with each worker process sorting within its own work_mem space. Based on the volume of data and complexity of the query, it could require lots of memory to reduce the disk overhead and often, there could only be few such queries with requirement for large memory space. Setting a value it at the system level that is only required for fewer queries can create memory bottleneck and cause out of memory situations.
      - As all queries do not require large sort area, it is highly recommended to identify and target those complex queries, and set the optimal value at the session level, then ensure to reset it back to system setting at the end of the request so that the memory is released back to system and used for other activities.
      - In order to estimate the right value for work_mem at the instance level, consider to enable log_temp_files with an initial value, say, 64MB at the instance level, and review the PostgreSQL’s log file to find how many temp files are generated in a day and pick a setting based on the log analysis to minimize disk spills for rest of the queries/requests.
      - DO NOT "EXPLAIN (analyze)" in a busy production system as it actually executes the query behind the scenes to provide more accurate planner information and its impact is significant. Always do the analysis on a non-production system with production quality "data and volume" for accurate estimate of work_mem. The one advantage is that it does not result in network I/O traffic since no output rows are delivered to the client.
  - check possible open long-running transactions
     ```sql 
     SELECT  pid
          , now() - pg_stat_activity.query_start AS duration, query, state
     FROM pg_stat_activity
     WHERE (now() - pg_stat_activity.query_start) >  interval '2 minutes'
     ;
     ```
- PG
  - PG Upgrade
    - running an ANALYZE VERBOSE operation to refresh the pg_statistic table, Optimizer statistics aren't transferred during a major version upgrade, so you need to regenerate all statistics to avoid performance issues. 
    - Performance Tuning https://www.youtube.com/watch?v=XKPHbYe-fHQ&t=1091s
    - Aurora PostgreSQL Query Plan Management https://aws.amazon.com/blogs/database/introduction-to-aurora-postgresql-query-plan-management/
    - Tune query tools https://aws.amazon.com/blogs/database/optimizing-and-tuning-queries-in-amazon-rds-postgresql-based-on-native-and-external-tools/
  - There can be many executions from same query which caused the CPU utilization at the time
    - Create the pg_stat_statements extension and check the number of calls at the time. `create extension pg_stat_statements; SELECT * FROM pg_stat_statements;`
  - Log execution plans
    - https://aws.amazon.com/premiumsupport/knowledge-center/rds-postgresql-tune-query-performance/
    - https://aws.amazon.com/premiumsupport/knowledge-center/rds-postgresql-query-logging/
    - Note: Ensure that you do not set the above parameters at values that generate extensive logging. For example, setting log_statement to "all" or setting log_min_duration_statement to "0" or a very small number can generate a huge amount of logging information. This impacts your storage consumption. If you need to set the parameters to these values, make sure you are only doing so for a short period of time for troubleshooting purposes, and closely monitor the storage space, throughout.
  - SELECT query having any `sorting` operations we have to tuning up the `work_mem`
    - Tune up the work_mem and sorting operations. https://aws.amazon.com/blogs/database/tune-sorting-operations-in-postgresql-with-work_mem/
  - [Vaccum](https://www.percona.com/blog/tuning-autovacuum-in-postgresql-and-autovacuum-internals/)
    - autovacuum will be triggered when the n_dead_tuple of a table meet the threshold. The threshold is calculated using the formula: `vacuum threshold = vacuum base threshold + vacuum scale factor * number of live tuples`
    - For example the command below would make sure the autovacuum be triggered whenever there are 100 rows dead tuple on table users no matter how many live rows the table has: `ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0, autovacuum_vacuum_threshold = 100);`
    - The default value of vacuum setting through `select * from pg_settings where name like '%autovacuum%'` - autovacuum_vacuum_scale_factor, 0.1 autovacuum_vacuum_threshold, 50
    - `ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0.01)`
    - To accelerate the vacuum process is that you could consider to increase the parameter: autovacuum_vacuum_cost_limit value
  - [Using PostgreSQL as a Dead Letter Queue for Event-Driven Systems](https://www.diljitpr.net/blog-post-postgresql-dlq)
    - 在事件驱动系统中处理失败事件（例如：下游 API 不可用、消费者崩溃、字段缺失/格式错误等）的方式：不把失败消息继续留在 Kafka 的 DLQ topic，而是直接落库到 PostgreSQL 的 DLQ 表，以获得更强的可观测性与可操作性（SQL 可查询、可按原因筛选、可批量/定向重试）
    - 一个 DLQ retry scheduler，周期性扫描 PENDING 且“可重试”的事件，批量取出后重放；服务多实例部署时，用 ShedLock 确保同一时刻只有一个实例执行该定时任务，避免重复执行定时扫描逻辑
- Expire
    ```sql
    CREATE TABLE realdata (
       id bigint PRIMARY KEY,
       payload text,
       create_time timestamp with time zone DEFAULT current_timestamp NOT NULL
    ) PARTITION BY RANGE (create_time);
    CREATE VIEW visibledata AS
       SELECT * FROM realdata
       WHERE create_time > current_timestamp - INTERVAL '2 years';
    The view is simple enough that you can INSERT, UPDATE and DELETE on it directly; no need for triggers.
    Now all data will automagically vanish from visibledata after two years.
    ```
- 一次数据库连接池满问题排查与解决
  - Issue
    - 虽然乐观锁是不需要加锁的，通过CAS的方式进行无锁并发控制进行更新的
    - 但是InnoDB的update语句是要加锁的。当并发冲突比较大，发生热点更新的时候，多个update语句就会排队获取锁。
    - 这个排队的过程就会占用数据库链接，一旦排队的事务比较多的时候，就会导致数据库连接被耗尽。
  - Solution
    - 1、基于缓存进行热点数据更新，如Redis。
    - 2、通过异步更新的方式，将高并发的更新削峰填谷掉。
    - 3、将热点数据进行拆分，分散到不同的库、不同的表中，减少并发冲突。
    - 4、合并更新请求，通过打批执行的方式来降低冲突。
- [MySQL 隐式转换](https://mp.weixin.qq.com/s/5GYdHWlfi2-gA3fEIhZaBw)
  - 尽量要避免隐式转换，因为一旦发生隐式转换除了会降低性能外， 还有很大可能会出现不期望的结果. 之所以性能会降低，还有一个原因就是让本来有的索引失效。
- [PostgreSQL@K8s 性能优化]
  - WAL 清零对 PG 的性能和稳定性都有比较大的影响，如果文件系统支持清零特性，可以关闭 wal_init_zero 选项，可有效降低 CPU 和 TPS 抖动。
  - full_page_write 对 PG 的性能和稳定性也有比较大的影响，如果能从存储或备份上能保证数据的安全性，可以考虑关闭，可有效降低 CPU 和 TPS 抖动。
  - 增加 WAL segment size 大小，可降低日志轮转过程中加锁的频率，也可以降低 CPU 和 TPS 抖动，只是效果没那么明显。
  - PG 是多进程模型，引入 pgBouncer 可支持更大的并发链接数，并大幅提升稳定性，如果条件允许，可以开启 Huge Page，虽然原理不同，但效果和 pgBouncer 类似。
  - PG 在默认参数下，属于 IO-Bound，在经过上述优化后转化为 CPU-Bound。
- [数据库死锁排查思路](https://mp.weixin.qq.com/s/gKcAEzf4pMmSJ-FHrDrePA)
  - 死锁场景现场
    - 把业务礼物表A的数据删除，然后修改用户ID后，然后插入到礼物B表。其中，A表和B表，表示同一个礼物逻辑表下的不同分表。
    - 既然是死锁，为什么出现的却是Lock wait timeout exceeded; try restarting transaction 锁等待超时这个日志呢？这是因为在Innodb存储引擎中，当检测到死锁时，它会尝试自动解决死锁问题，通常是通过回滚(rollback)其中的一个或者多个事务来解除死锁。
  - 死锁排查思路
    - 用show engine innodb status，查看最近一次死锁日志。
    - 分析死锁日志，找到关键词TRANSACTION
    - 分析死锁日志，查看正在执行的SQL
    - 看它持有什么锁，等待什么锁。
- [SQL优化](https://mp.weixin.qq.com/s/VKccZmx9Gd6RgsyDEBbBQA)
  - 数据倾斜
  - 分桶解决大表与大表的关联
  - BitMap在多维汇总中的应用
- Postgresql数据库里的索引使用情况
  ```sql
  SELECT 
      relname AS table_name, 
      indexrelname AS index_name, 
      idx_scan as times_index_used, 
      pg_size_pretty(pg_relation_size(indexrelname::regclass)) as index_size
  FROM 
      pg_stat_user_indexes 
  ORDER BY 
      idx_scan DESC;
  ```
- [SQL优化](https://mp.weixin.qq.com/s/_s0hpGKSzSrSrHSUaLkIvA)
  - Case
    - 一张小表A，里面存储了一些ID，大约几百个. 有一张日志表B，每条记录中的ID是来自前面那张小表的，但不是每个ID都出现在这张日志表中
    - 那么我怎么快速的找出今天没有出现的ID呢
  - Solution
    - 递归查询，A表全扫，B表索引扫描了若干次（若干 = 唯一AID在B中出现的次数）。
      - 对B的取值区间，做递归的收敛查询，然后再做NOT IN就很快
        ```sql
        explain (analyze,verbose,timing,costs,buffers) 
          select a.id from a left join b on (a.id=b.aid) where b.* is null;
        
        explain (analyze,verbose,timing,costs,buffers) 
        select * from a where id not in 
        (
        with recursive skip as (  
          (  
            select min(aid) aid from b where aid is not null  
          )  
          union all  
          (  
            select (select min(aid) aid from b where b.aid > s.aid and b.aid is not null)   
              from skip s where s.aid is not null  
          )  -- 这里的where s.aid is not null 一定要加,否则就死循环了.  
        )   
        select aid from skip where aid is not null
        );
        ```
    - SUB QUERY，A表全扫，B表索引扫描了若干次（若干 = A表记录数）
      - 采用sub query，A表数据量小，查询A表的QUERY中使用SUB QUERY使得SUB QUERY的扫描次数下降到与A行数一致，SUB QUERY中采用LIMIT 1限定返回数，is null限定得出B表中未出现的aid
      ```sql
      explain analyze 
      select * from 
      (
        select 
          a.* ,  
          (select aid from b where b.aid=a.id limit 1) as aid   -- sub query, limit 1控制了扫描次数
        from a   -- a表很小  
      ) as t 
      where t.aid is null;  
      ```
- [In-memory disk for PostgreSQL temporary files](https://pgstef.github.io/2024/05/20/in_memory_tmp_files.html)
  - while debugging a performance issue of a CREATE INDEX operation, I was reminded that PostgreSQL might produce temporary files when executing a parallel query, 
  - including parallel index creation, because each worker process has its own memory and might need to use disk space for sorting or hash tables.
- [Postgresql Index](https://www.bestblogs.dev/article/d84aef)

  | Index Type       | Efficient for                    | Advantages                                  | Use Cases                                               |
  |------------------|----------------------------------|---------------------------------------------|---------------------------------------------------------|
  | B-Tree           | Equality and range searches      | Supporting ordering and searching           | Single-column indexes on frequently used columns         |
  | Hash             | Very fast for equality comparisons | Constant time search                        | Composite indexes on multiple columns in WHERE clauses  |
  | GIST             | Support indexing of complex data types (e.g., geometric objects) | Allows for custom indexing methods | Exact match lookups                                     |
  | GIN              | Optimized for indexing arrays and full-text search data types | Supports advanced search operations | Not suitable for queries or sorting                     |
  | BRIN             | Compact representation of data blocks | Efficient for very large tables          | Full-text search                                        |
  | Single-column    | Supports advanced search operations |                                             |                                                         |

- MySQL千万级数据查询优化
  - 优化MySQL千万级数据策略
    - 分表分库
    - 创建中间表，汇总表
    - 修改为多个子查询
  
- [图数据库如何实现多跳查询](https://mp.weixin.qq.com/s/4itqPINGnHLxzEpyDT7ZwQ)
  - 基于大规模并行处理（MPP）的理念，开发了一种图数据库上的分布式并行查询框架，成功将多跳查询的时延降低了 50% 以上
- Blink Tree
  - 在峰值交易场景中，会有大量涉及热点Page的更新及访问，会导致大量关于这些热点Page的SMO（Split Merge Operation）操作，之前PolarDB在SMO场景下由于B+Tree实现有如下的加锁限制：
    - 同一时刻，整个B+Tree只能有一个SMO操作。
    - 正在执行SMO操作的B+Tree分支上的读取操作会被阻塞，直到整个SMO操作完成。
  - 针对这个问题PolarDB做了如下优化：
    - 通过优化加锁，支持同一时刻多个SMO同时进行操作，这样原本等待在其它分支执行SMO的插入操作就无需等待，从而提高写入性能；
    - 引入Blink Tree来替换B+Tree，通过缩小SMO的加锁粒度，将原本需要将所有涉及SMO的各层Page加锁直到整个SMO完成后才释放的逻辑，优化成Ladder Latch，即逐层加锁。
- 数据库查询优化
  - 局部性原理又表现为：时间局部性和空间局部性
    - 堆排序有着最强的理论性能——最坏时间复杂度 O(NlgN)，但在实际应用中却干不过最坏时间复杂度 O(N^2) 的快速排序，便是由于堆排序的访存空间局部性太差
  - 要想利用空间局部性的关键就在于：缓存的数据加载粒度 > 缓存的数据查询粒度
    - 如何定义“相邻”以及如何确定缓存加载的最小数据单位
    - 如何定义“相邻”：在这个案例中，无论是按照 (userId, graphId) 还是 (userId) 的粒度来进行缓存加载，其都是符合数据库主键的最左前缀匹配的
    - 缓存加载最小数据单位：不同于 CPU 中的缓存行，面临晶体管数量等严格的物理限制，在分布式系统中，我们面临的物理限制通常来自于内存大小，而这类限制相比之下宽松得多
      - 我们可以以 (userId, graphId) 甚至 (userId) 这样的动态粒度来作为缓存加载的最小数据单位，从而将空间局部性挖掘到极致
- [慢查询分析](https://mp.weixin.qq.com/s/qeH5YYSBYh6cTEf4Cya2nA)
  - 衡量查询开销的三个指标
    - 响应时间：服务时间和排队时间之和。服务时间是指数据库处理这个查询真正花了多长时间 - 可能是等I/O操作完成，也可能是等待行锁。
    - 扫描的行数：一条查询，如果性能很差，最常见的原因是访问的数据太多。大部分性能低下的查询都可以通过减少访问的数据量的方式进行优化
    - 返回的行数：会给服务器带来额外的I/O、内存和CPU的消耗（使用limit限制返回行数）
- [Multi-Version Concurrency Control (MVCC) in PostgreSQL](https://www.red-gate.com/simple-talk/databases/postgresql/multi-version-concurrency-control-mvcc-in-postgresql-learning-postgresql-with-grant/)
  - The PostgreSQL database management system has three ways, isolation levels, for dealing with concurrency:
    - Read Committed
    - Repeatable Read
    - Serializable
- [数据库迁移全流程](https://mp.weixin.qq.com/s/xKr9k7uSILk4q64zIzJ3dA)
  - 三板斧(灰度/监控/回滚)
    - 可监控(数据对比读逻辑) 可监控(对比读逻辑) 可灰度(灰度切量读) 可回滚(灰度切量写)
- [Postgresql 18 async io](https://pganalyze.com/blog/postgres-18-async-io)
  - Asynchronous I/O support in Postgres 18 introduces worker (as the default) and io_uring options under the new io_method setting
  - Observability practices need to evolve: EXPLAIN ANALYZE may underreport I/O effort, and new views like pg_aios will help provide insights
- [使用一写多读的未分片架构，证明了 PostgreSQL 在海量读负载下也可以伸缩自如](https://mp.weixin.qq.com/s/ykrasJ2UeKZAMtHCmtG93Q)
  - OpenAI 使用 Azure 上的托管数据库，没有使用分片与Sharding，而是一个主库 + 四十多个从库的经典 PostgreSQL 主从复制架构
  - PostgreSQL 的 MVCC 设计存在一些已知的问题，
    - 例如表膨胀与索引膨胀，自动垃圾回收调优较为复杂，每次写入都会产生一个完整新版本，索引访问也可能需要额外的回表可见性检查。
    - 这些设计会带来一些 “扩容读副本” 的挑战：更多 WAL 通常会导致复制延迟变大，而且当从库数量疯狂增长时，网络带宽可能成为新的瓶颈。
  - 优化
    - 抹平主库上的写尖峰，尽可能的减少主库上的负载： 使用惰性写入来尽可能抹平写入毛刺
    - 查询层面进行优化： 长事务会阻止垃圾回收并消耗资源，因此他们使用 timeout 配置来避免 Idle in Transaction 长事务；使用 ORM 容易导致低效的查询，应当慎用。
    - 治理单点问题： 有许多只读从库，一个挂了应用还可以读其他的。实际上许多关键请求是只读的，所以即使主库挂了，它们也可以继续从主库上读取。
      - 低优先级请求与高优先级请求也进行了区分，对于那些高优先级的请求，OpenAI 分配了专用的只读从库，避免它们被低优先级的请求影响
    - 只允许在此集群上进行轻量的模式变更
  - PostgreSQL 开发者社区提出了一些问题与特性需求
    - 禁用索引问题的，不用的索引会导致写放大与额外的维护开销，他们希望移除没用的索引，然而为了最小化风险
      - PostgreSQL 其实是有禁用索引的功能的，只需要更新 pg_index 系统表中的 indisvalid 字段为 False，这个索引就不会被 Planner 使用，但仍然会在 DML 中被继续维护
    - 关于可观测性的，目前的 pg_stat_statement 只提供每类查询的平均响应时间，而没法直接获得 （p95, p99）延迟指标。他们希望拥有更多类似 histogram 与 percentile 延迟的指标。
    - PostgreSQL 默认参数的优化建议，PostgreSQL 默认参数值过于保守了
- WALs aren't just a recovery mechanism; they're a design principle
  - WAL logs every change before applying it, which ensures durability and crash safety
  - Writes are acknowledged only after the log hits persistent storage
  - WAL is used for crash recovery, replication, and backup
  - PostgreSQL, MongoDB, Kafka, and many other systems rely on WAL-like designs
- Postgresql 18 async io
  - io_method：选择异步 I/O 实现方式，取值 sync（与 PG17 一致的同步行为）、worker（由专门的 I/O worker 进程代执行）、io_uring（Linux 5.1+ 的高性能异步 I/O，需要编译启用 liburing）。该参数需重启生效。(postgresql.org, postgresqlco.nf)
    - io_workers：当 io_method=worker 时，控制 I/O worker 进程数（默认 3）。
    - io_max_concurrency：限制单进程同时进行的 I/O 数量上限
    - io_combine_limit / io_max_combine_limit：限制合并读写的最大 I/O 大小（默认通常 128kB，可在 18 中放宽上限以便试验更大的合并 I/O）。(postgresql.org, postgresqlco.nf)
    - effective_io_concurrency、maintenance_io_concurrency：在具备“预取建议”支持的平台上，控制预取距离/并发度；PG18 文档默认值提升为 16
  - 支持范围与收益（概览）
    - 以读路径为主：顺序扫描、位图堆扫描以及 VACUUM 等维护读操作首先受益；写入（含 WAL）仍以同步为主。实际测试在冷缓存、云盘/高时延存储上提升最明显
    - io_uring 一般优于 worker（更少上下文切换/系统调用开销），但要求 Linux 5.1+ 且构建包含 liburing
- [OLTP数据库能够如此快速地查找数据](https://mp.weixin.qq.com/s/xf1NTIyioE5oy7TNFxwYug)
  - OLTP 数据库利用 B+Tree 索引 在磁盘之上高效实现“快速定位单条记录”：
    - 每层节点容纳大量有序键，对应多个子指针；
    - 树的高度小，查找需要的页面访问次数极少；
    - 分裂/合并过程保证树始终平衡、节点半满，从而优化 I/O 和空间。
  - 与基于内存的 BST 相比，B+Tree 明显地：
    - 更适配磁盘块读写模型；
    - 减少随机磁盘 I/O 次数；
    - 兼顾读和写性能。
- [Robust Database Backup Recovery at Uber](https://www.uber.com/en-HK/blog/robust-database-backup-recovery-at-uber/?uclick_id=1710c187-26ed-4c1b-9134-278f616291e7)
  - Uber 在 Stateful Platform 上构建了 **统一的、技术无关的快照式备份与恢复体系**，核心是：
    - **Continuous Backup（Time Machine）**
    - **Backup/Restore Framework + 技术插件**
    - **Continuous Restore（持续恢复演练与验证）**
- [PG Waiting for Commit](https://www.postgresql.eu/events/pgconfeu2025/schedule/session/6990-waiting-for-commit/#slides)
  - 提交延迟分析
    - 理想情况（50µs 磁盘延迟，200µs 网络延迟）： 总延迟约 360µs（对比本地 50µs） ; 
    - 云环境（500µs 磁盘延迟，500µs 跨可用区网络延迟）： 总延迟约 1,560µs（对比本地 500µs）
  - 延迟的重要性
    - 高争用工作负载受影响：5ms 提交等待 = 每秒仅 200 次更新
    - 连接池受影响：每增加 1ms 延迟，每 1000 TPS 就需要多 1 个连接
  - synchronous_commit 的级别
    - off：仅靠"祈祷"模式
    - local：可以在崩溃后恢复
    - remote_write：可以在故障转移后恢复
    - on：可以在崩溃和故障转移后恢复
    - remote_apply：读写一致性
  - 问题诊断方法
    - 使用 pg_stat_statements 跟踪提交延迟
    - 关注等待事件：
      - Io/WalSync：同步 WAL 到本地磁盘
      - LWLock/WALWrite：等待他人同步 WAL
      - Ipc/SyncRep：等待备库确认
    - 使用 pg_wait_sampling 获取高分辨率数据
  - 解决方案
    1. 购买更好的存储设备
    2. 仲裁提交（Quorum Commit）：
    - 使用多个副本隐藏尾部延迟
    - 至少需要 2 个副本
    - 配置示例：synchronous_standby_names = 'ANY 1 (node2, node3, node4)'
- Postgresql 18
  - [EXPLAIN 新字段：Index Searches](https://www.pgmustard.com/blog/what-do-index-searches-in-explain-mean)
    - Index Searches 是什么
      - 在 **Postgres 18** 中，`EXPLAIN ANALYZE` 的索引相关算子（Index Scan / Index Only Scan / Bitmap Index Scan 等）新增一行：`Index Searches: N`。
      - 这个数字表示：执行该节点时，对**同一棵索引树进行了多少次独立的索引下降（index descent）**。
      - `Index Searches: 1` 是“普通情况”，只下降一次；`>1` 则意味着使用了多次下降的优化策略。
    - Index Searches > 1 的两类优化
      - 1. **IN/ANY 优化（Postgres 17 引入，18 中可见）**
        - 对 `WHERE col IN (v1, v2, ...)` 这类条件，btree 可以对每个值做一次单独的索引搜索，而不是扫描整个区间。
        - 在 Postgres 18 的 `EXPLAIN ANALYZE` 中，可以看到 `Index Searches` 等于 IN 列表中值的数量，比如 4。
        - 以前只能通过时间、缓冲区读（buffers）推断是否用了优化，现在直接可见；统计视图中 `pg_stat_user_indexes.idx_scan` 也在计这些下降次数。 
      - 2. **btree skip scan（Postgres 18 新增）**
        - 针对多列索引 `(col1, col2)`，当只对 `col2` 有选择性过滤，而 `col1` 取值较少时，规划器可以对不同的 `col1` 值做多次索引下降，而不是从头到尾扫整棵索引。
        - 例如：索引 `(four, unique1)`，查询 `four BETWEEN 1 AND 3 AND unique1 = 42` 时，`Index Searches: 3`，分别对应 four=1、2、3 的三次下降，比从 four=1 扫到 four=3 的整个范围要高效得多
    - 多次 Index Searches 好还是坏？
      - **理想情况**：有一个“完美匹配”的索引，只需一次下降（`Index Searches: 1`）就能读到最少的缓冲区（buffers），性能最好。
      - 但为每个重要查询都建一个理想索引成本极高：
        - 写入放大（更多索引要维护）；
        - 可能失去原本未建索引列上的 HOT 更新机会；
        - 额外索引会抢占 `shared_buffers` 空间。
      - 因此：
        - 对“普通”或不太重要的查询，多次 Index Searches 往往是利用现有索引的**自动优化**，是好事；
        - 对**非常关键且愿意专门建索引的查询**，若看到 `Index Searches > 1`，通常说明**存在更优的索引设计（比如列顺序）**。
    - 实用建议（作者观点）
      1. **看到 Index Searches > 1 时**：
        - 对关键查询，检查是否有更合适的索引设计（特别是列顺序）能把它降到 1。
      2. **总体调优时**：
        - 仍然把重心放在：行数、过滤比例、buffers、时间上，Index Searches 只是辅助信号。
      3. **考虑升级到 Postgres 18**：
        - 许多现有查询在不改 SQL、不改索引的情况下就能自动获益于 skip scan / 多次下降优化。
        - 新优化可能让“同列不同顺序”的一些索引变得冗余，可以评估是否删掉部分重叠索引，而对读延迟影响仍可接受
  - [Skip Scan - 摆脱最左索引限制](https://www.pgedge.com/blog/postgres-18-skip-scan-breaking-free-from-the-left-most-index-limitation)
    - B-tree Skip Scan，它解决了多列 B-tree 索引长期以来的“必须从最左前缀列开始用起”的问题，并简要提到 RETURNING 子句的增强
    - Skip Scan 能力，使即便缺失前导列的等值条件，也仍然可以利用多列索引
      - 当查询按索引后面的列（如 customer_id、order_date）做条件过滤，而没有约束最左列（如 status）时：
        - 优化器可以：
        - 找出被省略前导列的所有不同取值（如 status 有 pending、active、shipped）。
        - 把查询逻辑上等价为多个子查询的并集，每个子查询都带上一个具体的前导列值，然后使用现有的索引扫描机制。
    - Skip Scan 在以下情况最有用：
      - 前导列基数低（Low Cardinality）
        - 省略的前导列（如 status、region）只有少数几个不同值（例如 3～5 个）。
        - 优化器只需对少数前导值执行多次索引扫描，总代价仍远小于顺序扫描。
        - 若前导列有成千上万种不同值，Skip Scan 成本会暴涨，收益大幅下降。
      - 索引后面列上有等值条件
        - 当前实现主要针对后续列存在等值条件的场景（如 customer_id = 123、product_category = 'Category_5'）
- PG vs Mysql
  - 索引与存储结构
    - MySQL（默认指 InnoDB）
      - 使用 聚簇索引（clustered index） 存储表数据：
        - 主键索引的 B+ 树叶子节点中，存的是“主键值 + 整行记录（所有列）”。
        - 因此通过主键查询时，只需一次 B+ 树查找即可拿到整行数据。
      - 二级索引（secondary index）：
        - 叶子节点存的是“二级索引键值 + 主键值”。
        - 查询流程：先通过二级索引找到主键值，再到聚簇索引中根据主键值查找整行数据（第二次 B+ 树查找）。
    - PostgreSQL
      - 表是 堆组织表（heap organized table）：
        - 表数据存放在堆中，物理上基本无序，不按主键或任何索引顺序排列。
      - 所有 B+ 树索引本质上都是“二级索引”：
        - 索引中存储的是“键值 + 元组位置（TID/c_tid）”。
        - 通过索引查到 TID 后，再去堆中做一次“堆访问（heap fetch）”取出整行数据。
      - 因为更新会产生新版本（MVCC），行在堆中的物理位置可能越来越分散。
  - 主键类型对索引大小的影响
    - MySQL（InnoDB）
      - 所有二级索引的叶子记录中都存有主键值。
      - 若主键使用 UUID 等“又长又随机”的类型：
      - 所有二级索引都会变“胖”，占空间更大，读 I/O 也增加。
    - PostgreSQL
      - 二级索引中存的是 TID（指向堆的定位信息），大小固定，不随主键列类型变化。
      - 因此主键是 UUID 对二级索引体积影响不大。
  - 主键范围查询
    - MySQL
      在主键聚簇索引上：先查到起点 PK=1，然后沿叶子链表顺序扫描到 PK=3。
      叶子节点本身包含完整行数据，顺序范围扫描特别高效。
    - PostgreSQL
      在主键 B+ 树索引（其实也是普通 btree 索引）上做范围扫描，得到一串 TID。
      然后对这些 TID 逐个做堆访问，堆中的行可能分布在多个页面上，且更新较多时更分散。
      对写频繁表，这是一个弱点；通过设置合适 FILLFACTOR（填充因子），预留页面空间，可增加在同一页面内作 HOT 更新的概率，减缓碎片
- [PostgreSQL 19](https://mp.weixin.qq.com/s/Q9qEKwVDftyGjB4BgIdyJw)
  - 并行 TID Range Scan（并行 ctid 范围扫描
    - 什么是 Tid Range Scan（ctid 范围扫描）
      - ctid 是 PostgreSQL 的系统列，表示元组（tuple）在堆表中的物理位置（页号 + 页内偏移）。
      - Tid Range Scan 是一个专门的执行器节点：当 WHERE 子句中出现对 ctid 的“范围条件”时，规划器可生成 Tid Range Scan 路径，对指定的 TID/页面范围做更“定向”的堆扫描。
    - Tid Range Scan 本身无法并行，因此在大表范围过滤查询里，规划器容易遇到冲突：
      - Tid Range Scan：I/O 更精准（只读目标范围块），但不并行（CPU 并发差）。
      - Parallel Seq Scan：CPU 可以并发，但会扫描范围外的块，I/O 浪费。
    - PG19（master）修复：并行化 + chunk 衰减分配
      - Tid Range Scan 允许并行化：把需要扫描的块范围像 Parallel Seq Scan 那样“分发”给 worker。
      - chunk 分配采用“逐步衰减到 1 block”的策略：开始分大块减少共享状态/锁竞争；越接近结束 chunk 越小，避免只剩一个 worker 扫尾导致其他 worker 空转，达到更好的负载均衡
- [SQLite3 如果突发断电、关机，数据到底会不会丢](https://mp.weixin.qq.com/s/Z7lv8WzuwIpYTATnY60QHw)
  - SQLite 的断电安全与性能权衡，主要由一组 PRAGMA（尤其 journal_mode/synchronous/WAL checkpoint 相关）决定
- [慢SQL说起：交易订单表如何做索引优化](https://mp.weixin.qq.com/s/sCBOvzUkX7O4fqGTVM68Uw)
  - 非典型慢 SQL：分页 + 复合排序触发 filesort
    - 因为 create_time desc, order_id asc 这组排序无法由“单一索引顺序”直接满足（作者给出的原因：create_time 在二级索引里，order_id 在主键索引里，跨索引无法用于复合排序）
  - 去掉 order_id 排序的尝试与事故复盘：稳定分页的必要性
    - 线上出现“订单重复/订单缺失”的舆情案例。
      - 原因链路：
        - 当排序 key（这里是 create_time）存在大量相同值时，filesort 的相同 key 行的相对顺序不保证稳定；
        - 分页用 limit offset, size 时：
          - 第 1 页与第 2 页分别执行各自的排序
          - 如果两次排序对“create_time 相等的记录”内部顺序不同，则 offset 切片会出现重复/漏项
      - 解决策略：在排序里加入唯一/更稳定的 tie-breaker
        - order by create_time desc, order_id asc
        - 由于二级索引叶子节点包含主键值（InnoDB 二级索引叶子通常带主键），加入 order_id 不一定引入额外回表，但会引入“更重的排序成本”（作者通过压测评估可接受）。
- [Postgresql 约束底层](https://mp.weixin.qq.com/s/q4yHAWhO9Yeav4_DM5YyBA)
  - 无论列/表约束、域约束、约束触发器，本质上都能用系统目录（尤其 pg_constraint，配合 pg_class/pg_type/pg_attribute）串起来理解。
  - 外键“像触发器”并非比喻：在实现层面确实由系统触发器执行检查与联动动作，只是这些触发器被约束系统管理/关联
- [Multiversion Concurrency Control (MVCC): A Practical Deep Dive](https://celerdata.com/glossary/multiversion-concurrency-control)
- [Databases in 2025: A Year in Review](https://www.cs.cmu.edu/~pavlo/blog/2026/01/2025-databases-retrospective.html)
  -  PostgreSQL v18
    - 异步 I/O（asynchronous I/O）存储子系统：被描述为让 PostgreSQL 走向“降低对 OS page cache 依赖”的路径。
    - Skip Scan 支持：允许查询在 多列 B+Tree 索引上，即使缺失“前导列（leading key / prefix）”也仍可能利用索引
  - 列式文件格式（尤其是 Parquet）进入“新一轮格式战争
    - Parquet 的“真正问题”：实现碎片化导致互操作性差
- [Oracle、MySQL和PostgreSQL三大数据库执行计划的区别](https://mp.weixin.qq.com/s/26GoD8Xs5EYrIcqKs5I2rA)
  - 干预方式不同
    - PostgreSQL 只能通过对表进行分析来改变执行计划，不支持通过添加hint的方式干预执行计划
    - Oracle 不仅可以通过对表进行收集统计来改变执行计划，而且支持通过添加hint的方式直接干预执行计划的生成
    - MySQL 虽然支持类似Oracle的hint功能，但其优化器相对简单，对复杂查询的处理能力不如Oracle强大
  - 缓存机制差异
    - Oracle和SQL Server 会自动缓存执行计划，相同的SQL语句（甚至大小写不同都会被当作不同语句）可以重用执行计划，减少解析开销
    - PostgreSQL 并不会自动缓存执行计划，每次执行SQL查询都会从头开始解析、优化生成执行计划。但它在预处理语句和PL/pgSQL函数中会缓存执行计划
  - 查询效率特点
    - PostgreSQL 在单条数据处理、空间查询和转换方面表现出色，支持很多方法函数
    - MySQL 在简单查询和读写操作上表现良好，但在复杂查询和大数据量分析方面不如Oracle
- Partitioning 还是 Sharding
  - Partition
    - 在同一个数据库实例/同一张逻辑表里，把数据按规则切成多个分区”。对应用来说还是一张表（同一个 schema），只是底层存储被分成了多个分区（partition）
    - Partitioning：同库内把表“切块”，更偏“管理 + 单库性能优化”
  - Sharding
    - 一般指“把同一张逻辑表的数据，拆到多个数据库实例/多台机器上”
    - Sharding：跨库/跨机把数据“拆家”，更偏“水平扩展 + 吞吐/容量上限”



















