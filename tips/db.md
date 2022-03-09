
- [如何解决MySQL深分页问题](https://juejin.cn/post/7012016858379321358)
  - Issue
    - 我们日常做分页需求时，一般会用limit实现，但是当偏移量特别大的时候，查询效率就变得低下
  - Why

    深分页的执行SQL如下：

    `select id,name,balance from account where update_time> '2020-09-19' limit 100000,10;`

    我们先来看下这个SQL的执行流程：
    - 通过普通二级索引树idx_update_time，过滤update_time条件，找到满足条件的记录ID。
    - 通过ID，回到主键索引树，找到满足记录的行，然后取出展示的列（回表）
    - 扫描满足条件的100010行，然后扔掉前100000行，返回。
    
    SQL变慢原因有两个：
    - limit语句会先扫描offset+n行，然后再丢弃掉前offset行，返回后n行数据。也就是说limit 100000,10，就会扫描100010行，而limit 0,10，只扫描10行。
    - limit 100000,10 扫描更多的行数，也意味着回表更多的次数。

  - Solution
    - 通过子查询优化, 可以通过减少回表次数来优化, 把条件转移到主键索引树
      ```sql
      select id,name,balance FROM account 
      where id >= (
           select a.id from account a 
           where a.update_time >= '2020-09-19' 
            limit 100000, 1)
       LIMIT 10;（可以加下时间条件到外面的主查询）
      ```
    - INNER JOIN 延迟关联
      ```sql
      SELECT  acct1.id,acct1.name,acct1.balance 
      FROM account acct1 
         INNER JOIN 
         (SELECT a.id 
          FROM account a 
         WHERE a.update_time >= '2020-09-19' ORDER BY a.update_time 
         LIMIT 100000, 10) AS  acct2 
      on acct1.id= acct2.id;
      ```
      先通过idx_update_time二级索引树查询到满足条件的主键ID，再与原表通过主键ID内连接，这样后面直接走了主键索引了，同时也减少了回表
    - 标签记录法
      - limit 深分页问题的本质原因就是：偏移量（offset）越大，mysql就会扫描越多的行，然后再抛弃掉。这样就导致查询性能的下降。
      - 可以采用标签记录法，就是标记一下上次查询到哪一条了，下次再来查的时候，从该条开始往下扫描。就好像看书一样，上次看到哪里了，你就折叠一下或者夹个书签，下次来看的时候，直接就翻到啦。
      - 这种方式有局限性：需要一种类似连续自增的字段。
    - 使用between...and...
      - 可以将limit查询转换为已知位置的查询，这样MySQL通过范围扫描between...and，就能获得到对应的结果
      
- [Redis主节点的Key已过期，但从节点依然读到过期数据](https://mp.weixin.qq.com/s?__biz=Mzg2NzYyNjQzNg==&mid=2247487738&idx=1&sn=e7e6b10b81736ba9775f485ce39585a6&chksm=ceb9ec51f9ce6547ce6378692b11cc09d3c46ded964bddf5052373b1c8cbf5b02a2327f4d23e&scene=132#wechat_redirect)
  - 大部分的业务场景都是读多写少，为了利用好这个特性，提升Redis集群系统的吞吐能力，通常会采用主从架构、读写分离
  - 主从架构的风险
    - 拉取过期数据
      - 通常会设置过期时间，对于一些使用不是很频繁的数据，会定期删除，提高资源的利用率
        - 被动删除，当数据过期后，并不会马上删除。而是等到有请求访问时，对数据检查，如果数据过期，则删除数据
        - 定期删除。每隔一段时间，默认100ms，Redis会随机挑选一定数量的Key，检查是否过期，并将过期的数据删除。
      - 如果读从库，则有可能拿到过期数据。原因有两个
        - 跟 Redis 的版本有关系，Redis 3.2 之前版本，读从库并不会判断数据是否过期，所以有可能返回过期数据。
        - 跟过期时间的设置方式有关系，我们一般采用 EXPIRE 和 PEXPIRE，表示从执行命令那个时刻开始，往后延长 ttl 时间。严重依赖于 开始时间 从什么时候算起
          ![img.png](db_redis_expire.png)
        - 解决方案： 可以采用Redis的另外两个命令，EXPIREAT 和 PEXPIREAT，相对简单，表示过期时间为一个具体的时间点。避免了对开始时间从什么时候算起的依赖
        - EXPIREAT 和 PEXPIREAT 设置的是时间点，所以要求主从节点的时钟保持一致，需要与NTP 时间服务器保持时钟同步。
    - 主从节点数据不一致
      - 从库同步落后的原因主要有两个：
        - 1、主从服务器间的网络传输可能有延迟
        - 2、从库已经收到主库的命令，由于是单线程执行，前面正在处理一些耗时的命令（如：pipeline批处理），无法及时同步执行。
      - 解决方案：
        - 1、主从服务器尽量部署在同一个机房，并保持服务器间的网络良好通畅
        - 2、监控主从库间的同步进度，通过`info replication`命令 ，查看主库接收写命令的进度信息（master_repl_offset），从库的复制写命令的进度信息（slave_repl_offset）
          `master_repl_offset - slave_repl_offset` 得到从库与主库间的复制进度差. 我们可以开发一个监控程序，定时拉取主从服务器的进度信息，计算进度差值。如果超过我们设置的阈值，则通知客户端断开从库的连接，全部访问主库，一定程度上减少数据不一致情况。
- [Redis 用数据类型实现亿级数据统计](https://mp.weixin.qq.com/s?__biz=Mzg2NzYyNjQzNg==&mid=2247487680&idx=1&sn=7877648fac1fe8bf98b65bdeaf50a7ea&chksm=ceb9ec6bf9ce657df72a03491b2532d2c3bf3714a244459eb0609754a4f39ae496c1c9e0e584&scene=132#wechat_redirect)
  
  常见的场景如下：
  - 给一个 userId ，判断用户登陆状态；
  - 两亿用户最近 7 天的签到情况，统计 7 天内连续签到的用户总数；
  - 统计每天的新增与第二天的留存用户数；
  - 统计网站的对访客（Unique Visitor，UV）量
  - 最新评论列表

- [详谈水平分库分表](https://mp.weixin.qq.com/s/vqYRUEPnzFHExo4Ly7DPWw)
  - 什么是一个好的分库分表方案
    - 方案可持续性 - 业务数据量级和业务流量未来进一步升高达到新的量级的时候，我们的分库分表方案可以持续使用
    - 数据偏斜问题 - 定义分库分表最大数据偏斜率为 ：（数据量最大样本 - 数据量最小样本）/ 数据量最小样本。一般来说，如果我们的最大数据偏斜率在5%以内是可以接受的
  - 常见的分库分表方案
    - Range分库分表
      - TiDB数据库，针对TiKV中数据的打散，也是基于Range的方式进行，将不同范围内的[StartKey,EndKey)分配到不同的Region上
      - 该方案的缺点：
        - 最明显的就是数据热点问题
        - 新库和新表的追加问题
        - 业务上的交叉范围内数据的处理
        - 通过年份进行分库分表，那么元旦的那一天，你的定时任务很有可能会漏掉上一年的最后一天的数据扫描
    - Hash分库分表
      - 几个常见的错误案例
        - 非互质关系导致的数据偏斜问题
          ```go
          public static ShardCfg shard(String userId) {
              int hash = userId.hashCode();
              // 对库数量取余结果为库序号
              int dbIdx = Math.abs(hash % DB_CNT);
              // 对表数量取余结果为表序号
              int tblIdx = Math.abs(hash % TBL_CNT);
           
              return new ShardCfg(dbIdx, tblIdx);
          }
          ```
          发现，以10库100表为例，如果一个Hash值对100取余为0，那么它对10取余也必然为0。
          事实上，只要库数量和表数量非互质关系，都会出现某些表中无数据的问题。
          ![img.png](db_shard_table.png)
          当然，如果分库数和分表数不仅互质，而且分表数为奇数(例如10库101表)，则理论上可以使用该方案
        - 扩容难以持续

          我们把10库100表看成总共1000个逻辑表，将求得的Hash值对1000取余，得到一个介于[0，999)中的数，然后再将这个数二次均分到每个库和每个表中，大概逻辑代码如下
          ```go
          public static ShardCfg shard(String userId) {
                  // ① 算Hash
                  int hash = userId.hashCode();
                  // ② 总分片数
                  int sumSlot = DB_CNT * TBL_CNT;
                  // ③ 分片序号
                  int slot = Math.abs(hash % sumSlot);
                  // ④ 计算库序号和表序号的错误案例
                  int dbIdx = slot % DB_CNT ;
                  int tblIdx = slot / DB_CNT ;
           
                  return new ShardCfg(dbIdx, tblIdx);
              }
          ```
          该方案确实很巧妙的解决了数据偏斜的问题，只要Hash值足够均匀，那么理论上分配序号也会足够平均. 但是该方案有个比较大的问题，那就是在计算表序号的时候，依赖了总库的数量，那么后续翻倍扩容法进行扩容时，会出现扩容前后数据不在同一个表中，从而无法实施
      - 几种Hash分库分表的方案
        - 标准的二次分片法
          ```go
          public static ShardCfg shard2(String userId) {
                  // ① 算Hash
                  int hash = userId.hashCode();
                  // ② 总分片数
                  int sumSlot = DB_CNT * TBL_CNT;
                  // ③ 分片序号
                  int slot = Math.abs(hash % sumSlot);
                  // ④ 重新修改二次求值方案
                  int dbIdx = slot / TBL_CNT ;
                  int tblIdx = slot % TBL_CNT ;
           
                  return new ShardCfg(dbIdx, tblIdx);
              }
          ```
          和错误案例二中的区别就是通过分配序号重新计算库序号和表序号的逻辑发生了变化. 那为何使用这种方案就能够有很好的扩展持久性呢？我们进行一个简短的证明：
          ![img.png](db_table_part.png)
          通过翻倍扩容后，我们的表序号一定维持不变，库序号可能还是在原来库，也可能平移到了新库中(原库序号加上原分库数)，完全符合我们需要的扩容持久性方案
          
          缺点：
          - 翻倍扩容法前期操作性高，但是后续如果分库数已经是大几十的时候，每次扩容都非常耗费资源。
          - 连续的分片键Hash值大概率会散落在相同的库中，某些业务可能容易存在库热点（例如新生成的用户Hash相邻且递增，且新增用户又是高概率的活跃用户，那么一段时间内生成的新用户都会集中在相邻的几个库中）。
        - 关系表冗余

          该方案还是通过常规的Hash算法计算表序号，而计算库序号时，则从路由表读取数据。因为在每次数据查询时，都需要读取路由表，故我们需要将分片键和库序号的对应关系记录同时维护在缓存中以提升性能。
          ````go
          public static ShardCfg shard(String userId) {
                  int tblIdx = Math.abs(userId.hashCode() % TBL_CNT);
                  // 从缓存获取
                  Integer dbIdx = loadFromCache(userId);
                  if (null == dbIdx) {
                      // 从路由表获取
                      dbIdx = loadFromRouteTable(userId);
                      if (null != dbIdx) {
                          // 保存到缓存
                          saveRouteCache(userId, dbIdx);
                      }
                  }
                  if (null == dbIdx) {
                      // 此处可以自由实现计算库的逻辑
                      dbIdx = selectRandomDbIdx();
                      saveToRouteTable(userId, dbIdx);
                      saveRouteCache(userId, dbIdx);
                  }
           
                  return new ShardCfg(dbIdx, tblIdx);
              }
          ````
          selectRandomDbIdx方法作用为生成该分片键对应的存储库序号，这边可以非常灵活的动态配置。例如可以为每个库指定一个权重，权重大的被选中的概率更高，权重配置成0则可以将关闭某些库的分配。当发现数据存在偏斜时，也可以调整权重使得各个库的使用量调整趋向接近。

          该方案还有个优点，就是理论上后续进行扩容的时候，仅需要挂载上新的数据库节点，将权重配置成较大值即可，无需进行任何的数据迁移即可完成。

          缺点：
          - 每次读取数据需要访问路由表，虽然使用了缓存，但是还是有一定的性能损耗。
          - 路由关系表的存储方面，有些场景并不合适。例如上述案例中用户id的规模大概是在10亿以内，我们用单库百表存储该关系表即可。但如果例如要用文件MD5摘要值作为分片键，因为样本集过大，无法为每个md5值都去指定关系（当然我们也可以使用md5前N位来存储关系）。
          - 饥饿占位问题
        - 基因法
          - 我们发现案例一不合理的主要原因，就是因为库序号和表序号的计算逻辑中，有公约数这个因子在影响库表的独立性。
          - 我们计算库序号的时候做了部分改动，我们使用分片键的前四位作为Hash值来计算库序号。
            ```java
            public static ShardCfg shard(String userId) {
                int dbIdx = Math.abs(userId.substring(0, 4).hashCode() % DB_CNT );
                int tblIdx = Math.abs(userId.hashCode() % TBL_CNT);
                return new ShardCfg(dbIdx, tblIdx);
            }
            ```
          - 我们发现该方案中，分库数为16，分表数为100，数量最小行数仅为10W不到，但是最多的已经达到了15W+，最大数据偏斜率高达61%。按这个趋势发展下去，后期很可能出现一台数据库容量已经使用满，而另一台还剩下30%+的容量。
        - 剔除公因数法
          - 在很多场景下我们还是希望相邻的Hash能分到不同的库中。就像N库单表的时候，我们计算库序号一般直接用Hash值对库数量取余
            ```java
            public static ShardCfg shard(String userId) {
                    int dbIdx = Math.abs(userId.hashCode() % DB_CNT);
                    // 计算表序号时先剔除掉公约数的影响
                    int tblIdx = Math.abs((userId.hashCode() / TBL_CNT) % TBL_CNT);
                    return new ShardCfg(dbIdx, tblIdx);
            }
            ```
          - 经过测算，该方案的最大数据偏斜度也比较小，针对不少业务从N库1表升级到N库M表下，需要维护库序号不变的场景下可以考虑
        - 一致性Hash法
          - 正规的一致性Hash算法会引入虚拟节点，每个虚拟节点会指向一个真实的物理节点。这样设计方案主要是能够在加入新节点后的时候，可以有方案保证每个节点迁移的数据量级和迁移后每个节点的压力保持几乎均等。
          - 但是用在分库分表上，一般大部分都只用实际节点，引入虚拟节点的案例不多，主要有以下原因：
            - 应用程序需要花费额外的耗时和内存来加载虚拟节点的配置信息。如果虚拟节点较多，内存的占用也会有些不太乐观。
            - 由于mysql有非常完善的主从复制方案，与其通过从各个虚拟节点中筛选需要迁移的范围数据进行迁移，不如通过从库升级方式处理后再删除冗余数据简单可控。
            - 虚拟节点主要解决的痛点是节点数据搬迁过程中各个节点的负载不均衡问题，通过虚拟节点打散到各个节点中均摊压力进行处理。
      - 常见扩容方案
        - 翻倍扩容法
        - 一致性Hash扩容
- [一次不寻常的慢查调优经历](https://mp.weixin.qq.com/s/s1QmqB7Xf3IrgRebd3RFow)
  - 索引失效
    - 问题出在参数eq_range_index_dive_limit，关于这个参数
    - The eq_range_index_dive_limit system variable enables you to configure the number of values at which the optimizer switches from one row estimation strategy to the other. To permit use of index dives for comparisons of up to N equality ranges, set eq_range_index_dive_limit to N + 1. To disable use of statistics and always use index dives regardless of N, set eq_range_index_dive_limit to 0.
    - Even under conditions when index dives would otherwise be used, they are skipped for queries that satisfy all these conditions:
      - A single-index FORCE INDEX index hint is present. The idea is that if index use is forced, there is nothing to be gained from the additional overhead of performing dives into the index.
      - The index is nonunique and not a FULLTEXT index.
      - No subquery is present.
      - No DISTINCT, GROUP BY, or ORDER BY clause is present.
  - eq_range_index_dive_limit 原本配置的就是200，我们直接设置成1来关闭index dive。
  - 统计信息分为持久化统计和动态统计，由参数innodb_stats_persistent控制
    - 持久化统计
      - 启用持久化统计信息，修改超过10%数据就要更新
      - 动态自动统计，修改1/16数据就要更新
      - innodb_stats_method控制统计信息针对索引中NULL值的算法当设置为nulls_equal所有的NULL值都视为一个value group；当设置为nulls_unequal每一个NULL值被视为一个value group；设置为nulls_ignore时，NULL值被忽略
      - 执行show table status、show index，访问I_S.TABLES/STATISTICS视图时更新统计信息
    - 动态统计
      - innodb_stats_persistent=0
        - 统计信息不持久化，每次动态采集，存储在内存中，重启失效（需重新统计），不推荐
      - innodb_stats_transient_sample_pages
        - 动态采集page，默认8个
      - 每个表设定统计模式
        - CREATE/ALTER TABLE … STATS_PERSISTENT=1,STATS_AOTU_RECALC=1,STATS_SAMPLE_PAGES=200;
      - mysql -auto-rehash
- [SQL优化系列之 in与range 查询](https://mp.weixin.qq.com/s/LmBH5Acl-GxtRmEMuLITaw)
  - 用in这种方式可以有效的替代一定的range查询，提升查询效率，因为在一条索引里面，range字段后面的部分是不生效的（ps.需要考虑 ICP）。MySQL优化器将in这种方式转化成 n*m 种组合进行查询，最终将返回值合并，有点类似union但是更高效。
  - 这里的一定数在MySQL5.6.5以及以后的版本中是由eq_range_index_dive_limit这个参数控制 。默认设置是10，一直到5.7以后的版本默认修改为200
    - eq_range_index_dive_limit = 0 只能使用index dive
    - 0 < eq_range_index_dive_limit <= N 使用index statistics
    - eq_range_index_dive_limit > N 只能使用index dive
  - 估计方法有2种:
    - dive到index中即利用索引完成元组数的估算,简称index dive;
    - index statistics:使用索引的统计数值,进行估算;
    - 对比这两种方式
      - index dive: 速度慢,但能得到精确的值（MySQL的实现是数索引对应的索引项个数，所以精确）
      - index statistics: 速度快,但得到的值未必精确
  - range查询与索引使用
    ```sql
     SELECT * FROM pre_forum_post WHERE tid=7932552 AND invisible IN('0','-2') ORDER BY dateline DESC LIMIT 10;
    ```
    - 优化器认为这是一个range查询，那么(tid,invisible,dateline)这条索引中，dateline字段肯定用不上了，也就是说这个SQL最后的排序肯定会生成一个临时结果集，然后再结果集里面完成排序，而不是直接在索引中直接完成排序动作
  - 如何使用optimize_trace
    ```sql
    set optimizer_trace='enabled=on';
    select * from information_schema.optimizer_trace
    ```
  - 如何使用profile
    ```sql
    set profiling=ON;
    执行sql;
    show profiles;
    show profile for query 2;
    show profile block io,cpu for query 2;
    ```
- [PG index](https://blog.crunchydata.com/blog/postgres-indexes-for-newbies)
  - Indexes are their own data structures and they’re part of the Postgres data definition language (the DDL). They're stored on disk along with data tables and other objects.
    - B-tree indexes are the most common type of index and would be the default if you create an index and don’t specify the type. B-tree indexes are great for general purpose indexing on information you frequently query.
    - BRIN indexes are block range indexes, specially targeted at very large datasets where the data you’re searching is in blocks, like timestamps and date ranges. They are known to be very performant and space efficient.
    - GIST indexes build a search tree inside your database and are most often used for spatial databases and full-text search use cases.
    - GIN indexes are useful when you have multiple values in a single column which is very common when you’re storing array or json data.
- Redis 为什么变慢了
  - 使用复杂度过高的命令
    - 分析
      - slowlog 命令
      - 使用聚合命令 - sort sunion
      - O(N)命令，但是N很大
      - 命令排队
    - 规避
      - 聚合操作在客户端
      - O(N)命令，N尽量小 （N <= 300）
  - 操作bigkey
    - bigkey申请、释放内存，耗时比较久
    - 规避
      - 避免bigkey （10KB以下）
      - UNLINK代替 DEL
  - 集中过期
    - 现象
      - 整点变慢，时间间隔固定， slowlog没有记录， expire keys突增
    - 过期策略
      - 被动 （惰性）
      - 主动 （定期清理，主线程）
    - 规避
      - 过期时间打散
      - lazyfree-lazy-expire=yes (后台进程)
  - 内存达到maxmemory
    - 现象
      - 满容之后写请求变慢
      - 写OPS越大越明显
      - 淘汰bigkey耗时久
    - 规避
      - no bigkey
      - 选择合适的淘汰策略
      - 拆分实例，分摊压力
      - lazyfree-lazy-eviction = yes
  - rehash - 翻倍扩容
    - 现象
      - 写入新key，偶发性延迟
      - rehash + maxmemory 触发大量key淘汰
    - 规避
      - key的数量在1亿以下
      - 升级 6.0 - 即将超过maxmemory，不做rehash
  - 持久化
    - RDB AOF - fork子进程
    - 规避
      - 单个实例在10G以下
      - slave节点备份
      - 关闭AOF AOF rewrite - 纯缓存case
      - 不要部署虚拟机
      - 避免全量同步：调大 repl-backlog-size
  - 内存大页
    - 现象
      - RDB AOF rewrite期间写请求变慢
    - 分析
      - 默认内存页4KB
      - 内存大页2MB
      - COW： fork的时候调用
    - 关闭内存大页
      - `echo never > /sys/kernel/mm/transparent_hugepage/enabled`
  - AOF
    - 现象
      - AOF everysec
      - 主线程阻塞
      - 主线程 写入到 page cache，当磁盘负载高的时候，导致AOF子线程fsync卡住
    - 规避
      - `no-appendfsync-on-rewrite = yes` AOF rewrite 期间，appendfsync = no
  - 绑定CPU
    - 现象
      - Redis进程绑定固定一个CPU核心
      - RDB AOF rewrite期间慢
    - Redis server
      - 主线程 - 处理请求
      - 后台线程 - 异步释放fd，异步AOF刷盘，lazyfree
      - 子进程 - RDB AOF rewrite
    - 分析
      - 子进程集成父进程CPU偏好，竞争关系
    - 缓解
      - 绑定多个CPU核心
      - 同一个物理核心
    - 规避
      - 不同进程，不同CPU
      - ```shell
        server_cpulist 
        bio_cpulist
        aof_cpulist
        bgsave_publist
        ```
      - 绑定CPU需谨慎
  - 使用SWAP
    - 现象
      - 所有请求变慢
      - 响应延迟- 几百毫秒，秒级
    - 分析
      - 内存数据放到磁盘
    - 规避
      - 足够内存，避免swap
       ```shell
       cat /proc/$pid/smaps | egrep '^(Swap|Size)'
       ```
      - 监控
  - 内存碎片
    - 现象
      - 开启内存碎片整理
      - 请求变慢 - 碎片整理在主线程
    - 规避
      - 合理调整阈值 `activefrag`
  - 网络负载高
    - 现象
      - 丢包，重传
    - 规避
      - 扩容，迁移
  - 监控
    - 配置有问题，脚本有bug：connection 数量
  
- [Multi Part AOF](https://mp.weixin.qq.com/s/v9yvJo7mKb5Hffw8Dw7gDQ)
  - AOF
    - 由于AOF会以追加的方式记录每一条redis的写命令，因此随着Redis处理的写命令增多，AOF文件也会变得越来越大，命令回放的时间也会增多，为了解决这个问题，Redis引入了AOF rewrite机制
  - AOFRW
    - 当AOFRW被触发执行时，Redis首先会fork一个子进程进行后台重写操作，该操作会将执行fork那一刻Redis的数据快照全部重写到一个名为temp-rewriteaof-bg-pid.aof的临时AOF文件中。 
  - AOFRW存在的问题
    - Memory
      - 在AOFRW期间，主进程会将fork之后的数据变化写进aof_rewrite_buf中，aof_rewrite_buf和aof_buf中的内容绝大部分都是重复的，因此这将带来额外的内存冗余开销。
    - CPU
      - 在AOFRW期间，主进程需要花费CPU时间向aof_rewrite_buf写数据，并使用eventloop事件循环向子进程发送aof_rewrite_buf中的数据：
      - 在子进程执行重写操作的后期，会循环读取pipe中主进程发送来的增量数据，然后追加写入到临时AOF文件：
      - 在子进程完成重写操作后，主进程会在backgroundRewriteDoneHandler 中进行收尾工作。其中一个任务就是将在重写期间aof_rewrite_buf中没有消费完成的数据写入临时AOF文件。如果aof_rewrite_buf中遗留的数据很多，这里也将消耗CPU时间。
    - Disk IO
      - 在AOFRW期间，主进程除了会将执行过的写命令写到aof_buf之外，还会写一份到aof_rewrite_buf中。aof_buf中的数据最终会被写入到当前使用的旧AOF文件中，产生磁盘IO。同时，aof_rewrite_buf中的数据也会被写入重写生成的新AOF文件中，产生磁盘IO。因此，同一份数据会产生两次磁盘IO。
  - MP-AOF实现
    - 将AOF分为三种类型，分别为：
      - BASE：表示基础AOF，它一般由子进程通过重写产生，该文件最多只有一个。
      - INCR：表示增量AOF，它一般会在AOFRW开始执行时被创建，该文件可能存在多个。
      - HISTORY：表示历史AOF，它由BASE和INCR AOF变化而来，每次AOFRW成功完成时，本次AOFRW之前对应的BASE和INCR AOF都将变为HISTORY，HISTORY类型的AOF会被Redis自动删除。
- [mysql主库更新后，从库都读到最新值了，主库还有可能读到旧值吗](https://mp.weixin.qq.com/s/EaTI063DJSH3gDNQhi-OZg)
  - 主从同步
    ![img.png](db_mysql_relay_log.png)
  - 主库更新后，主库都读到最新值了，从库还有可能读到旧值吗？
    - 如果此时主从延迟过大，这时候读从库，同步可能还没完成，因此读到的就是旧值
  - 主库更新后，从库都读到最新值了，主库还有可能读到旧值吗？
    - 假设当前的数据库事务隔离级别是可重复读
    - ![img.png](db_mysql_isolation.png)













