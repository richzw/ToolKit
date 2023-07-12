
- [基于趋势和季节性的时间序列预测](https://mp.weixin.qq.com/s/Ln4E9iZd3b3EZqeEjNNsag)
  - 时间序列模式
    - 时间序列预测模型使用数学方程(s)在一系列历史数据中找到模式。然后使用这些方程将数据[中的历史时间模式投射到未来。
      - 趋势:数据的长期增减。趋势可以是任何函数，如线性或指数，并可以随时间改变方向。
      - 季节性:以固定的频率(一天中的小时、星期、月、年等)在系列中重复的周期。季节模式存在一个固定的已知周期
      - 周期性:当数据涨跌时发生，但没有固定的频率和持续时间，例如由经济状况引起。
      - 噪音:系列中的随机变化。
    - 当季节波动不随时间序列水平变化时，加法分解是最合适的方法。相反，当季节成分的变化与时间序列水平成正比时，则采用乘法分解更为合适。
  - 分解数据
    - 从数学意义上讲，如果一个时间序列的均值和方差不变，且协方差与时间无关，那么这个时间序列就是平稳的。
    - 如何检验时间序列的平稳性呢?
      - 一方面，我们可以通过检查时间序列的均值和方差来手动检查。另一方面，我们可以使用测试函数来评估平稳性。
      - 查看趋势
        - ADF检验的结果(p值低于0.05)表明，存在的原假设可以在95%的置信水平上被拒绝。因此，如果p值低于0.05，则时间序列是平稳的
        - KPSS检验的结果(p值高于0.05)表明，在95%的置信水平下，不能拒绝的零假设。因此如果p值低于0.05，则时间序列不是平稳的。
        - 统计结果还显示了时间序列的平稳性的影响。虽然两个检验的零假设是相反的。ADF检验表明时间序列是平稳的(p值> 0.05)，而KPSS检验表明时间序列不是平稳的(p值> 0.05)。但这个数据集创建时带有轻微的趋势，因此结果表明，KPSS测试对于分析这个数据集更准确。
      - 检查季节性
        - 正如在之前从滑动窗口中观察到的，在我们的时间序列中有一个季节模式。因此应该采用差分方法来去除时间序列中潜在的季节或周期模式。由于样本数据集具有12个月的季节性，我使用了365个滞后差值:
      - 分解模式
        - 在看了分解图的四个部分后，可以说，在我们的时间序列中有很强的年度季节性成分，以及随时间推移的增加趋势模式
  - 时序建模
    - Autoregression (AR)
    - Moving Average (MA)
    - Autoregressive Moving Average (ARMA)
    - Autoregressive Integrated Moving Average (ARIMA)
    - Seasonal Autoregressive Integrated Moving-Average (SARIMA)
    - Seasonal Autoregressive Integrated Moving-Average with Exogenous Regressors (SARIMAX)
    - Vector Autoregression (VAR)
    - Vector Autoregression Moving-Average (VARMA)
    - Vector Autoregression Moving-Average with Exogenous Regressors (VARMAX)
    - Simple Exponential Smoothing (SES)
    - Holt Winter’s Exponential Smoothing (HWES)
  - 由于我们的数据中存在季节性，因此选择HWES，因为它适用于具有趋势和/或季节成分的时间序列数据。
  - 这种方法使用指数平滑来编码大量的过去的值，并使用它们来预测现在和未来的“典型”值。指数平滑指的是使用指数加权移动平均(EWMA)“平滑”一个时间序列。使用均方根误差(RMSE)作为评估模型误差的度量的实现。
- [AB 实验](https://mp.weixin.qq.com/s/2sE-KxdRAvnp3GBOBU4Cfg)
  - AB 实验需要注意️辛普森悖论、幸存者偏差、选择偏差等，注意事项都是来源于对撞因子，简单来说就是「是指同时被两个以上的变数影响的变数」
  - 如何衡量
    - 对于任何一个想法我们很难去衡量它的好坏，大胆假设小心求证。短期目标可能会与更关键的长期目标发生冲突。
    - 新奇效应如何避免？足够的样本量能保证一个合理的实验周期，可以使用我们的流量计算器中计算流量和实验周期，从而避免这种新奇效应的影响。
  - 架构
    - 流量分割 分流和分层
      - 每个独立实验为一层，层与层之间流量是正交的（简单来讲，就是一份流量穿越每层实验时，都会再次随机打散，且随机效果离散）。实验在同一层拆分流量，不论如何拆分，不同组的流量是不重叠的。
      - 分流是指我们直接将整体用户切割为几块，用户只能在一个实验中。但是这种情况很不现实，因为如果我要同时上线多个实验，流量不够切怎么办？那为了达到最小样本量，我们就得延长实验周期，要是做一个实验，要几个月。
        - 分流是指对流量进行整体切割，实验之间互斥。
        - 目的：为了获取纯净的分区，不会互相影响。
        - 缺点：浪费流量，导致流量不够。
      - 分层就是将同一批用户，不停的随机后，处于不同的桶。也就是说，一个用户会处于多个实验中，只要实验之间不相互影响，我们就能够无限次的切割用户。这样在保证了每个实验都能用全流量切割的同时，也保证了实验数据是置信的。
        - 目的：同一个用户在不同的实验组，相互不会影响。
        - 缺点：不同层之间的 hash 值尽量不要重合。
    - 随机算法
      - 按照密码学来将「随机」分为三种级别：1. 伪随机 (PRNG) 2. 密码学安全的伪随机 (CSPRNG) 3. 真随机 (TRNG)
  - 实验结果显著
    - 两类统计学错误
      - 在统计学的世界里，我们往往只说概率，不说确定，在现实世界中往往只能基于样本进行推断。在 AB 实验中，我们 不知道真实情况是什么，因此做假设检验的时候就会犯错误，这种错误可以划分为两类：
        - 这是第一类错误：实际没有区别，但实验结果表示有区别，我们得到显著结果因此否定原假设，认为实验组更优，发生的概率用 𝛂 表示。
        - 这是第二类错误：实际有区别，但是实际结果表示没有区别，我们得到不显著的结果因此无法拒绝原假设，认为实验组和对照组没有区别，发生的概率用 𝜷 表示。
  - ![img.png](ml_abtest.png)
- [ChatGPT如何获取的超能力](https://mp.weixin.qq.com/s/X5ZcCkuEVtrTz0lJnt5a7w)
  - ChatGPT有人类语言中的所有词（又称token），这是它的搜索空间。
  - 然后，精心选择高质量的文本数据（包括代码），训练Transformer模型，需要很多的GPU算力，进行大量的矩阵运算，达到预定的训练目标即可结束训练。这里，Transformer模型是一个包含所有token的概率模型或开放空间。
  - 然后再用含有人类反馈的强化学习（RLHF）来进一步调整Transformer模型来适应人类的价值观和使用规则。现在，Transformer模型被人类调教后的包含所有token的概率模型或限制空间。
  - 最后，执行任务的时候，就是给出一些提示tokens，或上下文context，在Transformer构成的所有token的限制空间中使用贪婪，集束，温度采用等策略来找到概率最大的可能的token的排列组合。这个组合，就是看到的ChatGPT的输出。
- [Prompt Engineering Guide](https://www.promptingguide.ai/techniques/knowledge)
- [Parameter optimization in neural networks](https://www.deeplearning.ai/ai-notes/optimization/index.html?_hsmi=218814757&utm_campaign=The%20Batch&utm_medium=email&utm_content=218804890&utm_source=hs_email&_hsenc=p2ANqtz-_FluhJbN2619klYO-hikBLp6-aEAP60t0VaLzoiEItfCyfrdJguDchLz7Q6h5imUeQp3SkfQaBZnlD8_aUcP5U97FiMA)
- [Introduction to Uplift Modeling](https://juanitorduz.github.io/uplift/)
- [What is Uplift modelling and how can it be done with CausalML](https://analyticsindiamag.com/what-is-uplift-modelling-and-how-can-it-be-done-with-causalml/)
- [Prometheus for anomaly detection](https://about.gitlab.com/blog/2019/07/23/anomaly-detection-using-prometheus/)
  - z-score
    - z-score is measured in the number of standard deviations from the mean
    - Assuming the underlying data has a normal distribution, 99.7% of the samples should have a z-score between zero to three. The further the z-score is from zero, the less likely it is to exist.
    ```shell
    # Z-Score for aggregation
    (
    job:http_requests:rate5m -
    job:http_requests:rate5m:avg_over_time_1w
    ) /  job:http_requests:rate5m:stddev_over_time_1w
    ```
    - normal distribution?
      - There are numerous statistical techniques for testing your data for a normal distribution, but the best option is to test that your underlying data has a z-score of about +4 to -4.
      ```shell
      (
      max_over_time(job:http_requests:rate5m[1w]) - avg_over_time(job:http_requests:rate5m[1w])
      ) / stddev_over_time(job:http_requests:rate5m[1w])
      
      (
      min_over_time(job:http_requests:rate5m[1w]) - avg_over_time(job:http_requests:rate5m[1w])
      ```
  - Seasonality
    - Seasonality is a characteristic of a time series metric in which the metric experiences regular and predictable changes that recur every cycle.
    ```shell
      quantile(0.5,
         label_replace(
           avg_over_time(job:http_requests:rate5m[4h] offset 166h)
           + job:http_requests:rate5m:avg_over_time_1w - job:http_requests:rate5m:avg_over_time_1w offset 1w
           , "offset", "1w", "", "")
         or
         label_replace(
           avg_over_time(job:http_requests:rate5m[4h] offset 334h)
           + job:http_requests:rate5m:avg_over_time_1w - job:http_requests:rate5m:avg_over_time_1w offset 2w
           , "offset", "2w", "", "")
         or
         label_replace(
           avg_over_time(job:http_requests:rate5m[4h] offset 502h)
           + job:http_requests:rate5m:avg_over_time_1w - job:http_requests:rate5m:avg_over_time_1w offset 3w
           , "offset", "3w", "", "")
       )
       without (offset)
    ```
- [如何用 PPO 算法让 AI 学会玩 FlappyBird](https://mp.weixin.qq.com/s/5DYBCCU3xsmTHtN5Ciz0WA)
- [Ray 的大规模离线推理](https://mp.weixin.qq.com/s/2-jWtYcO0CVnttRrJOYcnA)
  - Ray Core：是 Ray 框架的底层框架，提供了一整套的分布式计算的框架，可以将普通的应用转化成分布式的系统
    - [Ray Core](https://mp.weixin.qq.com/s?__biz=MzA5NTUxNzE4MQ==&mid=2659281279&idx=1&sn=42604ee42f6bad25321e8b38eae34d33&scene=21#wechat_redirect)
    - [Ray Core](https://mp.weixin.qq.com/s?__biz=MzA5NTUxNzE4MQ==&mid=2659281407&idx=1&sn=548bd7f7421714f6262fee7a3c94a8ab&scene=21#wechat_redirect)
  - Ray Serve：是一个可扩展的模型服务库，用于构建在线推理 API
- [Ray 云原生探索之路--分布式构建本地知识库](https://mp.weixin.qq.com/s/K96d-UUnIX0tyWpL6Z7cQA)
  - 本地向量处理
    - 离线:  HuggingFace 的 Embeddings 的模型 “text2vec-large-chinese” 来完成这个能力
    - 基于 pgvector 完成向量处理和向量数据的保存
    - 基于 elasticsearch 完成向量处理和向量数据的保存
  - 串行向量化
    - 串行指的是在处理的过程中没有并发多任务处理能力，有一个 worker 顺序执行的方式去处理整个过程，包括数据文件的读取、文本的拆分以及文本的向量处理，到写入向量数据库。
    - 串行向量化的方式，可以通过 Ray 的 Actor 模型来完成，Actor 模型是 Ray 的核心模型，可以将普通的 Python 类转化成分布式的 Actor，Actor 之间可以通过消息的方式进行通信，Actor 之间的通信是异步的，Actor 之间的通信是通过 Ray 的 Plasma 存储来完成的。
  - 并行向量化
    - 并行指的是在处理的过程中有并发多任务处理能力，有 n 个 worker 并行的方式去运行各种任务。如果在数据量很大的情况下，整个数据的向量化处理能力，会随着可用资源的增多，有很明显的提升。能充分的利用好整个集群的可用资源去处理相关的任务。
    - 并行向量化的方式，可以通过 Ray 的 Task 模型来完成，Task 模型是 Ray 的核心模型，可以将普通的 Python 函数转化成分布式的 Task，Task 之间可以通过消息的方式进行通信，Task 之间的通信是异步的，Task 之间的通信是通过 Ray 的 Plasma 存储来完成的。
  - 向量构建相关
    - CPU 类型的镜像，用于启动 Ray Cluster 的 Head 节点
    - GPU 类型的镜像，用于启动 Ray Cluster 的 Worker 节点
- [LLM Agent](https://lilianweng.github.io/posts/2023-06-23-agent/)
  - Agent = LLM + memory + planning skill + tool use
  - 算法蒸馏（Algorithm Distillation）
    - 将相同的思想应用于强化学习任务中的跨剧情轨迹，其中算法被封装在一个长期历史条件策略中。考虑到代理与环境的多次交互，每一集中代理都会表的更好一些，AD 将这个学习历史连接起来并将其输入到模型中。因此，我们应该期望下一个预测的动作比之前的试验表现更好。我们的目标是学习强化学习的过程，而不是训练一个用于特定任务的策略本身。
  - 思维链（CoT，Chain of thought）
    - 已成为一种标准prompting技术，用于增强复杂任务上的模型性能。指示该模型“逐步思考”，以利用更多的测试时间计算将困难任务分解为更小，更简单的步骤。COT将重大任务转换为多个可管理的任务，并将注意力放到对模型思考过程的可解释性中。
  - 思维树（Tree of Thoughts）
    - 通过探索每个步骤的多种推理可能性来扩展COT。它首先将问题分解为多个思考步骤，并且每个步骤都生成多个想法，从而可以创建一个树形结构。
    - 通过将思维树与算法蒸馏相结合，我们可以将多个思维树的输出连接起来，以形成一个更长的思维链。这种方法可以将复杂的任务分解为更小的任务，从而使模型能够更好地处理复杂的任务。
    - 思维树的搜索过程可以是BFS（广度优先搜索）或DFS（深度优先搜索），每个状态都由分类器（通过prompt）或多数投票决定
  - ReAct
    - 通过将行动空间扩展为特定任务的离散行动和语言空间的组合，将推理和行动集成到 LLM中。前者使 LLM 能够与环境交互（例如使用维基百科搜索API），后者能够促使LLM 生成自然语言的推理轨迹。
  - 反思
    - 是一个框架，它为代理提供动态记忆和自我反思的能力，以提高它的推理技能。反思采用标准的强化学习设置，其中奖励模型提供简单的二元奖励，行动空间遵循 ReAct 中的设置，同时特定任务的行动空间通过语言来增强复杂的推理步骤。在每个行动at之后，Agent会计算一个启发式值ht，并根据自我反思的结果决定是否重置环境以开始新的试验。
  - Chain of Hindsight，CoH
    - （Hindsight可以翻译为“事后诸葛亮”）通过明确呈现一系列过去的输出序列，并为每个输出注释反馈，鼓励模型改进自己的输出
    - 为了避免过拟合，CoH添加了一个正则化项来最大化预训练数据集的对数似然。为了避免捷径和复制（因为反馈序列中有许多常见单词），他们在训练期间随机mask 0%-5%的历史token。
  - 近似最近邻 (ANN)算法
    - 「LSH」（Locality-Sensitive Hashing）」它引入了一种哈希函数，使得相似的输入能以更高的概率映射到相同的桶中，其中桶的数量远小于输入的数量。
    - 「ANNOY（Approximate Nearest Neighbors）」它的核心数据结构是随机投影树，实际是一组二叉树，其中每个非叶子节点表示一个将输入空间分成两半的超平面，每个叶子节点存储一个数据。二叉树是独立且随机构建的，因此在某种程度上，它模仿了哈希函数。ANNOY会在所有树中迭代地搜索最接近查询的那一半，然后不断聚合结果。这个想法与 KD 树非常相关，但更具可扩展性。
    - 「HNSW（Hierarchical Navigable Small World）」它受到小世界网络思想的启发，其中大多数节点可以在很少的步骤内被任何其他节点到触达；例如社交网络的“六度分隔”理论。HNSW构建这些小世界图的层次结构，其中底层结构包含实际数据。中间的层创建快捷方式以加快搜索速度。执行搜索时，HNSW从顶层的随机节点开始，导航至目标。当它无法靠近时，它会向下移动到下一层，直到到达最底层。上层中的每个移动都可能覆盖数据空间中的很长一段距离，而下层中的每个移动都可以细化搜索质量。
    - 「FAISS（facebook AI Similarity Search）」它运行的假设是：高维空间中节点之间的距离服从高斯分布，因此这些数据点之间存在着聚类点。faiss通过将向量空间划分为簇，然后在簇内使用用向量量化。faiss首先使用粗粒度量化方法来查找候选簇，然后进一步使用更精细的量化方法来查找每个簇。
    - 「ScaNN（Scalable Nearest Neighbors）」的主要创新在于各向异性向量量化。它将数据点量化为一个向量，使得它们的内积与原始距离尽可能相似，而不是选择最接近的量化质心点。
- 时间序列异常值检测
  - 正确体现各种指标多样的变化趋势和行为特性
    - 为消除每个分组中的趋势和季节性影响因素，我们利用了 statsmodels 库中强大的 seasonal_decompose（季节性分解函数）
    - 这一函数可以识别并消除每个分组时间序列中的趋势和季节性成分，是将时间序列分解为其核心组成部分的简单方法
  - 异常检测
    - 采用了将时间序列作为输入的 Matrix Profiling（MP）算法。MP 算法还将计算时间序列中每个点的分数，以此测量该值与其他值的差异 - Stumpy
    - MP 的定义为：
      - 一种存储着时间序列中任意子序列与其最近邻的子序列的欧式距离（标准化后的欧氏距离）的向量。
      - 一个时间序列被划分成许多连续的固定长度子序列，并使用欧式距离或其他距离计算方法进行相互间的比较这种比较是通过滑动窗口的方式进行的，直到覆盖了所有可能的组合
    - 最终实现的异常检测方式如下
      - 时间序列数据经过预处理，消除趋势和季节性。
      - 预处理后的数据输入到不同版本的 Matrix Profile 函数中，以提高结果的稳定性：
        · 原始版本 —— 在分析时间序列数据之前，不对其进行任何更改。
        · 移动块抽样版本 —— 将时间序列分割成较小的片段，随机洗牌并创建用于分析的新序列，以减小数据中任何趋势或模式带来的影响。
        · 随机窗口分割版本 —— 将时间序列分割成较小的多个重叠窗口，选择这些窗口的一个随机子集用于分析，以捕捉数据的局部结构，并减小任何趋势或周期性模式带来的影响。
      - 计算每个数据点的周度百分比变化。
      - 每个数据点的最终异常得分，等于 MP 结果之和与周度变化的乘积。
      - 任何超过某个阈值的得分都将被标记为异常，并在表中有所记录。执行 Matrix Profile 旨在检测时间序列数据中的异常点，以优化每日下降的平均程度
  - 趋势检测
    - 某个指标可能不会出现具有警示性的骤变，而会经历一个缓慢持续下降的过程。为识别这种情况，我们采用了 Moving Average Convergence / Divergence（MACD）技术。
    - MACD 是一种趋势跟踪技术，用于分析时间序列数据的趋势。它通过计算两个移动平均线之间的差异来实现这一目的。MACD 由三个主要组件组成：
      - MACD 线 —— 两个移动平均线之间的差异。
      - 信号线 —— MACD 线的移动平均线。
      - MACD 柱 —— MACD 线和信号线之间的差异。
    - MACD 逻辑
      - 时间序列数据经过预处理，去除了趋势和季节性。
      - 使用两个不同的时间窗口参数，对数据使用指数加权移动平均（EWMA）函数。一个参数用于慢速滑动窗口，另一个参数用于快速滑动窗口，这有助于识别数据在不同时间尺度上的趋势。
      - 从慢速趋势中减去快速趋势得到 MACD 曲线，并再次应用指数加权平均，获得 MACD 信号曲线。
      - 步骤 3 的指数移动平均 MACD 信号曲线减去 MACD 曲线，我们会得到 MACD 直方图。这个直方图有助于我们检测时间序列数据中的渐变变化。
- [GPT-4 Architecture, Infrastructure, Training Dataset, Costs, Vision, MoE](https://hub.baai.ac.cn/view/27744)

