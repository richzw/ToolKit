
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



