
- [手机居然可以这么偷听你的秘密](https://mp.weixin.qq.com/s/U8fZbVgEHmKSZmt62XFEkw)
  - https://www.ndss-symposium.org/wp-content/uploads/2020/02/24076-paper.pdf
  - 可以采集内置加速传感器的信号，然后通过深度学习算法来解析出语音文字
  - 智能手机加速度传感器的采样频率在持续提升，足以覆盖人的语音的频段
- 37% rule 和 Optimal Stop Theory
  - 选择一种策略，总计N个选项，拒绝前K个选项，从K+1个选项开始只要看到比前K个选项优的选项则选择；K为多少时我们的策略能选到最优解的概率最大？
- copilot-explorer
  - https://thakkarparth007.github.io/copilot-explorer/posts/copilot-internals
  - https://mp.weixin.qq.com/s/dtfLeEfcwbz3fb4mLROVYQ
- Google Tools
  - [Market Finder](https://marketfinder.thinkwithgoogle.com/intl/en)
  - [Google Trends](https://trends.google.com/trends?)
  - [Tools](https://www.thinkwithgoogle.com/tools/)
    - Find my audience
- [Node.js Addon](https://mp.weixin.qq.com/s/6Qm5DpNWEyCBkI9Fh_Z0CA)
- [搜索引擎]
  - 多模态语义理解技术在用户意图分析、向量召回、倒排召回以及相关性排序四个方面的业务实践。
  - 短文本理解是用户意图分析的主要手段
    - 对于长尾流量，利用知识库、实体链接方法，将实体的附加信息引入到判别模型，提高长尾 Query 预测准确性；
    - 对于头部流量，采用日志挖掘、系统模拟的方式引入后验数据，提高头部 Query 的预测准确率。
  - 向量检索
    - 跨模态对齐：通过“笔记中的图片和文字“的对比学习、“ Query 和图片“的对比学习，将文本和图像表示到同一个语义空间中。
    - 多模态融合：尝试多种模态融合模型结构，引入多模态 Mask Language Modeling（MLM） 和 Mask Image Modeling (MIM) ，以实现更好的多模态信息融合。
    - 负样本的构造：通过对 Query 和图像进行 Masking、改写和替换，来构造困难的负样本
  -  倒排索引
    - 第一，为笔记生成 Query。针对曝光量较小的笔记，使用生成式模型生成 Query，从而有效提高长尾笔记的召回率。
    - 第二，将多模态内容转化成文本。团队通过视频全文生成技术，生成视频的转写文本，此类语料用于倒排索引中，能在不影响相关性指标的前提下，显著提高视频的召回率。
    - 第三，对笔记进行篇章级的标签提取。团队通过笔记内容与标签的相关性算法剔除无关的 Hashtag（用户上传标签），获取的 Hashtag 语料可以通过弱监督训练来增强多模态内容理解模型能力
  -  相关性排序
    - 多阶段的语言模型训练范式、推理效率问题以及多模态相关性。在相关性训练中，语言模型训练可分为三个阶段：
      - 预训练阶段使用站内文本语料进行无监督预训练；
      - 连续预训练阶段在预训练模型基础上使用搜索日志进行监督训练；
      - 微调阶段在连续训练模型基础上使用人工标注语料进行监督训练。
- [Protobuf编码](https://mp.weixin.qq.com/s/hAfrPlPD2KBCWxpIuGkQTQ)
  - 基本类型
    - int32、int64、uint32、uint64会直接使用varint编码，
    - bool类型会直接使用一个字节存储，
    - enum可以看成是一个int32类型。
    - 对于sint32、sint64类型会先进行zigzag编码，再进行varint编码
    - varint编码：变长编码，对于小正整数有较好的压缩效果，对于大整数或负数编码后字节流长度会变大。
    - zigzag编码：定长编码，将小正整数和小负整数转换到小正整数再进行varint编码，对绝对值较小的整数有良好的压缩效果。
  -  复合类型
    - map的底层存储key-value键值对，采用和数组类型一样的存储方法，数组中每个元素是kv键值对
    - 结构体类型 typeid、length、data三部分长度会根据实际情况发生改变
  - protobuf既然有了int32 为什么还要用sint32 和 fixed32 ？
    - int32使用varint编码，对于小正数有较好的压缩效果，对于大整数和负数会导致额外的字节开销。因此引入fixed32，该类型不会对数值进行任何编码，对大于2^28-1的整数比int32占用更少的字节。而对于负数使用zigzag编码，这样绝对值较小的负数都能被有效压缩。
- [Protobuf 动态反射 - Dynamicgo](https://mp.weixin.qq.com/s/OeQwlgZJtYOGTHnN50IdOA)
- [优秀程序员的共性特征](https://mp.weixin.qq.com/s/FKRedldguFVPred7johg8A)
  - 偏执 - 当所有人都真的在给你找麻烦的时候，偏执就是一个好主意
  - 控制软件的熵 
  - 为测试做设计 - 在编码时就考虑怎么测试。不然，你永远没有机会考虑了 
  - 不要面向需求编程 - 应该面向业务模型编程























