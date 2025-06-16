# 技术分析指标库 (Technical Analysis Library)

这是一个用Go语言实现的技术分析指标库，提供了常用的技术分析指标计算功能。

兼容`go-binance`库K线数据结构。

## 项目结构

- `ta.go`: 核心数据结构和通用工具函数
- `macd.go`: MACD(移动平均趋势指标)
- `ema.go`: EMA(指数移动平均线)
- `sma.go`: SMA(简单移动平均线)
- `superTrend.go`: SuperTrend(超级趋势指标)
- `superTrendPivot.go`: SuperTrend的轴点计算实现
- `boll.go`: BOLL(布林带)
- `atr.go`: ATR(平均真实波幅)
- `volatilityRatio.go`: 波动比率指标
- `rsi.go`: RSI(相对强弱指标)
- `stochRsi.go`: Stochastic RSI(随机相对强弱指标)
- `williamsR.go`: Williams %R(威廉指标)
- `cci.go`: CCI(顺势指标)
- `obv.go`: OBV(能量潮指标)
- `cmf.go`: CMF(蔡金货币流量)
- `kdj.go`: KDJ(随机指标)
- `adx.go`: ADX(平均趋向指标)
- `t3.go`: T3(三重指数移动平均线)
- `rma.go`: RMA(移动平均)
- `superTrendPivotHl2.go`: SuperTrend的HL2轴点计算实现

## 使用示例

```go

binanceKline, err := binance.NewFuturesKlinesService().  
    Limit(1000).
    Symbol("BTCUSDT").
    Interval("1H").
    Do(context.Background())
if err != nil {
    log.Fatal(err)
}


kline, err := ta.NewKlineDatas(binanceKline, true)
if err != nil {
    log.Fatal(err)
}


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
```

## 免责声明

本项目仅提供技术分析工具，不构成投资建议。合约交易有高杠杆风险，请谨慎使用。