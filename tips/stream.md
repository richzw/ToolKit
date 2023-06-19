
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



