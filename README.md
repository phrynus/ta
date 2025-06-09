# 技术分析指标库 (Technical Analysis Library)

这是一个用Go语言实现的技术分析指标库，提供了常用的技术分析指标计算功能。

## 功能特点

- 支持多种技术分析指标
- 提供完整的信号判断功能
- 统一的API设计和错误处理
- 详细的文档和使用示例
- 高性能的计算实现

## 项目结构

- `ta.go`: 核心数据结构和通用工具函数
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


## 使用示例

```go
// 初始化K线数据
kline := &KlineDatas{...}

// 计算MACD
macd, err := kline.MACD(12, 26, 9)
if err != nil {
    return err
}

// 判断是否出现金叉买入信号
if macd.IsGoldenCross() {
    // 执行买入逻辑
}

// 计算RSI
rsi, err := kline.RSI(14)
if err != nil {
    return err
}

// 判断是否超买
if rsi.IsOverbought() {
    // 执行卖出逻辑
}
```

## 注意事项

1. 所有指标都需要足够的历史数据才能产生有效信号
2. 建议将多个指标结合使用，互相验证
3. 不同指标适用于不同的市场环境
4. 参数可以根据实际需求进行调整
5. 技术指标仅供参考，需要结合基本面分析和市场环境

## 免责声明

本项目仅提供技术分析指标的计算功能，不涉及任何交易策略或投资建议。请在使用前自行评估风险。