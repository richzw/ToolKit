
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
  - LSM Tree 的设计思想 - 保存一组合理组织、后台合并的 SSTables
    - ![img.png](datastructure_lsm.png)
    - SSTable（Sorted String Table）
      - 构建 SSTable 文件 
        - 在内存中维护一个有序结构（称为 MemTable）。红黑树、AVL 树、条表。
        - 到达一定阈值之后全量 dump 到外存。
      - 维护 SSTable 文件
        - 先去 MemTable 中查找，如果命中则返回。
        - 再去 SSTable 按时间顺序由新到旧逐一查找。
      - 如果 SSTable 文件越来越多，则查找代价会越来越大。因此需要将多个 SSTable 文件合并，以减少文件数量，同时进行 GC，我们称之为紧缩（ Compaction）
      - 性能优化
        - 优化 SSTable 的查找。常用 Bloom Filter。
        - 层级化组织 SSTable。以控制 Compaction 的顺序和时间。常见的有 size-tiered 和 leveled compaction。
    - leveldb 的 compact 分为两种：
      - minor compact ：这个是 memtable 到 Level 0 的转变；
      - major compact ：这个是 Level 0 到底层或者底下的 Level 的 sstable 的相互转变
  - 为什么越来越多“唱衰” LSM 的声音呢
    - SSD 多通道并发、超高的随机性能是变革的核心因素。
    - 《WiscKey: Separating Keys from Values in SSD-Conscious Storage》就讨论了在 SSD 时代，LSM 架构的优化思路
  - LSM vs B-Tree
    - B 树是数据可变的代表结构
      - B 树的难点在于平衡性维护和并发控制，一般用在读多写少的场景
      - 以页（page）为粒度对磁盘数据进行修改, 面向页、查找树
      - 维护了所有数据的有序性，读取性能必然起飞，但写入性能你也别抱太大希望。
      - 优化：
        - 不使用 WAL，而在写入时利用 Copy On Write 技术。同时，也方便了并发控制。如 LMDB、BoltDB。
        - 对中间节点的 Key 做压缩，保留足够的路由信息即可。以此，可以节省空间，增大分支因子。
        - 为了优化范围查询，有的 B 族树将叶子节点存储时物理连续。但当数据不断插入时，维护此有序性的代价非常大。
    - LSM 树是数据不可变的代表结构。你只能在尾部追加新数据，不能修改之前已经插入的数据。变随机写为顺序写
      - LSM 树的难点在于 compact 操作和读取数据时的效率优化，一般用在写多读少的场景。
      - 可以维护局部数据的有序性，从而一定程度提升读性能。
    - ![img.png](data_structure_lsm_vs_btree.png)
    
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
- [Why doesn't Dijkstra's algorithm work for negative weight edges?](https://stackoverflow.com/questions/13159337/why-doesnt-dijkstras-algorithm-work-for-negative-weight-edges)
  - The reason for this is that Dijkstra's algorithm are greedy algorithms that assume that once they've computed the distance to some node, the distance found must be the optimal distance. In other words, the algorithm doesn't allow itself to take the distance of a node it has expanded and change what that distance is. In the case of negative edges, your algorithm, and Dijkstra's algorithm, can be "surprised" by seeing a negative-cost edge that would indeed decrease the cost of the best path from the starting node to some other node.
  - Note that this is important, because in each relaxation step, the algorithm assumes the "cost" to the "closed" nodes is indeed minimal, and thus the node that will next be selected is also minimal.
  - The idea of it is: If we have a vertex in open such that its cost is minimal - by adding any positive number to any vertex - the minimality will never change.
  - Without the constraint on positive numbers - the above assumption is not true.
  - Since we do "know" each vertex which was "closed" is minimal - we can safely do the relaxation step - without "looking back". If we do need to "look back" - Bellman-Ford offers a recursive-like (DP) solution of doing so.
- [一致性 Hash 算法原理总结](https://mp.weixin.qq.com/s/b9wRO3q-9XW4yQYDtbJyvQ)
  - 算法详述
    - 一致性哈希解决了简单哈希算法在分布式哈希表（Distributed Hash Table，DHT）中存在的动态伸缩等问题
    - 一致性 hash 是对固定值 2^32 取模. 使用服务器 IP 地址进行 hash 计算，用哈希后的结果对2^32取模，结果一定是一个 0 到2^32-1之间的整数；
    - 一致性 Hash 就是：将原本单个点的 Hash 映射，转变为了在一个环上的某个片段上的映射
  - [数据偏斜&服务器性能平衡问题](https://github.com/JasonkayZK/consistent-hashing-demo)
    - 引入虚拟节点来解决负载不均衡的问题 
      - 即将每台物理服务器虚拟为一组虚拟服务器，将虚拟服务器放置到哈希环上，如果要确定对象的服务器，需先确定对象的虚拟服务器，再由虚拟服务器确定物理服务器；
    - 分配的虚拟节点个数越多，映射在 hash 环上才会越趋于均匀，节点太少的话很难看出效果；
    - 引入虚拟节点的同时也增加了新的问题，要做虚拟节点和真实节点间的映射，对象key->虚拟节点->实际节点之间的转换；
  - [含有负载边界值的一致性 Hash](https://ai.googleblog.com/2017/04/consistent-hashing-with-bounded-loads.html)
    - 如果很多的热点数据都落在了同一台缓存服务器上，则可能会出现性能瓶颈
    - Google 提出了含有[负载边界值的一致性 Hash 算法](https://arxiv.org/abs/1608.01350)，此算法主要应用于在实现一致性的同时，实现负载的平均性；
    - 这个算法将缓存服务器视为一个含有一定容量的桶（可以简单理解为 Hash 桶），将客户端视为球，则平均性目标表示为：所有约等于平均密度（球的数量除以桶的数量）
- [一致性哈希算法](https://writings.sh/post/consistent-hashing-algorithms-part-1-the-problem-and-the-concept)
  - 问题的提出
    ![img.png](ds_consistent_hash.png)
  - [哈希环法](https://writings.sh/post/consistent-hashing-algorithms-part-2-consistent-hash-ring)
    - 实现哈希环的方法一般叫做 ketama 或 hash ring。 核心的逻辑在于如何在环上找一个和目标值 z 相近的槽位， 我们把环拉开成一个自然数轴， 所有的槽位在环上的哈希值组成一个有序表。 在有序表里做查找， 这是二分查找可以解决的事情， 所以哈希环的映射函数的时间复杂度是
    - 带权重的一致性哈希环
      - 采用影子节点可以减少真实节点之间的负载差异。
      - 影子节点是一个绝妙的设计，不仅提高了映射结果的均匀性， 而且为实现加权映射提供了方式。 但是，影子节点增加了内存消耗和查找时间
    - 一致性哈希环下的热扩容和容灾 
      - 对于增删节点的情况，哈希环法做到了增量式的重新映射， 不再需要全量数据迁移的工作。 但仍然有部分数据出现了变更前后映射不一致， 技术运营上仍然存在如下问题：
        - 扩容：当增加节点时，新节点需要对齐下一节点的数据后才可以正常服务。
        - 缩容：当删除节点时，需要先把数据备份到下一节点才可以停服移除。
        - 故障：节点突然故障不得不移除时，面临数据丢失风险。
      - 如果我们要实现动态扩容和缩容，即所谓的热扩容，不停止服务对系统进行增删节点， 可以这样做：
        - 数据备份(双写)： 数据写入到某个节点时，同时写一个备份(replica)到顺时针的邻居节点。
        - 请求中继(代理)： 新节点刚加入后，数据没有同步完成时，对读取不到的数据，可以把请求中继(replay)到顺时针方向的邻居节点。
  - [跳跃一致性哈希法](https://writings.sh/post/consistent-hashing-algorithms-part-3-jump-consistent-hash)
    ```cgo
    int32_t JumpConsistentHash(uint64_t key, int32_t num_buckets) {
      int64_t b = -1, j = 0;
      while (j < num_buckets) {
        b = j;
        key = key * 2862933555777941757ULL + 1;
        j = (b + 1) * (double(1LL << 31) / double((key >> 33) + 1));
      }
      return b;
    }
    ```
    - 用随机数来决定一个 k 每次要不要跳到新槽位中去。 但是请注意，这里所说的「随机数」是指伪随机数，即只要种子不变，随机序列就不变。
    - 跳跃一致性哈希算法的设计非常精妙， 我认为最美的部分是利用了伪随机数的一致性和分布均匀性。
    - 跳跃一致性哈希在执行速度、内存消耗、映射均匀性上都比经典的哈希环法要好。
    - 跳跃一致性哈希算法有两个显著缺点：
      - 无法自定义槽位标号
        - 跳跃一致性哈希算法中， 因为我们没有存储任何数据结构， 所以我们无法自定义槽位标号， 标号是从 0 开始数过来的。
      - 只能在尾部增删节点
    - 跳跃一致性哈希下的热扩容和容灾 
      - 热扩容 -  可以采用和一致性哈希环法类似的办法， 即请求中继： 新加入的节点对于读取不到的数据，可以把请求中继(relay)到老节点，并把这个数据迁移过来。
      - 容灾 - 在执行数据写操作时，同时写一份数据到备份节点。 备份节点这样选定：
         - 尾部节点备份一份数据到老节点。
         - 非尾部节点备份一份数据到右侧邻居节点。
  - [Maglev一致性哈希法](https://writings.sh/post/consistent-hashing-algorithms-part-4-maglev-consistent-hash)
    - Maglev一致性哈希的思路是查表： 建立一个槽位的查找表(lookup table)， 对输入 k 做哈希再取余，即可映射到表中一个槽位。
    - 为每个槽位生成一个大小为 M 的序列 permutation 叫做「偏好序列」吧。 然后， 按照偏好序列中数字的顺序，每个槽位轮流填充查找表。 将偏好序列中的数字当做查找表中的目标位置，把槽位标号填充到目标位置。 如果填充的目标位置已经被占用，则顺延该序列的下一个填
    - 这是一种类似「二次哈希」的方法， 使用了两个独立无关的哈希函数来减少映射结果的碰撞次数，提高随机性。
    - 查找表的长度 M 必须是一个质数。 和「哈希表的槽位数量最好是质数」是一个道理， 这样可以减少哈希值的聚集和碰撞，让分布更均匀
    - Maglev一致性哈希的算法的内容， 简单来说：
      - 为每个槽位生成一个偏好序列， 尽量均匀随机。
      - 建表：每个槽位轮流用自己的偏好序列填充查找表。
      - 查表：哈希后取余数的方法做映射。
    - 难以实现后端节点的数据备份逻辑，因此工程上更适合弱状态后端的场景
  - ![img.png](data_structure_consistent_hash_summary.png)
- [Go 语言高性能哈希表的设计与实现](https://mp.weixin.qq.com/s/KB-VwshP7FlzO-OutuT-2w)
  - 碰撞处理
    - 链地址法（chaining）
      - 实现最简单直观
      - 空间浪费较少
    - 开放寻址法（open-addressing）
      - 每次插入或查找操作只有一次指针跳转，对CPU缓存更友好
      - 所有数据存放在一块连续内存中，内存碎片更少
      - 当max load factor较大时，性能不如链地址法。
      - 然而当我们主动牺牲内存，选择较小的max load factor时（例如0.5），形势就发生逆转，开放寻址法反而性能更好。因为这时哈希碰撞的概率大大减小，缓存友好的优势得以凸显。
      - 空闲桶探测方法
        - 线性探测（linear probing）：对i = 0, 1, 2...，依次探测第H(k, i) = H(k) + ci mod |T|个桶。
        - 平方探测（quadratic probing）：对i = 0, 1, 2...，依次探测H(k, i) = H(k) + c1i + c2i2 mod |T|。其中c2不能为0，否则退化成线性探测。
        - 双重哈希（double hashing）：使用两个不同哈希函数，依次探测H(k, i) = (H1(k) + i * H2(k)) mod |T|
  - Max load factor
    - 对链地址法哈希表，指平均每个桶所含元素个数上限。 
    - 对开放寻址法哈希表，指已填充的桶个数占总的桶个数的最大比值。 
    - max load factor越小，哈希碰撞的概率越小，同时浪费的空间也越多。
  - Growth factor
    - 指当已填充的桶达到max load factor限定的上限，哈希表需要rehash时，内存扩张的倍数。growth factor越大，哈希表rehash的次数越少，但是内存浪费越多。
  - 基本设计与参数选择 [Source](https://github.com/matrixorigin/matrixone/tree/main/pkg/container/hashtable)
    - 我们照搬了ClickHouse的如下设计：
      - 开放寻址
      - 线性探测
      - max load factor = 0.5，growth factor = 4
      - 整数哈希函数基于CRC32指令
    - 具体原因前面已经提到，当max load factor不大时，开放寻址法要优于链地制法，同时线性探测法又优于其他的探测方法。
    - 并做了如下修改（优化）：
      - 字符串哈希函数基于AESENC指令
      - 插入、查找、扩张时批量计算哈希函数
      - 扩张时直接遍历旧表插入新表





