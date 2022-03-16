
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
  - [Serialized Roaring Bitmaps](https://github.com/dgraph-io/sroar)
    - Sroar operates on 64-bit integers and uses a single byte slice to store all the keys and containers. This byte slice can then be stored on disk, operated on directly in memory, or transmitted over the wire. There’s no encoding/decoding step required. For all practical purposes, sroar can be treated just like a byte slice.
- [When Bloom filters don't bloom](https://blog.cloudflare.com/when-bloom-filters-dont-bloom/)
- [布谷鸟过滤器](https://juejin.cn/post/6844903861749055502)
  - 布隆过滤器有以下不足：查询性能弱(布隆过滤器存储空间和插入/查询时间都是O(k))、空间利用效率低、不支持反向操作（删除）以及不支持计数
    - 查询性能弱是因为布隆过滤器需要使用多个 hash 函数探测位图中多个不同的位点，这些位点在内存上跨度很大，会导致 CPU 缓存行命中率低。
    - 空间效率低是因为在相同的误判率下，布谷鸟过滤器的空间利用率要明显高于布隆，空间上大概能节省 40% 多。不过布隆过滤器并没有要求位图的长度必须是 2 的指数，而布谷鸟过滤器必须有这个要求。从这一点出发，似乎布隆过滤器的空间伸缩性更强一些。
    - 不支持反向删除操作这个问题着实是击中了布隆过滤器的软肋。在一个动态的系统里面元素总是不断的来也是不断的走。布隆过滤器就好比是印迹，来过来就会有痕迹，就算走了也无法清理干净。比如你的系统里本来只留下 1kw 个元素，但是整体上来过了上亿的流水元素，布隆过滤器很无奈，它会将这些流失的元素的印迹也会永远存放在那里。随着时间的流失，这个过滤器会越来越拥挤，直到有一天你发现它的误判率太高了，不得不进行重建。
    - 布谷鸟过滤器在论文里声称自己解决了这个问题，它可以有效支持反向删除操作。而且将它作为一个重要的卖点，诱惑你们放弃布隆过滤器改用布谷鸟过滤器。
  - 布谷鸟哈希
    - 最简单的布谷鸟哈希结构是一维数组结构，会有两个 hash 算法将新来的元素映射到数组的两个位置。如果两个位置中有一个位置为空，那么就可以将元素直接放进去。但是如果这两个位置都满了，它就不得不「鸠占鹊巢」，随机踢走一个，然后自己霸占了这个位置。不同于布谷鸟的是，布谷鸟哈希算法会帮这些受害者（被挤走的蛋）寻找其它的窝。因为每一个元素都可以放在两个位置，只要任意一个有空位置，就可以塞进去。所以这个伤心的被挤走的蛋会看看自己的另一个位置有没有空，如果空了，自己挪过去也就皆大欢喜了。但是如果这个位置也被别人占了呢？好，那么它会再来一次「鸠占鹊巢」，将受害者的角色转嫁给别人。然后这个新的受害者还会重复这个过程直到所有的蛋都找到了自己的巢为止。
  - 优化
    - 改良的方案之一是增加 hash 函数，让每个元素不止有两个巢，而是三个巢、四个巢。这样可以大大降低碰撞的概率，将空间利用率提高到 95%左右。
    - 另一个改良方案是在数组的每个位置上挂上多个座位，这样即使两个元素被 hash 在了同一个位置，也不必立即「鸠占鹊巢」，因为这里有多个座位，你可以随意坐一个。除非这多个座位都被占了，才需要进行挤兑。
  - 布谷鸟过滤器
    - 首先布谷鸟过滤器还是只会选用两个 hash 函数，但是每个位置可以放置多个座位。这两个 hash 函数选择的比较特殊，因为过滤器中只能存储指纹信息。当这个位置上的指纹被挤兑之后，它需要计算出另一个对偶位置。而计算这个对偶位置是需要元素本身的，我们来回忆一下前面的哈希位置计算公式。
    ```
    fp = fingerprint(x)
    p1 = hash1(x) % l
    p2 = hash2(x) % l
    ```
    - 特殊的 hash 函数
      - 布谷鸟过滤器巧妙的地方就在于设计了一个独特的 hash 函数，使得可以根据 p1 和 元素指纹 直接计算出 p2，而不需要完整的 x 元素。
        ```
        fp = fingerprint(x)
        p1 = hash(x)
        p2 = p1 ^ hash(fp)  // 异或
        ```
      - 从上面的公式中可以看出，当我们知道 fp 和 p1，就可以直接算出 p2。同样如果我们知道 p2 和 fp，也可以直接算出 p1 —— 对偶性。
        `p1 = p2 ^ hash(fp)`
      - 布谷鸟过滤器强制数组的长度必须是 2 的指数，所以对数组的长度取模等价于取 hash 值的最后 n 位。在进行异或运算时，忽略掉低 n 位 之外的其它位就行。将计算出来的位置 p 保留低 n 位就是最终的对偶位置。
    - 一个明显的弱点
      - 如果布谷鸟过滤器对同一个元素进行多次连续的插入会怎样？
      - 根据上面的逻辑，毫无疑问，这个元素的指纹会霸占两个位置上的所有座位 —— 8个座位。这 8 个座位上的值都是一样的，都是这个元素的指纹。如果继续插入，则会立即出现挤兑循环。从 p1 槽挤向 p2 槽，又从 p2 槽挤向 p1 槽。
      - 如果想要让布谷鸟过滤器支持删除操作，那么就必须不能允许插入操作多次插入同一个元素，确保每一个元素不会被插入多次（kb+1）。这里的 k 是指 hash 函数的个数 2，b 是指单个位置上的座位数，这里我们是 4
    - 优点：
      - 查询性能较高；
      - 空间利用率较高；
      - 保证了一个比特只被一个元素映射，所以允许删除操作；
    - 缺点：
      - 不能完美的支持删除，存在误删的情况；
      - 存储空间的大小必须为2的指数的限制让空间效率打了折扣；
- [Queue](https://github.com/gammazero/deque)
  - Most queue implementations are in one of three flavors: slice-based, linked list-based, and circular-buffer (ring-buffer) based.
    - Slice-based queues tend to waste memory because they do not reuse the memory previously occupied by removed items. Also, slice based queues tend to only be single-ended.
    - Linked list queues can be better about memory reuse, but are generally a little slower and use more memory overall because of the overhead of maintaining links. They can offer the ability to add and remove items from the middle of the queue without moving memory around, but if you are doing much of that a list is the wrong data structure.
    - Ring-buffer queues offer all the efficiency of slices, with the advantage of not wasting memory. Fewer allocations means better performance. They are just as efficient adding and removing items from either end so you naturally get a double-ended queue. So, as a general recommendation I would recommend a ring-buffer based queue implementation. This is what is discussed in the rest of this post.




