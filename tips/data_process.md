
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
- Iceberg 的真正优势
  - 提供了多个变革性的能力，如模式演进（Schema evolution）、时间旅行（Time travel）、以及使用各种工具进行数据分析（兼容多种引擎）
  - Iceberg 可以将 S3 buckets 转变为结构化、可查询的数据集，加上适当的访问控制，兼容任何现代查询引擎
  - Iceberg 广泛的兼容性则可以摆脱厂商锁定。
  - Iceberg 支持多引擎，用户可以根据任务类型选择最合适的工具。
    - 例如，将 Iceberg 与 Snowflake 配对以处理复杂的分析查询（OLAP），与 DuckDB 配对进行轻量级分析。这类组合既节省成本，又不影响灵活性。
- [表格格式” vs. “文件（存储）格式]
  - Parquet 等文件格式与 Iceberg 等表格格式之间的主要区别在于它们的用途。
    - 文件格式专注于高效存储和压缩数据。它们定义了如何在磁盘或分布式文件系统（如 Amazon S3）中组织和编码表示记录和列的原始字节。
    - 表格格式在存储的数据之上提供了逻辑抽象，以方便组织、查询和更新。它们使 SQL 引擎能够将文件集合视为具有行和列的表格，可以以事务方式查询和更新这些行和列。
  - 文件（存储）格式（File Format）
    - 列式存储
      • Parquet：高压缩率，支持复杂嵌套结构，适合 OLAP 场景，应用于大数据分析（如 Spark、Hive）和数据湖存储。
      • ORC (Optimized Row Columnar)：优化行列混合存储，支持索引和谓词下推，常用于 Hive 数据仓库和批量 ETL 处理。
    - 行式存储
      • Avro：基于 Schema 的行式存储，支持动态模式演化，适合流式数据传输（如 Kafka 消息序列化），应用于跨语言数据交换和实时数据管道。
      • CSV/TSV：纯文本格式，人类可读，兼容性强，但无压缩和模式信息，适用于数据导入导出和小型数据集交换。
      • JSON：半结构化，支持嵌套数据，但解析效率低，常用于 Web API 响应和日志存储（需后续转换为高效格式如 Parquet）。
    - 混合存储
      • Arrow：内存列式格式，支持零拷贝读取，用于高速内存计算（如 Pandas、Spark 内存计算），不用于持久化存储。
  - 表格格式（Table Format）
    - 数据湖表格格式
      • Apache Iceberg：支持 ACID 事务、隐藏分区、时间旅行（数据版本控制），引擎无关（如 Flink、Spark、Trino），应用于实时数据湖和多引擎协作。
      • Delta Lake：基于 Spark 生态，提供 ACID 事务和 Upsert 操作，深度集成 Spark，适用于湖仓一体和频繁更新的场景（底层默认使用 Parquet）。
      • Apache Hudi：专注于增量更新（CDC），支持高效的 Upsert 和增量拉取，应用于实时数据管道和 CDC 场景。
    - 传统表格格式
      • Hive 表：基于目录分区，元数据存储在 Hive Metastore，支持分区、分桶等管理，但功能有限（缺乏 ACID 事务），适用于离线批处理（如 Hive/Spark SQL）。
  - 实时数仓：文件格式为 Parquet，表格格式为 Iceberg，流程为 Kafka → Flink 实时处理 → 写入 Iceberg（Parquet 文件）→ Trino 查询。
  - 频繁更新的用户数据：文件格式为 Parquet，表格格式为 Delta Lake，流程为 Spark 读取用户表 → Merge 操作更新 → 写入 Delta Lake
  - 日志分析：文件格式为 JSON（初始导入）→ 转换为 ORC/Parquet，表格格式为 Hive 表，流程为日志文件（JSON）→ Hive 表分区存储（列式格式）→ Hive SQL 分析。
- [ETL Tools for Unstructured Data](https://zilliz.com/blog/selecting-the-right-etl-tools-for-unstructured-data-to-prepare-for-ai)























