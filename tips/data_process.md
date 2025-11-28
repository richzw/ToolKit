
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
  - [Advancing the Lakehouse with Apache Iceberg v3 on Databricks](https://www.databricks.com/blog/advancing-lakehouse-apache-iceberg-v3-databricks)
    - With Iceberg v3, deletion vectors, row-level lineage, and the Variant data type are now available on all managed tables
      -  V2： 支持实时，行级删除 + CDC + 流式更新
      - V3： 性能飞跃，删除向量 + 新数据类型 + 格式统一
    - Deletion Vectors：更高效的更新/删除
      - 行级删除/更新通过“逻辑标记 + 延迟合并”实现。
      - 物理文件更稳定，减少重写与小文件问题
    - Row Lineage：行级并发控制
      - 行级唯一 ID（row-id）是并发控制、审计、回溯、CDC 的基础。
      - 写入流程中，乐观并发控制可以利用 row-id 做更精细的冲突判定
    - Variant 数据类型：面向半结构化数据的灵活建模
      - Variant = 半结构化数据的统一存储抽象 + 查询时结构化（shredding）。
      - 兼容多格式导入（JSON/CSV/XML），对真实业务杂乱数据友好
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
- 𝐃𝐚𝐭𝐚 𝐖𝐚𝐫𝐞𝐡𝐨𝐮𝐬𝐞, 𝐃𝐚𝐭𝐚 𝐋𝐚𝐤𝐞, 𝐃𝐚𝐭𝐚 𝐋𝐚𝐤𝐞𝐡𝐨𝐮𝐬𝐞, 𝐃𝐚𝐭𝐚 𝐌𝐞𝐬𝐡.
  - 𝐃𝐚𝐭𝐚 𝐖𝐚𝐫𝐞𝐡𝐨𝐮𝐬𝐞: 𝐒𝐜𝐡𝐞𝐦𝐚-𝐨𝐧-𝐰𝐫𝐢𝐭𝐞, 𝐝𝐞𝐟𝐢𝐧𝐞 𝐟𝐢𝐫𝐬𝐭 𝐭𝐡𝐞𝐧 𝐬𝐭𝐨𝐫𝐞
    - A centralized storage system optimized for structured data and business intelligence.
    - ✅ Fast queries, strong governance—ideal for BI and compliance.
    - ❌ Rigid schemas, not ideal for raw/unstructured data, expensive at scale.
    - Go-to Tools:Snowflake , BigQuery, Redshift.
  - 💧 𝐃𝐚𝐭𝐚 𝐋𝐚𝐤𝐞: 𝐒𝐜𝐡𝐞𝐦𝐚-𝐨𝐧-𝐫𝐞𝐚𝐝, 𝐬𝐭𝐨𝐫𝐞 𝐟𝐢𝐫𝐬𝐭 𝐭𝐡𝐞𝐧 𝐝𝐞𝐟𝐢𝐧𝐞
    - A centralized repository that stores massive volumes of raw structured and unstructured data in native formats.
    - ✅ Cheap, flexible, great for ML and exploration.
    - ❌ Lacks governance, slower queries without tuning, data swamp risk.
    - Go-to Tools: S3+Glue, Azure Data Lake.
  - 🏞️ 𝐃𝐚𝐭𝐚 𝐋𝐚𝐤𝐞𝐡𝐨𝐮𝐬𝐞: 𝐋𝐚𝐤𝐞 𝐜𝐨𝐬𝐭𝐬 + 𝐖𝐚𝐫𝐞𝐡𝐨𝐮𝐬𝐞 𝐩𝐞𝐫𝐟𝐨𝐫𝐦𝐚𝐧𝐜𝐞
    - A next-gen data platform that combines the flexibility of data lakes with the performance of data warehouses.
    - ✅ Unified storage with strong analytics + ML performance.
    - ❌ Complex to build and operate, tools still evolving.
    - Go-to Tools: databricks , Apache Iceberg.
  - 🌐 𝐃𝐚𝐭𝐚 𝐌𝐞𝐬𝐡: 𝐃𝐚𝐭𝐚 𝐚𝐬 𝐚 𝐩𝐫𝐨𝐝𝐮𝐜𝐭, 𝐝𝐨𝐦𝐚𝐢𝐧 𝐚𝐮𝐭𝐨𝐧𝐨𝐦𝐲
    - A distributed architecture treating data as products, with each business domain owning and managing their own data.
    - ✅ Scales with teams, empowers domain ownership.
    - ❌ High governance overhead, needs strong org maturity.
    - Go-to Tools: Requires combining multiple tools to implement.
- [VictoriaMetrics联创揭示PB级日志处理性能奥秘]
  - 10 秒查询的四级加速路径
    -  压缩：VictoriaLogs 对结构化/半结构化日志可做到 8–50 倍压缩；1 PB 原始日志变 64 TB，扫描时间自 70 h 降到 4.6 h。
    -  列式存储：只读所需列＋针对同质数据做专用编码，使扫描量再降到 4 TB，约 17 min。
    -  时间分区：绝大多数查询带时间条件；按小时/天切分，可跳过 90 % 数据，剩 400 GB，≈100 s。
    -  日志流索引：按 service／pod 等标签拆分流，只扫相关流，再降 90 %，剩 40 GB，≈10 s。 稀疏索引 + append-only 写模式消除了行库随机 I/O 与写放大。）
    -  Bloom Filter（可选）：为每块建立过滤器，做“trace_id 大海捞针”式查询时可再提速百倍。
- [Big Table](https://mp.weixin.qq.com/s/JS_ca31ODCl2MbBSv-xR_g)
  - 为什么二维表不够用
    - • 只能存 Key 和列，无法天然存多版本 Time。
    - • 通过 “Key+Time” 拼接主键虽可勉强存储，但带来
      - 无法按 URL 做前缀/模糊查询
      - 大量空洞行造成存储浪费
      - 单机扩展瓶颈
  - BigTable 的核心模型
    - key + column + timestamp ➜ value
    - • 三维结构：
      - Row Key（如 URL）
      - Column Family/Qualifier（如 content、PV、UV）
      - Timestamp（多版本）
    - • 稀疏、按列存储；天然支持版本管理与按时间排序。
    - • 分布式、持久化，可横向扩容到数千节点。
    - 将 LSM-tree + 列族式模型 + Tablet 分片 + GFS/Chubby 组合成可在数千节点上运行的工业级系统
  - BigTable 解决了什么
    - • 高吞吐、低延迟地存取 PB 级多维度数据。
    - • 统一支撑网页抓取、日志统计、地图、消息等众多内部产品。
- [Parquet格式]
  - 优势
    - Parquet作为一种列式存储的开源文件格式
    - Parquet 支持多种压缩算法（如 Snappy、GZIP 等），并通过字典编码和运行长度编码（RLE）等技术进一步减少数据体积。
    - Parquet 不仅支持基本的数据类型（如整数、字符串等），还支持嵌套数据结构（如数组、映射等）。这使得 Parquet 非常适合存储半结构化数据（如 JSON、XML）
  - 高级特性
    - 对于重复值较多的列（如“城市”），Parquet 会使用字典编码来压缩数据。例如，将“北京”、“上海”、“广州”映射为整数索引，从而减少存储空间。
    - 对于连续重复的值（如“年龄”），Parquet 会使用 RLE 来进一步压缩数据。例如，如果某列的值为 [25, 25, 25, 30, 30]，RLE 会将其编码为 (25, 3), (30, 2)。
    - Parquet 文件的元数据记录了每个行组的最小值和最大值，查询引擎可以根据这些信息跳过不相关的行组
  - 列式存储的 Repetition Level 与 Definition Level
    - Parquet 的 Repetition Level（重复层级和 Definition Level（定义层级） 是处理嵌套数据结构的关键机制，尤其在列式存储中高效编码和重建复杂数据。
    - Repetition Level（重复层级）
      • 作用：标记当前值在嵌套结构的哪个层级开始重复。
      • 通俗理解：当遇到一个数组或列表时，它告诉我们“当前值属于哪个层级的重复结构”。例如，一个用户有多个联系人，每个联系人有多个电话，Repetition Level 会标记电话属于哪个联系人
    - Definition Level（定义层级）
      • 作用：标记当前值在嵌套结构中的存在深度。
      • 通俗理解：如果某个字段是可选的（比如 null），Definition Level 会告诉我们“这个字段的父级路径存在到哪里”。例如，如果字段 a.b.c 存在，而路径 a.b 是必需的，但 c 是可选的，Definition Level 会表示 c 是否存在。
    -  Repetition Level：回答“当前值从哪个层级开始重复”，用于重建数组的嵌套结构。
    - Definition Level：回答“当前值的父级路径存在到哪里”，用于处理可选字段（如 null）
  - Arrow 更关注高性能的数据在内存中的表达与跨系统实时交换，Parquet 则为磁盘/长时间存储而优化，注重压缩和高效存储。
    - Arrow 是“内存数据交换格式”，强调高效的数据分析、计算和跨语言兼容，适用于内存高速计算场景。
    - Parquet 是“磁盘存储格式”，专为数据仓库和大数据存储设计，注重压缩比和 IO 优化。
    - Arrow 的数据结构在 CPU 内部采用紧凑的向量化内存布局，支持 O(1) 随机访问；Parquet 采用分块、分段及字典等多种编码策略，读取时一般批量加载。
  - 数据批量从 Parquet 读取后用 Arrow 表达，再进行内存分析计算；分析结果可再转回 Parquet 持久存储。




















