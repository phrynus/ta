# 技术分析指标库 (Technical Analysis Library)

这是一个用Go语言实现的技术分析指标库，提供了常用的技术分析指标计算功能。

兼容`go-binance`库K线数据结构。

## 项目结构

- `ta.go`: 核心数据结构和通用工具函数
- `mlFactor.go`: MLFactor(机器学习因子系统)
- `dtr.go`: CalculateDTR(计算决策树回归指标)
- `macd.go`: MACD(移动平均趋势指标)的实现
- `ema.go`: EMA(指数移动平均线)的实现
- `sma.go`: SMA(简单移动平均线)的实现
- `superTrend.go`: SuperTrend(超级趋势指标)的实现
- `superTrendPivot.go`: SuperTrend的轴点计算实现
- `boll.go`: BOLL(布林带)的实现
- `atr.go`: ATR(平均真实波幅)的实现
- `volatilityRatio.go`: 波动比率指标的实现
- `rsi.go`: RSI(相对强弱指标)的实现
- `stochRsi.go`: Stochastic RSI(随机相对强弱指标)的实现
- `williamsR.go`: Williams %R(威廉指标)的实现
- `cci.go`: CCI(顺势指标)的实现
- `obv.go`: OBV(能量潮指标)的实现
- `cmf.go`: CMF(蔡金货币流量)的实现
- `ichimoku.go`: Ichimoku Cloud(一目均衡图)的实现
- `kdj.go`: KDJ(随机指标)的实现
- `adx.go`: ADX(平均趋向指标)的实现
- `t3.go`: T3(三重指数移动平均线)的实现
- `rma.go`: RMA(移动平均)的实现
- `superTrendPivotHl2.go`: SuperTrend的HL2轴点计算实现
- `kama.go`: KAMA (考夫曼自适应移动平均线)
- `deepts.go`: DeepTS (深度学习时间序列)
- `dtr.go`: CalculateDTR(计算决策树回归指标)

## 使用示例

```go
// 1. 初始化binanceKline数据
binanceKline, err := binance.NewFuturesKlinesService().  // 使用期货K线接口
    Limit(1000).
    Symbol("BTCUSDT").
    Interval("1H").
    Do(context.Background())
if err != nil {
    log.Fatal(err)
}

// 2. 转换为ta.KlineDatas格式
kline, err := ta.NewKlineDatas(binanceKline, true)
if err != nil {
    log.Fatal(err)
}

// 3. 计算技术指标
macd, err := kline.MACD("close", 12, 26, 9)
if err != nil {
    log.Fatal(err)
}

rsi, err := kline.RSI(14, "close")
if err != nil {
    log.Fatal(err)
}

atr, err := kline.ATR(14)
if err != nil {
    log.Fatal(err)
}

// 4. 计算机器学习因子
mlFactor, err := kline.MLFactor(20, "ensemble")
if err != nil {
    log.Fatal(err)
}

prediction, probability, confidence := mlFactor.Value()
log.Printf("预测涨幅: %.2f%%", prediction*100)
log.Printf("预测概率: %.2f%%", probability*100)
log.Printf("预测置信度: %.2f%%", confidence*100)

```

## 注意事项

1. 合约交易风险较大，建议：
   - 严格执行止损
   - 控制杠杆倍数
   - 单笔风险不超过账户的2%
   - 总持仓风险不超过账户的6%

2. 技术指标使用建议：
   - 趋势市场：以MACD、SuperTrend为主
   - 震荡市场：以RSI、KDJ为主
   - 突破行情：以ATR、Bollinger Bands为主

3. 风险控制要点：
   - 设置合理的止损位
   - 使用追踪止损保护利润
   - 根据波动率动态调整仓位
   - 避免过度交易

4. 机器学习预测注意：
   - 预测概率低于60%时谨慎入场
   - 趋势强度低于0.5时减小仓位
   - 定期重新训练模型

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。合约交易有高杠杆风险，请谨慎使用。