
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
      

    








