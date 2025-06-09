package ta

import (
	"fmt"
)

// TaT3 T3移动平均线指标结构体(Tillson T3 Moving Average)
// 说明：
//
//	T3是由Tim Tillson开发的一种自适应移动平均线指标
//	它通过多重EMA计算和体积因子(Volume Factor)来减少滞后性
//	主要应用场景：
//	1. 趋势跟踪：识别中长期趋势方向
//	2. 动态支撑阻力：提供移动支撑和阻力位
//	3. 交叉信号：与价格或其他均线的交叉产生交易信号
//
// 计算公式：
//  1. 首先计算第一重EMA
//  2. 根据体积因子(Volume Factor)计算多重EMA
//  3. 最终T3 = c1*e6 + c2*e5 + c3*e4 + c4*e3
//     其中c1,c2,c3,c4是基于体积因子的常数
type TaT3 struct {
	Values []float64 `json:"values"` // T3值序列
	Period int       `json:"period"` // 计算周期
	VFact  float64   `json:"vfact"`  // 体积因子(Volume Factor)
}

// CalculateT3 计算T3移动平均线
// 参数：
//   - prices: 价格序列
//   - period: 计算周期
//   - vfact: 体积因子，通常在0到1之间，默认为0.7
//
// 返回值：
//   - *TaT3: T3指标结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	T3移动平均线结合了EMA的敏感性和多重平滑的稳定性
//	计算步骤：
//	1. 计算初始EMA
//	2. 根据体积因子计算权重
//	3. 进行多重EMA计算
//	4. 合成最终的T3值
//
// 示例：
//
//	t3, err := CalculateT3(prices, 10, 0.7)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前T3值：%.2f\n", t3.Values[len(t3.Values)-1])
func CalculateT3(prices []float64, period int, vfact float64) (*TaT3, error) {
	if len(prices) < period*6 {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 7) // [ema1, ema2, ema3, ema4, ema5, ema6, t3]
	ema1, ema2, ema3, ema4, ema5, ema6, t3 := slices[0], slices[1], slices[2], slices[3], slices[4], slices[5], slices[6]

	// 计算第一个EMA
	k := 2.0 / float64(period+1)
	ema1[0] = prices[0]
	for i := 1; i < length; i++ {
		ema1[i] = prices[i]*k + ema1[i-1]*(1-k)
	}

	// 计算后续的EMA
	for i := 1; i < length; i++ {
		ema2[i] = ema1[i]*k + ema2[i-1]*(1-k)
		ema3[i] = ema2[i]*k + ema3[i-1]*(1-k)
		ema4[i] = ema3[i]*k + ema4[i-1]*(1-k)
		ema5[i] = ema4[i]*k + ema5[i-1]*(1-k)
		ema6[i] = ema5[i]*k + ema6[i-1]*(1-k)
	}

	// T3计算常数
	b := vfact
	c1 := -b * b * b
	c2 := 3*b*b + 3*b*b*b
	c3 := -6*b*b - 3*b - 3*b*b*b
	c4 := 1 + 3*b + b*b*b + 3*b*b

	// 计算T3值
	for i := period * 6; i < length; i++ {
		t3[i] = c1*ema6[i] + c2*ema5[i] + c3*ema4[i] + c4*ema3[i]
	}

	return &TaT3{
		Values: t3,
		Period: period,
		VFact:  vfact,
	}, nil
}

// T3 计算K线数据的T3移动平均线
// 参数：
//   - period: 计算周期
//   - vfact: 体积因子，通常在0到1之间，默认为0.7
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaT3: T3指标结构体指针
//   - error: 可能的错误信息
//
// 说明：
//
//	T3适用于需要减少滞后性的趋势跟踪
//	特点：
//	1. 对价格变化反应较快
//	2. 具有良好的平滑效果
//	3. 可通过体积因子调整灵敏度
//
// 示例：
//
//	t3, err := kline.T3(10, 0.7, "close")
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) T3(period int, vfact float64, source string) (*TaT3, error) {
	prices, err := k._ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateT3(prices, period, vfact)
}

// T3_ 获取最新的T3值
// 参数：
//   - period: 计算周期
//   - vfact: 体积因子
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的T3值
//
// 示例：
//
//	value := kline.T3_(10, 0.7, "close")
func (k *KlineDatas) T3_(period int, vfact float64, source string) float64 {
	_k, err := k._Keep(period * 10)
	if err != nil {
		_k = *k
	}
	prices, err := _k._ExtractSlice(source)
	if err != nil {
		return 0
	}
	t3, err := CalculateT3(prices, period, vfact)
	if err != nil {
		return 0
	}
	return t3.Value()
}

// Value 返回最新的T3值
// 返回值：
//   - float64: 最新的T3值
//
// 示例：
//
//	value := t3.Value()
func (t *TaT3) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossOver 判断是否向上穿越指定价格
// 参数：
//   - price: 目标价格
//
// 返回值：
//   - bool: 如果T3从下向上穿越价格返回true，否则返回false
//
// 说明：
//
//	向上穿越通常是买入信号
//
// 示例：
//
//	isCrossOver := t3.IsCrossOver(100)
func (t *TaT3) IsCrossOver(price float64) bool {
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
//   - bool: 如果T3从上向下穿越价格返回true，否则返回false
//
// 说明：
//
//	向下穿越通常是卖出信号
//
// 示例：
//
//	isCrossUnder := t3.IsCrossUnder(100)
func (t *TaT3) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

// IsTrendUp 判断是否处于上升趋势
// 返回值：
//   - bool: 如果T3值在上升返回true，否则返回false
//
// 说明：
//
//	通过比较当前值和前一个值来判断趋势方向
//
// 示例：
//
//	isUp := t3.IsTrendUp()
func (t *TaT3) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

// IsTrendDown 判断是否处于下降趋势
// 返回值：
//   - bool: 如果T3值在下降返回true，否则返回false
//
// 说明：
//
//	通过比较当前值和前一个值来判断趋势方向
//
// 示例：
//
//	isDown := t3.IsTrendDown()
func (t *TaT3) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

// GetSlope 获取T3斜率
// 返回值：
//   - float64: T3值的变化率
//
// 说明：
//
//	斜率反映了趋势的强度，正值表示上升趋势，负值表示下降趋势
//
// 示例：
//
//	slope := t3.GetSlope()
func (t *TaT3) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

// IsCrossOverT3 判断是否与另一个T3发生交叉
// 参数：
//   - other: 另一个T3指标
//
// 返回值：
//   - golden: 金叉信号(买入)
//   - death: 死叉信号(卖出)
//
// 说明：
//
//	两条T3线的交叉通常用于判断趋势的转换
//
// 示例：
//
//	golden, death := t3.IsCrossOverT3(otherT3)
func (t *TaT3) IsCrossOverT3(other *TaT3) (bool, bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastT3 := t.Values[len(t.Values)-1]
	prevT3 := t.Values[len(t.Values)-2]
	lastOther := other.Values[len(other.Values)-1]
	prevOther := other.Values[len(other.Values)-2]

	return lastT3 > lastOther && prevT3 <= prevOther, lastT3 < lastOther && prevT3 >= prevOther
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
//	isAccelerating := t3.IsAccelerating()
func (t *TaT3) IsAccelerating() bool {
	if len(t.Values) < 4 {
		return false
	}
	lastIndex := len(t.Values) - 1
	diff1 := t.Values[lastIndex] - t.Values[lastIndex-1]
	diff2 := t.Values[lastIndex-1] - t.Values[lastIndex-2]
	diff3 := t.Values[lastIndex-2] - t.Values[lastIndex-3]
	return (diff1 > diff2 && diff2 > diff3) || (diff1 < diff2 && diff2 < diff3)
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
//	strength := t3.GetTrendStrength()
func (t *TaT3) GetTrendStrength() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	diff := t.Values[lastIndex] - t.Values[0]
	return (diff / t.Values[0]) * 100
}

// GetDeviation 获取当前价格与T3的偏离度
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
//	deviation := t3.GetDeviation(100)
func (t *TaT3) GetDeviation(price float64) float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastValue := t.Values[len(t.Values)-1]
	diff := lastValue - price
	return (diff / price) * 100
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
//	isOverbought := t3.IsOverbought(100, 10)
func (t *TaT3) IsOverbought(price float64, threshold float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	deviation := t.GetDeviation(price)
	return deviation > threshold
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
//	isOversold := t3.IsOversold(100, 10)
func (t *TaT3) IsOversold(price float64, threshold float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	deviation := t.GetDeviation(price)
	return deviation < -threshold
}

// GetVolumeFactorEffect 获取体积因子的影响程度
// 返回值：
//   - float64: 体积因子对T3值的影响程度
//
// 说明：
//
//	体积因子影响T3对价格变化的敏感度
//
// 示例：
//
//	effect := t3.GetVolumeFactorEffect()
func (t *TaT3) GetVolumeFactorEffect() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	// 使用最新值和初始值计算影响程度
	return t.VFact * 100
}

// GetOptimalPeriod 获取最优计算周期
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - int: 建议的最优计算周期
//
// 说明：
//
//	根据价格波动性自动计算最优的T3周期
//
// 示例：
//
//	optPeriod := t3.GetOptimalPeriod(prices)
func (t *TaT3) GetOptimalPeriod(prices []float64) int {
	if len(prices) < 2 {
		return 0
	}
	bestPeriod := 0
	bestSlope := 0.0
	for period := 5; period <= 20; period++ {
		t3, _ := CalculateT3(prices, period, t.VFact)
		slope := t3.GetSlope()
		if slope > bestSlope {
			bestSlope = slope
			bestPeriod = period
		}
	}
	return bestPeriod
}

// GetOptimalVolumeFactor 获取最优体积因子
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - float64: 建议的最优体积因子
//
// 说明：
//
//	根据价格波动性自动计算最优的体积因子
//
// 示例：
//
//	optVFact := t3.GetOptimalVolumeFactor(prices)
func (t *TaT3) GetOptimalVolumeFactor(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}
	bestVFact := 0.0
	bestSlope := 0.0
	for vfact := 0.1; vfact <= 1.0; vfact += 0.1 {
		t3, _ := CalculateT3(prices, t.Period, vfact)
		slope := t3.GetSlope()
		if slope > bestSlope {
			bestSlope = slope
			bestVFact = vfact
		}
	}
	return bestVFact
}
