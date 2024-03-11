- [RAG 系统开发中的 12 大痛点及解决方案](https://baoyu.io/translations/rag/12-rag-pain-points-and-proposed-solutions)
  - 缺失内容
    - 数据清洗的重要性
    - 精心设计的提示有助于提高准确性
  - 关键文档被遗漏
    - 通过调整 chunk_size 和 similarity_top_k 参数优化检索效果 -  [LlamaIndex 实现超参数自动调整](https://levelup.gitconnected.com/automating-hyperparameter-tuning-with-llamaindex-72fdd68e3b90)
    - 检索结果的优化排序 - CohereRerank 进行优化排序
  - 文档整合限制 —— 超出上下文
    
- [ Semantic Chunking for RAG](https://pub.towardsai.net/advanced-rag-05-exploring-semantic-chunking-97c12af20a4d)
  - Embedding-based chunking 
  -  BERT-based chunking techniques (naive, cross-segment, SeqModel)
  -  LLM-based chunking
- [提升RAG检索质量](https://mp.weixin.qq.com/s/rjsymQQwgE78fNZ60opE0w)
  - 查询扩展（Query expansion）
    - 使用生成的答案进行查询扩展
    - 用多个相关问题扩展查询
      - 利用 LLM 生成 N 个与原始查询相关的问题，然后将所有问题（加上原始查询）发送给检索系统
  - 跨编码器重排序（Cross-encoder re-ranking）
    - 根据输入查询与检索到的文档的相关性的分数对文档进行重排序
    - 交叉编码器是一种深度神经网络，它将两个输入序列作为一个输入进行处理。这样，模型就能直接比较和对比输入，以更综合、更细致的方式理解它们之间的关系。
  - 嵌入适配器（Embedding adaptors），可以支持检索到更多与用户查询密切匹配的相关文档
    - 适配器是以小型前馈神经网络的形式实现的，插入到预训练模型的层之间。训练适配器的根本目的是改变嵌入查询，从而为特定任务产生更好的检索结果。
    - 嵌入适配器是在嵌入阶段之后、检索之前插入的一个阶段。可以把它想象成一个矩阵（带有经过训练的权重），它采用原始嵌入并对其进行缩放。


































