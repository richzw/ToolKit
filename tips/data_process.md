
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
- 


