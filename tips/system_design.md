
- [估算的一些方法](https://mp.weixin.qq.com/s/fH-AJpE99ulSLbC_1jxlqw)
  - 每个程序员都要了解的延迟数字
  ![img.png](sd_delay_number.png)
  - 一些数字积累
    - 某支付服务的支付峰值 60w QPS
    - Go GC 打开写屏障需要花费 10-30us
    - 内网中，一次网络请求的耗时是 ms 级别
    - 万兆网卡，1.25GB/s 打满
    - 4C8G 建 10w 到 20w 的连接没有问题
    - 因为机械硬盘的机械结构，随机 I/O 与顺序的 I/O 性能可能相差几百倍。固态硬盘则只有十几倍到几十倍之间
    - twitter 工程师认为，良好体验的网站平均响应时间应该在 500ms 左右，理想的时间是 200-300ms
    - 平均 QPS：日平均用户请求除以 4w。日平均用户请求，一般来自产品的评估。峰值 QPS：平均 QPS 的 2~4 倍
  - 本章最后有一个实战的例子：评估 twitter 的 QPS 和存储容量。
    
    先给出了一些预设：
      - 300 个 million 的月活跃用户
      - 50% 的用户每天都使用 twitter
      - 用户平均每天发表 2 条 tweets
      - 10% 的 tweets 包含多媒体
      - 多媒体数据保存 5 年
    
    下面是估算的过程：
      - 先预估 QPS：
        - DAU（每天的活跃用户数，Daily Active Users）为：300 million（总用户数） * 50% = 150 million
        - 发 tweets 的平均 QPS：150 million * 2 / 24 hour / 3600 second = ~3500
        - 高峰期 QPS 一般认为是平均 QPS 的 2 倍：2 * 3500 = 7000 QPS
      - 再来估算存储容量：
        - 假设多媒体的平均大小为 1MB，那么每天的存储容量为：150 million * 2 * 10% * 1MB = 30 TB。5 年的存储容量为 30 TB * 365 * 5 = 55 PB。
      - 最后这两个的估算过程是这样的：
        - 300 个 million * 10%* 1MB，1 MB 其实就是 6 个 0，相当于 million 要进化 2 次：million -> billion -> trillion，即从 M -> G -> T，于是结果等于 300 T * 10% = 30 T。
        - 30 TB * 365 * 5 = 30 TB * 1825 = 30 TB * 10^3 * 1.825，TB 进化一次变成 PB，于是等于 30 * 1.825 PB = 55 PB。

    



