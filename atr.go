package ta

import (
	"fmt"
	"math"
)

// TaATR ATR指标结构体(Average True Range)
type TaATR struct {
	Values    []float64 `json:"values"`     // ATR值序列
	Period    int       `json:"period"`     // 计算周期
	TrueRange []float64 `json:"true_range"` // 真实波幅序列
}

// CalculateATR 计算平均真实波幅(Average True Range)
// 参数：
//   - klineData: K线数据集合
//   - period: 计算周期，通常为14
//
// 返回值：
//   - *TaATR: ATR指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	ATR是衡量市场波动性的重要指标，可用于判断市场趋势强度和设置止损点
//	计算步骤：
//	1. 计算真实波幅TR = max(high-low, |high-prevClose|, |low-prevClose|)
//	2. 计算第一个ATR值为前period个TR的简单平均
//	3. 使用Wilder平滑法计算后续ATR：ATR = (前一日ATR * (period-1) + 当日TR) / period
//	4. 重复步骤3直到计算完所有数据
//
// 示例：
//
//	atr, err := CalculateATR(klineData, 14)
func CalculateATR(klineData KlineDatas, period int) (*TaATR, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 2) // [tr, atr]
	trueRange, atr := slices[0], slices[1]

	// 计算真实范围(TR)
	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevClose := klineData[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	// 计算初始ATR
	var sumTR float64
	for i := 1; i <= period; i++ {
		sumTR += trueRange[i]
	}
	atr[period] = sumTR / float64(period)

	// 计算后续ATR值
	for i := period + 1; i < length; i++ {
		atr[i] = (atr[i-1]*(float64(period)-1) + trueRange[i]) / float64(period)
	}

	return &TaATR{
		Values:    atr,
		Period:    period,
		TrueRange: trueRange,
	}, nil
}

// ATR 计算K线数据的平均真实范围(Average True Range)
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - *TaATR: ATR指标结构体
//   - error: 可能的错误信息
//
// 示例：
//
//	atr, err := k.ATR(14)
func (k *KlineDatas) ATR(period int) (*TaATR, error) {
	return CalculateATR(*k, period)
}

// ATR_ 计算最新的一个ATR值
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - float64: 最新的ATR值，如果计算出错则返回0
//
// 示例：
//
//	value := k.ATR_(14)
func (k *KlineDatas) ATR_(period int) float64 {
	_k, err := k._Keep(period * 2)
	if err != nil {
		_k = *k
	}
	atr, err := CalculateATR(_k, period)
	if err != nil {
		return 0
	}
	return atr.Value()
}

// Value 返回最新的ATR值
// 返回值：
//   - float64: 最新的ATR值
//
// 示例：
//
//	value := atr.Value()
func (t *TaATR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// GetTrueRange 返回最新的真实波幅值
// 返回值：
//   - float64: 最新的真实波幅值
//
// 示例：
//
//	tr := atr.GetTrueRange()
func (t *TaATR) GetTrueRange() float64 {
	return t.TrueRange[len(t.TrueRange)-1]
}

// IsVolatilityHigh 判断波动率是否较高
// 参数：
//   - threshold: 波动率阈值
//
// 返回值：
//   - bool: 如果当前波动率高于阈值返回true，否则返回false
//
// 示例：
//
//	isHigh := atr.IsVolatilityHigh(1.5)
func (t *TaATR) IsVolatilityHigh(threshold float64) bool {
	if len(t.Values) < t.Period {
		return false
	}
	lastIndex := len(t.Values) - 1
	avgATR := 0.0
	for i := 0; i < t.Period; i++ {
		avgATR += t.Values[lastIndex-i]
	}
	avgATR /= float64(t.Period)
	return t.Values[lastIndex] > avgATR*threshold
}

// IsVolatilityLow 判断波动率是否较低
// 参数：
//   - threshold: 波动率阈值
//
// 返回值：
//   - bool: 如果当前波动率低于阈值返回true，否则返回false
//
// 示例：
//
//	isLow := atr.IsVolatilityLow(0.5)
func (t *TaATR) IsVolatilityLow(threshold float64) bool {
	if len(t.Values) < t.Period {
		return false
	}
	lastIndex := len(t.Values) - 1
	avgATR := 0.0
	for i := 0; i < t.Period; i++ {
		avgATR += t.Values[lastIndex-i]
	}
	avgATR /= float64(t.Period)
	return t.Values[lastIndex] < avgATR*threshold
}

// IsVolatilityIncreasing 判断波动率是否在增加
// 返回值：
//   - bool: 如果波动率在增加返回true，否则返回false
//
// 示例：
//
//	isIncreasing := atr.IsVolatilityIncreasing()
func (t *TaATR) IsVolatilityIncreasing() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

// IsVolatilityDecreasing 判断波动率是否在减少
// 返回值：
//   - bool: 如果波动率在减少返回true，否则返回false
//
// 示例：
//
//	isDecreasing := atr.IsVolatilityDecreasing()
func (t *TaATR) IsVolatilityDecreasing() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

// GetVolatilityChange 获取波动率变化百分比
// 返回值：
//   - float64: 波动率变化的百分比值
//
// 示例：
//
//	change := atr.GetVolatilityChange()
func (t *TaATR) GetVolatilityChange() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return (t.Values[lastIndex] - t.Values[lastIndex-1]) / t.Values[lastIndex-1] * 100
}

// GetStopLoss 计算基于ATR的止损价格
// 参数：
//   - currentPrice: 当前价格
//   - multiplier: ATR乘数
//
// 返回值：
//   - float64: 计算得到的止损价格
//
// 示例：
//
//	stopLoss := atr.GetStopLoss(100.0, 2.0)
func (t *TaATR) GetStopLoss(currentPrice float64, multiplier float64) (stopLoss float64) {
	atr := t.Value()
	return currentPrice - atr*multiplier
}

// GetTakeProfit 计算基于ATR的止盈价格
// 参数：
//   - currentPrice: 当前价格
//   - multiplier: ATR乘数
//
// 返回值：
//   - float64: 计算得到的止盈价格
//
// 示例：
//
//	takeProfit := atr.GetTakeProfit(100.0, 2.0)
func (t *TaATR) GetTakeProfit(currentPrice float64, multiplier float64) (takeProfit float64) {
	atr := t.Value()
	return currentPrice + atr*multiplier
}

// GetChannelBounds 获取基于ATR的通道边界
// 参数：
//   - currentPrice: 当前价格
//   - multiplier: ATR乘数
//
// 返回值：
//   - upper: 上边界价格
//   - lower: 下边界价格
//
// 示例：
//
//	upper, lower := atr.GetChannelBounds(100.0, 2.0)
func (t *TaATR) GetChannelBounds(currentPrice float64, multiplier float64) (upper, lower float64) {
	atr := t.Value()
	upper = currentPrice + atr*multiplier
	lower = currentPrice - atr*multiplier
	return
}

// IsBreakingOut 判断是否发生突破
// 参数：
//   - price: 当前价格
//   - prevPrice: 前一个价格
//
// 返回值：
//   - bool: 如果发生突破返回true，否则返回false
//
// 示例：
//
//	isBreakout := atr.IsBreakingOut(100.0, 95.0)
func (t *TaATR) IsBreakingOut(price, prevPrice float64) bool {
	if len(t.Values) < 1 {
		return false
	}
	atr := t.Value()
	priceChange := math.Abs(price - prevPrice)
	return priceChange > atr
}

// GetVolatilityRatio 获取波动率比率
// 返回值：
//   - float64: 当前ATR与平均ATR的比率
//
// 示例：
//
//	ratio := atr.GetVolatilityRatio()
func (t *TaATR) GetVolatilityRatio() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	currentATR := t.Values[lastIndex]
	var sumATR float64
	for i := 0; i < t.Period; i++ {
		sumATR += t.Values[lastIndex-i]
	}
	avgATR := sumATR / float64(t.Period)
	return currentATR / avgATR
}
