package ta

import (
	"fmt"
)

// TaEMA EMA指标结构体(Exponential Moving Average)
// 说明：
//
//	EMA是一种赋予近期数据更高权重的移动平均线指标
//	它通过给予最近的数据更高的权重，使得指标对价格变化的反应更加敏感
type TaEMA struct {
	Values []float64 `json:"values"` // EMA值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateEMA 计算指数移动平均线(Exponential Moving Average)
// 参数：
//   - prices: 价格序列
//   - period: 计算周期
//
// 返回值：
//   - *TaEMA: EMA值序列，长度与输入价格序列相同，前(period-1)个值为0
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	计算步骤：
//	1. 计算权重系数 K = 2/(period + 1)
//	2. 计算初始EMA值（使用period个数据的简单平均）
//	3. 使用递归公式计算后续值：EMA = (K × 当前价格) + ((1 - K) × 前一日EMA)
//	4. 重复步骤3直到计算完所有数据
//
// 示例：
//
//	ema, err := CalculateEMA(prices, 20)
func CalculateEMA(prices []float64, period int) (*TaEMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 1) // [ema]
	result := slices[0]

	// 计算第一个EMA值（使用SMA作为初始值）
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// 预计算乘数，避免重复计算
	multiplier := 2.0 / float64(period+1)
	oneMinusMultiplier := 1.0 - multiplier

	// 使用递推公式计算后续的EMA值
	for i := period; i < length; i++ {
		result[i] = prices[i]*multiplier + result[i-1]*oneMinusMultiplier
	}

	return &TaEMA{
		Values: result,
		Period: period,
	}, nil
}

// EMA 计算指定周期的指数移动平均线序列
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaEMA: EMA值序列
//   - error: 可能的错误信息
//
// 示例：
//
//	ema, err := k.EMA(20, "close")
func (k *KlineDatas) EMA(period int, source string) (*TaEMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateEMA(prices, period)
}

// EMA_ 获取最新的EMA值
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的EMA值，如果计算出错则返回0
//
// 示例：
//
//	value := k.EMA_(20, "close")
func (k *KlineDatas) EMA_(period int, source string) float64 {
	// 只保留必要的计算数据
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	ema, err := CalculateEMA(prices, period)
	if err != nil {
		return 0
	}
	return ema.Value()
}

// Value 返回最新的EMA值
// 返回值：
//   - float64: 最新的EMA值
//
// 示例：
//
//	value := ema.Value()
func (t *TaEMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossOver 判断是否向上穿越指定价格
// 参数：
//   - price: 价格水平
//
// 返回值：
//   - bool: 如果EMA从下向上穿越价格返回true，否则返回false
//
// 示例：
//
//	isCrossOver := ema.IsCrossOver(100)
func (t *TaEMA) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

// IsCrossUnder 判断是否向下穿越指定价格
// 参数：
//   - price: 价格水平
//
// 返回值：
//   - bool: 如果EMA从上向下穿越价格返回true，否则返回false
//
// 示例：
//
//	isCrossUnder := ema.IsCrossUnder(100)
func (t *TaEMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

// IsTrendUp 判断是否处于上升趋势
// 返回值：
//   - bool: 如果EMA向上倾斜返回true，否则返回false
//
// 示例：
//
//	isTrendUp := ema.IsTrendUp()
func (t *TaEMA) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

// IsTrendDown 判断是否处于下降趋势
// 返回值：
//   - bool: 如果EMA向下倾斜返回true，否则返回false
//
// 示例：
//
//	isTrendDown := ema.IsTrendDown()
func (t *TaEMA) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

// GetSlope 获取EMA斜率
// 返回值：
//   - float64: EMA的斜率值，正值表示上升，负值表示下降
//
// 示例：
//
//	slope := ema.GetSlope()
func (t *TaEMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

// IsCrossOverEMA 判断是否与另一个EMA发生交叉
// 参数：
//   - other: 另一个EMA指标
//
// 返回值：
//   - golden: 是否发生金叉（快线上穿慢线）
//   - death: 是否发生死叉（快线下穿慢线）
//
// 示例：
//
//	golden, death := ema.IsCrossOverEMA(otherEma)
func (t *TaEMA) IsCrossOverEMA(other *TaEMA) (golden, death bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	golden = t.Values[lastIndex-1] <= other.Values[lastIndex-1] && t.Values[lastIndex] > other.Values[lastIndex]
	death = t.Values[lastIndex-1] >= other.Values[lastIndex-1] && t.Values[lastIndex] < other.Values[lastIndex]
	return
}

// IsAccelerating 判断趋势是否在加速
// 返回值：
//   - bool: 如果趋势在加速返回true，否则返回false
//
// 说明：
//
//	通过比较连续三个值的变化率来判断趋势是否在加速
//
// 示例：
//
//	isAccelerating := ema.IsAccelerating()
func (t *TaEMA) IsAccelerating() bool {
	if len(t.Values) < 3 {
		return false
	}
	lastIndex := len(t.Values) - 1
	diff1 := t.Values[lastIndex] - t.Values[lastIndex-1]
	diff2 := t.Values[lastIndex-1] - t.Values[lastIndex-2]
	return (diff1 > 0 && diff1 > diff2) || (diff1 < 0 && diff1 < diff2)
}

// GetTrendStrength 获取趋势强度
// 返回值：
//   - float64: 趋势强度的百分比值
//
// 说明：
//
//	通过计算周期内的价格变化百分比来衡量趋势强度
//
// 示例：
//
//	strength := ema.GetTrendStrength()
func (t *TaEMA) GetTrendStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	startIndex := lastIndex - t.Period + 1
	startValue := t.Values[startIndex]
	endValue := t.Values[lastIndex]
	return (endValue - startValue) / startValue * 100
}
