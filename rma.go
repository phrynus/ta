package ta

import (
	"fmt"
)

// TaRMA 相对移动平均线(Relative Moving Average)指标结果结构体
// 说明：
//
//	存储RMA计算结果及相关参数，提供多种基于RMA的技术分析方法
//
// 字段：
//   - Values: RMA计算结果数组 ([]float64)
//   - Period: 计算周期 (int)
type TaRMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

// CalculateRMA 计算相对移动平均线(RMA)
// 参数：
//   - prices: 价格数据数组 ([]float64)
//   - period: 计算周期 (int)
//
// 返回值：
//   - *TaRMA: RMA计算结果对象
//   - error: 错误信息，若输入数据不足则返回错误
//
// 说明/注意事项：
//
//	RMA是一种平滑的移动平均线，与EMA类似但计算方式略有不同
//	计算公式：RMA[i] = alpha * price[i] + (1-alpha) * RMA[i-1]，其中alpha=1/period
//	要求输入价格数据长度不小于计算周期
//
// 示例：
//
//	prices := []float64{10.0, 11.0, 12.0, 13.0, 14.0}
//	rma, err := CalculateRMA(prices, 3)
func CalculateRMA(prices []float64, period int) (*TaRMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
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

// RMA 从K线数据中提取指定源数据并计算RMA
// 参数：
//   - period: 计算周期 (int)
//   - source: 数据源字段名 (string)
//
// 返回值：
//   - *TaRMA: RMA计算结果对象
//   - error: 错误信息
//
// 说明/注意事项：
//
//	该方法是KlineDatas结构体的成员方法，用于便捷地从K线数据计算RMA
//	支持的数据源字段由ExtractSlice方法决定
func (k *KlineDatas) RMA(period int, source string) (*TaRMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRMA(prices, period)
}

// RMA_ 获取RMA的最新值
// 参数：
//   - period: 计算周期 (int)
//   - source: 数据源字段名 (string)
//
// 返回值：
//   - float64: RMA的最新值，若计算失败则返回0
//
// 说明/注意事项：
//
//	该方法会尝试保留period*14长度的数据以确保计算准确性
//	计算失败时会静默返回0，不会返回错误
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

// Value 获取RMA的最新值
// 返回值：
//   - float64: RMA数组中的最后一个值
func (t *TaRMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossOver 判断价格是否上穿RMA
// 参数：
//   - price: 当前价格 (float64)
//
// 返回值：
//   - bool: 若上穿则返回true，否则返回false
//
// 说明/注意事项：
//
//	上穿定义：前一期RMA值 <= 价格 且 当前RMA值 > 价格
//	需要至少2个RMA值才能进行判断
func (t *TaRMA) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

// IsCrossUnder 判断价格是否下穿RMA
// 参数：
//   - price: 当前价格 (float64)
//
// 返回值：
//   - bool: 若下穿则返回true，否则返回false
//
// 说明/注意事项：
//
//	下穿定义：前一期RMA值 >= 价格 且 当前RMA值 < 价格
//	需要至少2个RMA值才能进行判断
func (t *TaRMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

// GetSlope 获取RMA的斜率（变化率）
// 返回值：
//   - float64: 当前RMA值与前一期的差值
//
// 说明/注意事项：
//
//	斜率为正表示上升，为负表示下降
//	需要至少2个RMA值才能计算
func (t *TaRMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

// IsCrossOverRMA 判断当前RMA是否与另一个RMA交叉
// 参数：
//   - other: 另一个RMA对象 (*TaRMA)
//
// 返回值：
//   - golden: 是否形成金叉 (bool)
//   - death: 是否形成死叉 (bool)
//
// 说明/注意事项：
//
//	金叉定义：当前RMA从前一期小于等于other变为当前大于other
//	死叉定义：当前RMA从前一期大于等于other变为当前小于other
//	需要两个RMA对象都至少有2个值才能判断
func (t *TaRMA) IsCrossOverRMA(other *TaRMA) (golden, death bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	golden = t.Values[lastIndex-1] <= other.Values[lastIndex-1] && t.Values[lastIndex] > other.Values[lastIndex]
	death = t.Values[lastIndex-1] >= other.Values[lastIndex-1] && t.Values[lastIndex] < other.Values[lastIndex]
	return
}
