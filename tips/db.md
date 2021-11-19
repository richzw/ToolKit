
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









