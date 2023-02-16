
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





