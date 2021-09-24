
- order by
  - 常规排序
    - explain结果  using index condition; using filesort
    - ![img.png](mysql_orderby_sort_buffer.png)
    - 上述流程只对原表的数据读了一遍，剩下的操作都是在 sort_buffer 和临时文件中执行的
  - rowid排序
    - max_length_for_sort_data，是 MySQL 中专门控制用于排序的行数据的长度的一个参数。它的意思是，如果单行的长度超过这个值，MySQL 就认为单行太大，要换一个算法
    - ![img.png](mysql_orderby_recalltable.png)
    - rowid排序多了一步回表操作
  - order by都需要排序？
    - explain结果 using index condition
    - ![img.png](mysql_orderby.png)
    - 这个查询过程不需要临时表，也不需要排序
  - 覆盖索引
    - explain结果 using index
    - Extra 字段里面多了“Using index”，表示的就是使用了覆盖索引，性能上会快很多


