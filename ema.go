package ta

import (
	"fmt"
)

// TaEMA 指数移动平均线（EMA）计算结果的结构体
// 说明：
//
//	该结构体用于存储 EMA 计算的结果，包含计算得到的 EMA 值数组和计算使用的周期数。
//
// 字段：
//   - Values: 存储 EMA 计算结果的浮点数数组 (float64 类型)
//   - Period: 计算 EMA 时使用的周期数 (int 类型)
type TaEMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

// CalculateEMA 计算指数移动平均线（EMA）
// 参数：
//   - prices: 输入的价格数据数组 (float64 类型)
//   - period: 计算 EMA 时使用的周期数 (int 类型)
//
// 返回值：
//   - *TaEMA: 存储 EMA 计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	当输入的价格数据长度小于指定的周期数时，会返回错误。
//	该函数使用标准的 EMA 计算公式进行计算。
//
// 示例：
//
//	prices := []float64{10, 11, 12, 13, 14}
//	period := 3
//	ema, err := CalculateEMA(prices, period)
func CalculateEMA(prices []float64, period int) (*TaEMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	result := slices[0]

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	multiplier := 2.0 / float64(period+1)
	oneMinusMultiplier := 1.0 - multiplier

	for i := period; i < length; i++ {
		result[i] = prices[i]*multiplier + result[i-1]*oneMinusMultiplier
	}

	return &TaEMA{
		Values: result,
		Period: period,
	}, nil
}

// EMA 从 KlineDatas 中提取数据并计算 EMA
// 参数：
//   - period: 计算 EMA 时使用的周期数 (int 类型)
//   - source: 提取数据的源字段名 (string 类型)
//
// 返回值：
//   - *TaEMA: 存储 EMA 计算结果的结构体指针
//   - error: 提取数据或计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法依赖 KlineDatas 的 ExtractSlice 方法提取数据。
//
// 示例：
//
//	klineData := KlineDatas{...}
//	period := 3
//	source := "close"
//	ema, err := klineData.EMA(period, source)
func (k *KlineDatas) EMA(period int, source string) (*TaEMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateEMA(prices, period)
}

// EMA_ 从 KlineDatas 中提取数据并计算 EMA 的最后一个值
// 参数：
//   - period: 计算 EMA 时使用的周期数 (int 类型)
//   - source: 提取数据的源字段名 (string 类型)
//
// 返回值：
//   - float64: EMA 计算结果的最后一个值
//
// 说明/注意事项：
//
//	该方法会保留最近的 period * 14 条数据进行计算。
//	若提取数据或计算过程中出现错误，会返回 0。
//
// 示例：
//
//	klineData := KlineDatas{...}
//	period := 3
//	source := "close"
//	emaValue := klineData.EMA_(period, source)
func (k *KlineDatas) EMA_(period int, source string) float64 {
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

// Value 获取 TaEMA 结构体中最后一个 EMA 值
// 返回值：
//   - float64: TaEMA 结构体中最后一个 EMA 值
//
// 说明/注意事项：
//
//	若 TaEMA 结构体的 Values 数组为空，可能会导致越界错误。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{10, 11, 12}, Period: 3}
//	value := ema.Value()
func (t *TaEMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossOver 判断 EMA 是否上穿给定价格
// 参数：
//   - price: 用于比较的价格 (float64 类型)
//
// 返回值：
//   - bool: 如果 EMA 上穿给定价格，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	当 EMA 的 Values 数组长度小于 2 时，直接返回 false。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{10, 11}, Period: 2}
//	price := 10.5
//	isCrossOver := ema.IsCrossOver(price)
func (t *TaEMA) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

// IsCrossUnder 判断 EMA 是否下穿给定价格
// 参数：
//   - price: 用于比较的价格 (float64 类型)
//
// 返回值：
//   - bool: 如果 EMA 下穿给定价格，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	当 EMA 的 Values 数组长度小于 2 时，直接返回 false。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{11, 10}, Period: 2}
//	price := 10.5
//	isCrossUnder := ema.IsCrossUnder(price)
func (t *TaEMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

// IsTrendUp 判断 EMA 是否处于上升趋势
// 返回值：
//   - bool: 如果 EMA 处于上升趋势，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	当 EMA 的 Values 数组长度小于 2 时，直接返回 false。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{10, 11}, Period: 2}
//	isTrendUp := ema.IsTrendUp()
func (t *TaEMA) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

// IsTrendDown 判断 EMA 是否处于下降趋势
// 返回值：
//   - bool: 如果 EMA 处于下降趋势，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	当 EMA 的 Values 数组长度小于 2 时，直接返回 false。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{11, 10}, Period: 2}
//	isTrendDown := ema.IsTrendDown()
func (t *TaEMA) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

// GetSlope 计算 EMA 的斜率
// 返回值：
//   - float64: EMA 的斜率
//
// 说明/注意事项：
//
//	当 EMA 的 Values 数组长度小于 2 时，返回 0。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{10, 11}, Period: 2}
//	slope := ema.GetSlope()
func (t *TaEMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

// IsCrossOverEMA 判断当前 EMA 是否与另一个 EMA 发生金叉或死叉
// 参数：
//   - other: 另一个 TaEMA 结构体指针
//
// 返回值：
//   - golden: 是否发生金叉 (bool 类型)
//   - death: 是否发生死叉 (bool 类型)
//
// 说明/注意事项：
//
//	当当前 EMA 或另一个 EMA 的 Values 数组长度小于 2 时，直接返回 false, false。
//
// 示例：
//
//	ema1 := &TaEMA{Values: []float64{10, 11}, Period: 2}
//	ema2 := &TaEMA{Values: []float64{10.5, 10.8}, Period: 2}
//	golden, death := ema1.IsCrossOverEMA(ema2)
func (t *TaEMA) IsCrossOverEMA(other *TaEMA) (golden, death bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	golden = t.Values[lastIndex-1] <= other.Values[lastIndex-1] && t.Values[lastIndex] > other.Values[lastIndex]
	death = t.Values[lastIndex-1] >= other.Values[lastIndex-1] && t.Values[lastIndex] < other.Values[lastIndex]
	return
}

// GetTrendStrength 计算 EMA 的趋势强度
// 返回值：
//   - float64: EMA 的趋势强度百分比
//
// 说明/注意事项：
//
//	当 EMA 的 Values 数组长度小于 Period 时，返回 0。
//
// 示例：
//
//	ema := &TaEMA{Values: []float64{10, 11, 12}, Period: 3}
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
