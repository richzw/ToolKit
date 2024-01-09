
- [优化 Milvus 性能](https://mp.weixin.qq.com/s/4gDsAF4QnmXWzomrSFRLLg)
    - Milvus 是读写分离且无状态的向量数据库，状态信息储存在 etcd 中，coordinator 节点去 etcd 请求状态并修改状态
        - 当用户需要查看状态信息、清理状态信息场景时，etcd 调试工具必不可少。
        - [BirdWatcher  是 Milvus 2.0 项目的调试工具，该工具连接 etcd 并检查 Milvus 系统的某些状态](https://mp.weixin.qq.com/s/ot-eMCKqM7aP5pEbGaMIQA)
    - Milvus 单机
        - 在单机模式下，milvus内置一个rocksdb用于代替pulsar的功能，rdb_data目录里的东西是rockdb管理的，所有insert/delete/upsert的数据都会先在rocksdb里存一份做为write ahead log，然后querynode datanode从rocksdb里把数据拉出来消费
        - 如果rocksdb里的数据消费完了，不会立刻删除，因为rocksdb有自己的gc，只不过这些数据对milvus来说已经消费过了，放着只是为了保证数据安全性，一旦milvus崩了，再启动的时候，那些没被持久化到minio里的数据还能从rocksdb里拉回来
    - 合理的预计数据量，表数目大小，QPS 参数等指标
    - 选择合适的索引类型和参数
        - 索引的选择对于向量召回的性能至关重要，Milvus 支持了 Annoy，Faiss，HNSW，DiskANN 等多种不同的索引，用户可以根据对延迟、内存使用和召回率的需求进行选择
        - 是否需要精确结果？
            - 只有 Faiss 的 Flat 索引支持精确结果，但需要注意 Flat 索引检索速度很慢，查询性能通常比其他 Milvus 支持的索引类型低两个数量级以上，因此只适合千万级数据量的小查询
        - 数据量是否能加载进内存？
            - 对于大数据量，内存不足的场景，Milvus 提供两种解决方案：
                - DiskANN
                    - DiskANN 依赖高性能的磁盘索引，借助 NVMe 磁盘缓存全量数据，在内存中只存储了量化后的数据。
                    - DiskANN 适用于对于查询 Recall 要求较高，QPS 不高的场景。
        - 构建索引和内存资源是否充足
            - 性能优先，选择 HNSW 索引
- [Milvus 2.0 数据插入与持久化](https://mp.weixin.qq.com/s/D0xdD9mqDgxFvNY19hvDgQ)
    - 删数据逻辑
        - 调用delete删除一条已存在的数据时，它只是在某个segment里把某条数据标记为deleted，但是这个segment的数据此时仍然是一个整体，那条被删的数在内存里仍然占着空间。
        - 当你又调用insert增加一条数据时，这条新数据实际上是放入一个新的segment中，这条新数据也会占用额外内存空间。因此，随着你继续删除+insert，你会看到内存用量增加，新的这个segment执行的是暴搜，cpu用量会增加。
        - 随着新的segment中的数据达到一定量可以建索引了，indexnode就开始给这个新segment建索引，建索引就会消耗cpu。只有当某个segment中被删的数据达到20%以上，datanode开始对这个segment进行compact，
        - 把deleted的数据去除掉，剩下的数据存为一个新的segmemt。在compact之后有可能会发生小segment合并成大segment。总之，删除和更新数据会产生很多额外的工作，消耗内存消耗cpu，设计上就如此
    - 如果用num_enrtities观察行数的话，是看不出变化的，因为num_entities不统计被删除的行数。如果你是删除之后再用query去查询主键，还能查到的话，那八成是因为你是删完就立即query，而consistency_level没有设为Strong
- [动态 Schema](https://mp.weixin.qq.com/s/jhyePhxjUbWBicEvqxIKGQ)
  - Milvus 如何实现动态 Schema 功能
    - Milvus 通过用隐藏的元数据列的方式，来支持用户为每行数据添加不同名称和数据类型的动态字段的功能。
    - 当用户创建表并开启动态字段时，Milvus 会在表的 Schema 里创建一个名为$meta的隐藏列。JSON 是一种不依赖语言的数据格式，被现代编程语言广泛支持，因此 Milvus 隐藏的动态实际列使用 JSON 作为数据类型。
  - 动态 Schema 的 A、B 面
    - 一方面，动态 Schema 设置简便，无需复杂的配置即可开启动态 Schema；动态 Schema 可以随时适应数据模型的变化，开发者无需进行重构或调整代码。
    - 另一方面，使用动态 Schema 进行过滤搜索比固定 Schema 慢得多；在动态 Schema 上进行批量插入比较复杂，推荐用户使用行式插入接口写入动态字段数据。













