package ta

import (
	"fmt"
)

// Package ta 提供技术分析指标的计算功能

// TaRMA RMA指标结构体(Running Moving Average)
// 说明：
//
//	RMA是一种特殊的移动平均线指标，使用Wilder平滑法计算
//	它比简单移动平均线(SMA)和指数移动平均线(EMA)具有更好的平滑效果
//	主要应用场景：
//	1. 趋势跟踪：识别和确认市场趋势
//	2. 价格过滤：减少价格波动中的噪音
//	3. 支撑阻力：提供动态支撑和阻力位
type TaRMA struct {
	Values []float64 `json:"values"` // RMA值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateRMA 计算移动平均(Running Moving Average)
// 参数：
//   - prices: 价格序列
//   - period: 计算周期
//
// 返回值：
//   - *TaRMA: RMA指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	RMA是一种特殊的移动平均线指标，使用Wilder平滑法计算
//	计算步骤：
//	1. 设定平滑系数 alpha = 1/period
//	2. 第一个值使用第一个价格
//	3. 后续值使用公式：RMA = alpha * price + (1 - alpha) * prevRMA
//	4. 重复步骤3直到计算完所有数据
//
// 示例：
//
//	rma, err := CalculateRMA(prices, 14)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前RMA值：%.2f\n", rma.Values[len(rma.Values)-1])
func CalculateRMA(prices []float64, period int) (*TaRMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 1) // [rma]
	rma := slices[0]

	alpha := 1.0 / float64(period)
	rma[0] = prices[0]

	for i := 1; i < length; i++ {
		rma[i] = alpha*prices[i] + (1-alpha)*rma[i-1]
	}

	return &TaRMA{
		Values: rma,
		Period: period,
	}, nil
}

// RMA 计算K线数据的修正移动平均线
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaRMA: RMA指标结构体
//   - error: 可能的错误信息
//
// 说明：
//
//	RMA适用于需要平滑处理的各种价格数据
//	特点：
//	1. 对价格变化反应较慢，但更稳定
//	2. 能有效过滤短期价格波动
//	3. 适合中长期趋势分析
//
// 示例：
//
//	rma, err := kline.RMA(14, "close")
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) RMA(period int, source string) (*TaRMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRMA(prices, period)
}

// RMA_ 获取最新的RMA值
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的RMA值，如果计算失败则返回0
//
// 示例：
//
//	value := kline.RMA_(14, "close")
func (k *KlineDatas) RMA_(period int, source string) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	rma, err := CalculateRMA(prices, period)
	if err != nil {
		return 0
	}
	return rma.Value()
}

// Value 返回最新的RMA值
// 返回值：
//   - float64: 最新的RMA值
//
// 示例：
//
//	value := rma.Value()
func (t *TaRMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossOver 判断是否向上穿越指定价格
// 参数：
//   - price: 目标价格
//
// 返回值：
//   - bool: 如果RMA从下向上穿越价格返回true，否则返回false
//
// 说明：
//
//	向上穿越通常是买入信号
//
// 示例：
//
//	isCrossOver := rma.IsCrossOver(100)
func (t *TaRMA) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

// IsCrossUnder 判断是否向下穿越指定价格
// 参数：
//   - price: 目标价格
//
// 返回值：
//   - bool: 如果RMA从上向下穿越价格返回true，否则返回false
//
// 说明：
//
//	向下穿越通常是卖出信号
//
// 示例：
//
//	isCrossUnder := rma.IsCrossUnder(100)
func (t *TaRMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

// IsTrendUp 判断是否处于上升趋势
// 返回值：
//   - bool: 如果RMA值在上升返回true，否则返回false
//
// 说明：
//
//	通过比较当前值和前一个值来判断趋势方向
//
// 示例：
//
//	isUp := rma.IsTrendUp()
func (t *TaRMA) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

// IsTrendDown 判断是否处于下降趋势
// 返回值：
//   - bool: 如果RMA值在下降返回true，否则返回false
//
// 说明：
//
//	通过比较当前值和前一个值来判断趋势方向
//
// 示例：
//
//	isDown := rma.IsTrendDown()
func (t *TaRMA) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

// GetSlope 获取RMA斜率
// 返回值：
//   - float64: RMA值的变化率
//
// 说明：
//
//	斜率反映了趋势的强度，正值表示上升趋势，负值表示下降趋势
//
// 示例：
//
//	slope := rma.GetSlope()
func (t *TaRMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

// IsCrossOverRMA 判断是否与另一个RMA发生交叉
// 参数：
//   - other: 另一个RMA指标
//
// 返回值：
//   - golden: 金叉信号(买入)
//   - death: 死叉信号(卖出)
//
// 说明：
//
//	两条RMA线的交叉通常用于判断趋势的转换
//
// 示例：
//
//	golden, death := rma.IsCrossOverRMA(otherRma)
func (t *TaRMA) IsCrossOverRMA(other *TaRMA) (golden, death bool) {
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
//	通过比较连续三个值的变化来判断趋势是否在加速
//
// 示例：
//
//	isAccelerating := rma.IsAccelerating()
func (t *TaRMA) IsAccelerating() bool {
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
//   - float64: 趋势强度，用百分比表示
//
// 说明：
//
//	通过计算周期内的价格变化百分比来衡量趋势强度
//
// 示例：
//
//	strength := rma.GetTrendStrength()
func (t *TaRMA) GetTrendStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	startIndex := lastIndex - t.Period + 1
	startValue := t.Values[startIndex]
	endValue := t.Values[lastIndex]
	return (endValue - startValue) / startValue * 100
}

// GetDeviation 获取当前价格与RMA的偏离度
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - float64: 偏离度，用百分比表示
//
// 说明：
//
//	偏离度可用于判断价格是否过度偏离均线
//
// 示例：
//
//	deviation := rma.GetDeviation(100)
func (t *TaRMA) GetDeviation(price float64) float64 {
	lastValue := t.Value()
	return (price - lastValue) / lastValue * 100
}

// IsOverbought 判断是否处于超买状态
// 参数：
//   - price: 当前价格
//   - threshold: 超买阈值
//
// 返回值：
//   - bool: 如果偏离度超过阈值返回true，否则返回false
//
// 说明：
//
//	当价格过度偏离均线上方时，可能出现回调
//
// 示例：
//
//	isOverbought := rma.IsOverbought(100, 10)
func (t *TaRMA) IsOverbought(price float64, threshold float64) bool {
	return t.GetDeviation(price) > threshold
}

// IsOversold 判断是否处于超卖状态
// 参数：
//   - price: 当前价格
//   - threshold: 超卖阈值
//
// 返回值：
//   - bool: 如果偏离度低于阈值返回true，否则返回false
//
// 说明：
//
//	当价格过度偏离均线下方时，可能出现反弹
//
// 示例：
//
//	isOversold := rma.IsOversold(100, 10)
func (t *TaRMA) IsOversold(price float64, threshold float64) bool {
	return t.GetDeviation(price) < -threshold
}

// GetSmoothing 获取平滑系数
// 返回值：
//   - float64: 平滑系数(1/period)
//
// 说明：
//
//	平滑系数决定了RMA对价格变化的敏感度
//
// 示例：
//
//	alpha := rma.GetSmoothing()
func (t *TaRMA) GetSmoothing() float64 {
	return 1.0 / float64(t.Period)
}
