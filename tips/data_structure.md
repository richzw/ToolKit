
- [map 缩容？](https://mp.weixin.qq.com/s/Slvgl3KZax2jsy2xGDdFKw)

  在 Go 底层源码 src/runtime/map.go 中，扩缩容的处理方法是 grow 为前缀的方法来处理的. 无论是扩容还是缩容，其都是由 hashGrow 方法进行处理
  若是扩容，则 bigger 为 1，也就是 B+1。代表 hash 表容量扩大 1 倍。不满足就是缩容，也就是 hash 表容量不变。

  可以得出结论：map 的扩缩容的主要区别在于 hmap.B 的容量大小改变。_而缩容由于 hmap.B 压根没变，内存空间的占用也是没有变化的_。

  若要实现 ”真缩容“，唯一可用的解决方法是：**创建一个新的 map 并从旧的 map 中复制元素**。

  - [为什么不支持？](https://github.com/golang/go/issues/20135)
  - 简单来讲，就是没有找到一个很好的方法实现，存在明确的实现成本问题，没法很方便的 告诉 Go 运行时，我要：
    - 记得保留存储空间，我要立即重用 map。
    - 赶紧释放存储空间，map 从现在开始会小很多。
- [LSM Tree](https://mp.weixin.qq.com/s/sK2qqYCM-dZJTbbFn-23zg)
  - Log Structured Merge Tree ，是一种在机械盘时代大放异彩的存储架构设计。LSM Tree 是一个把顺序写发挥到极致的设计架构
    - 一个 key/value 存储引擎
    - 用户递交数据流程分为两步：写 log 文件，修改内存
    - 删除也是写入。往存储里面写一条带删除标记的记录，而不是直接更新原来的数据。
    - 从用户 key/value 来讲是 log 的结构是一种无序的结构，它的查找效率非常低。所以，自然而然，LSM 的架构里就需要引入一种新型的**有序的数据结构，这个就是 sst** 文件（ 全名：sorted string  table ）
    - 把有效的数据从 sst 文件中读出来（删除或者被覆盖的旧数据丢弃）写到新的文件，然后修改指向关系，然后把旧的文件删掉。这个过程叫做 compact 
  - LSM Tree 的设计思想，考虑极致的提升写的性能，读的性能则靠其他的手段解决
    ![img.png](datastructure_lsm.png)
    - leveldb 的 compact 分为两种：
      - minor compact ：这个是 memtable 到 Level 0 的转变；
      - major compact ：这个是 Level 0 到底层或者底下的 Level 的 sstable 的相互转变
  - 为什么越来越多“唱衰” LSM 的声音呢
    - SSD 多通道并发、超高的随机性能是变革的核心因素。
    - 《WiscKey: Separating Keys from Values in SSD-Conscious Storage》就讨论了在 SSD 时代，LSM 架构的优化思路
- [BitMap Index](https://github.com/mkevac/gopherconrussia2019)
  - [Details](https://medium.com/bumble-tech/bitmap-indexes-in-go-unbelievable-search-speed-bb4a6b00851)
  - Indexing approach
    - Hierarchical division - *-trees
    - Hash Mapping - hash maps, reverse indexes
    - Instantly - Bloom filter, Cuckoo filter
    - Bitmap index
  - Bitmap Index
    - bitmap/bitset AND OR
    - high-performance for low cardinality
  - Bitmap Index Problems
    - High-cardinality problem
      - Solution 1: Roaring bitmap - bitmaps, arrays, bit runs
      - Solution 2: Binning
    - High-throughput problem
      - it can be expensive to update bitmaps.
      - Solution 1: Sharding
      - Solution 2: Versioned Indexes
    - Non-trivial queries
      - range query
      - geo query
- [Introducing Serialized Roaring Bitmaps in Golang](https://dgraph.io/blog/post/serialized-roaring-bitmaps-golang/)
  - [Roaring Bitmap](https://github.com/RoaringBitmap/roaring)
  - 







