package ta

import (
	"fmt"
)

// TaSMA SMA指标结构体(Simple Moving Average)
// 说明：
//
//	SMA是最基础的移动平均线指标，用于平滑价格数据并识别趋势
//	它通过计算一定周期内价格的算术平均值来反映价格的总体趋势
type TaSMA struct {
	Values []float64 `json:"values"` // SMA值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateSMA 计算简单移动平均线(Simple Moving Average)
// 参数：
//   - prices: 价格序列
//   - period: 计算周期
//
// 返回值：
//   - *TaSMA: 包含SMA值序列和计算周期的结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	计算步骤：
//	1. 选定计算周期N
//	2. 计算N个周期内价格的总和
//	3. 将总和除以周期N得到平均值
//	4. 移动计算窗口重复步骤2-3
//
// 示例：
//
//	sma, err := CalculateSMA(prices, 20)
func CalculateSMA(prices []float64, period int) (*TaSMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 1) // [sma]
	sma := slices[0]

	// 计算初始SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	sma[period-1] = sum / float64(period)

	// 计算后续SMA值
	for i := period; i < length; i++ {
		sum += prices[i] - prices[i-period]
		sma[i] = sum / float64(period)
	}

	return &TaSMA{
		Values: sma,
		Period: period,
	}, nil
}

// SMA 计算K线数据的简单移动平均线
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaSMA: 包含SMA值序列和计算周期的结构体
//   - error: 可能的错误信息
//
// 示例：
//
//	sma, err := k.SMA(20, "close")
func (k *KlineDatas) SMA(period int, source string) (*TaSMA, error) {
	prices, err := k._ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateSMA(prices, period)
}

// SMA_ 计算最新的一个SMA值
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的SMA值，如果计算出错则返回0
//
// 示例：
//
//	value := k.SMA_(20, "close")
func (k *KlineDatas) SMA_(period int, source string) float64 {
	_k, err := k._Keep(period * 2)
	if err != nil {
		_k = *k
	}
	prices, err := _k._ExtractSlice(source)
	if err != nil {
		return 0
	}
	sma, err := CalculateSMA(prices, period)
	if err != nil {
		return 0
	}
	return sma.Value()
}

// Value 返回最新的SMA值
// 返回值：
//   - float64: 最新的SMA值
//
// 示例：
//
//	value := sma.Value()
func (t *TaSMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossOver 判断是否向上穿越指定价格
// 参数：
//   - price: 价格水平
//
// 返回值：
//   - bool: 如果SMA从下向上穿越价格返回true，否则返回false
//
// 示例：
//
//	isCrossOver := sma.IsCrossOver(100)
func (t *TaSMA) IsCrossOver(price float64) bool {
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
//   - bool: 如果SMA从上向下穿越价格返回true，否则返回false
//
// 示例：
//
//	isCrossUnder := sma.IsCrossUnder(100)
func (t *TaSMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

// IsTrendUp 判断是否处于上升趋势
// 返回值：
//   - bool: 如果SMA向上倾斜返回true，否则返回false
//
// 示例：
//
//	isTrendUp := sma.IsTrendUp()
func (t *TaSMA) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

// IsTrendDown 判断是否处于下降趋势
// 返回值：
//   - bool: 如果SMA向下倾斜返回true，否则返回false
//
// 示例：
//
//	isTrendDown := sma.IsTrendDown()
func (t *TaSMA) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

// GetSlope 获取SMA斜率
// 返回值：
//   - float64: SMA的斜率值，正值表示上升，负值表示下降
//
// 示例：
//
//	slope := sma.GetSlope()
func (t *TaSMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

// IsCrossOverSMA 判断是否与另一个SMA发生交叉
// 参数：
//   - other: 另一个SMA指标
//
// 返回值：
//   - golden: 是否发生金叉（快线上穿慢线）
//   - death: 是否发生死叉（快线下穿慢线）
//
// 示例：
//
//	golden, death := sma.IsCrossOverSMA(otherSma)
func (t *TaSMA) IsCrossOverSMA(other *TaSMA) (golden, death bool) {
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
//	isAccelerating := sma.IsAccelerating()
func (t *TaSMA) IsAccelerating() bool {
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
//	strength := sma.GetTrendStrength()
func (t *TaSMA) GetTrendStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	startIndex := lastIndex - t.Period + 1
	startValue := t.Values[startIndex]
	endValue := t.Values[lastIndex]
	return (endValue - startValue) / startValue * 100
}

// GetDeviation 获取当前价格与SMA的偏离度
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - float64: 价格与SMA的偏离百分比
//
// 说明：
//
//	偏离度 = (价格 - SMA) / SMA * 100
//
// 示例：
//
//	deviation := sma.GetDeviation(100)
func (t *TaSMA) GetDeviation(price float64) float64 {
	lastValue := t.Value()
	return (price - lastValue) / lastValue * 100
}

// IsOverbought 判断是否处于超买状态
// 参数：
//   - price: 当前价格
//   - threshold: 超买阈值
//
// 返回值：
//   - bool: 如果偏离度大于阈值返回true，否则返回false
//
// 示例：
//
//	isOverbought := sma.IsOverbought(100, 10)
func (t *TaSMA) IsOverbought(price float64, threshold float64) bool {
	return t.GetDeviation(price) > threshold
}

// IsOversold 判断是否处于超卖状态
// 参数：
//   - price: 当前价格
//   - threshold: 超卖阈值
//
// 返回值：
//   - bool: 如果偏离度小于负阈值返回true，否则返回false
//
// 示例：
//
//	isOversold := sma.IsOversold(100, 10)
func (t *TaSMA) IsOversold(price float64, threshold float64) bool {
	return t.GetDeviation(price) < -threshold
}
