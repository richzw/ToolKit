
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






