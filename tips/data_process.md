
- [Pandas的各种操作](https://mp.weixin.qq.com/s/Rkz0fbI_Qw0dR4q_yvjszQ)
  - sort_values
   ```shell
   (dogs[dogs['size'] == 'medium']
    .sort_values('type')
    .groupby('type').median()
   )
   ```
  - groupby + multi aggregation
   ```shell
   (dogs
     .sort_values('size')
     .groupby('size')['height']
     .agg(['sum', 'mean', 'std'])
   )
   ```
  - filtering for columns `df.loc[:, df.loc['two'] <= 20]`
  - filtering for rows `dogs.loc[(dogs['size'] == 'medium') & (dogs['longevity'] > 12), 'breed']`
  - pivot table `dogs.pivot_table(index='size', columns='kids', values='price')`
  - stacking column index
  - unstacking row index
  - resetting index
  - [Source](https://pandastutor.com/index.html)
- [ClickHouse JOIN优化](https://mp.weixin.qq.com/s/SN1bbddO_qYmAWLSz3IhsA)
- [图解 Pandas](https://mp.weixin.qq.com/s/cSk9gCdUTlCV8csmbkj3KQ)
- [改进字典的大数据多维分析加速](https://mp.weixin.qq.com/s/XSrRc5ccHFJBE-IzORm-3Q)
  - 为了解决RoaringBitmap因数据连续性低而导致存储计算效率低的问题，我们围绕ClickHouse建设了一套稠密字典编码服务体系。
    - 正向查询：用原始key值查询对应的字典值value。
    - 反向查询：用字典值value查询对应的原始key。
- [ClickHouse高并发写入优化](https://mp.weixin.qq.com/s/3Q-Gu_CnU3ynL7hjujkCow)
- [StarRocks存算分离](https://mp.weixin.qq.com/s/9fvVtInwiR93GGVR8yarLA)
  - Clickhouse在大数据量下的困境
    - 物化视图缺乏透明改写能力
    - 缺乏离线导入功能
    - 扩容困难
  - 基于StarRocks降本增效
    - 存算分离带来成本下降
    - 在复杂SQL join能力上大幅领先Clickhouse
