package ta

import (
	"fmt"
	"math"
)

// TaCCI 商品通道指标结构体(Commodity Channel Index)
type TaCCI struct {
	Values []float64 `json:"values"` // CCI值序列
}

// CalculateCCI 计算商品通道指标(Commodity Channel Index)
// 参数：
//   - klineData: K线数据集合
//   - period: 计算周期，通常为20
//
// 返回值：
//   - *TaCCI: CCI值序列
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	CCI是衡量价格是否超出其统计常态分布范围的技术指标
//	计算步骤：
//	1. 计算典型价格TP = (High + Low + Close) / 3
//	2. 计算TP的period日简单移动平均线SMA
//	3. 计算TP与其移动平均线的偏差MD
//	4. 计算偏差的period日简单移动平均线
//	5. 计算CCI = (TP - SMA) / (0.015 * MD)
//
// 示例：
//
//	cci, err := CalculateCCI(klineData, 20)
func CalculateCCI(klineData KlineDatas, period int) (*TaCCI, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 2) // [typicalPrice, cci]
	typicalPrice, cci := slices[0], slices[1]

	// 计算典型价格(TP)
	for i := 0; i < length; i++ {
		typicalPrice[i] = (klineData[i].High + klineData[i].Low + klineData[i].Close) / 3
	}

	// 使用滑动窗口计算CCI
	for i := period - 1; i < length; i++ {
		// 计算当前窗口的TP平均值
		var sumTP float64
		for j := i - period + 1; j <= i; j++ {
			sumTP += typicalPrice[j]
		}
		smaTP := sumTP / float64(period)

		// 计算平均绝对偏差
		var sumAbsDev float64
		for j := i - period + 1; j <= i; j++ {
			sumAbsDev += math.Abs(typicalPrice[j] - smaTP)
		}
		meanDeviation := sumAbsDev / float64(period)

		// 计算CCI值
		if meanDeviation != 0 {
			cci[i] = (typicalPrice[i] - smaTP) / (0.015 * meanDeviation)
		}
	}

	return &TaCCI{
		Values: cci,
	}, nil
}

// CCI 计算K线数据的顺势指标
// 参数：
//   - period: 计算周期，通常为14或20日
//
// 返回值：
//   - *TaCCI: CCI值序列
//   - error: 可能的错误信息
func (k *KlineDatas) CCI(period int) (*TaCCI, error) {
	return CalculateCCI(*k, period)
}

// CCI_ 获取最新的CCI值
// 参数：
//   - period: 计算周期，通常为14或20日
//
// 返回值：
//   - float64: 最新的CCI值，如果计算失败则返回0
func (k *KlineDatas) CCI_(period int) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	cci, err := CalculateCCI(_k, period)
	if err != nil {
		return 0
	}
	return cci.Value()
}

// Value 返回最新的CCI值
func (t *TaCCI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsOverbought 判断是否处于超买状态
// CCI > 100表示超买
func (t *TaCCI) IsOverbought() bool {
	return t.Value() > 100
}

// IsOversold 判断是否处于超卖状态
// CCI < -100表示超卖
func (t *TaCCI) IsOversold() bool {
	return t.Value() < -100
}

// IsExtremeBought 判断是否处于极度超买状态
// CCI > 200表示极度超买
func (t *TaCCI) IsExtremeBought() bool {
	return t.Value() > 200
}

// IsExtremeSold 判断是否处于极度超卖状态
// CCI < -200表示极度超卖
func (t *TaCCI) IsExtremeSold() bool {
	return t.Value() < -200
}

// IsBullishDivergence 判断是否出现多头背离
// 当价格创新低而CCI未创新低时出现多头背离
func (t *TaCCI) IsBullishDivergence() bool {
	if len(t.Values) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	previousLow := t.Values[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.Values[lastIndex-i] < previousLow {
			previousLow = t.Values[lastIndex-i]
		}
	}
	return t.Values[lastIndex] > previousLow && t.Values[lastIndex] < -100
}

// IsBearishDivergence 判断是否出现空头背离
// 当价格创新高而CCI未创新高时出现空头背离
func (t *TaCCI) IsBearishDivergence() bool {
	if len(t.Values) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	previousHigh := t.Values[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.Values[lastIndex-i] > previousHigh {
			previousHigh = t.Values[lastIndex-i]
		}
	}
	return t.Values[lastIndex] < previousHigh && t.Values[lastIndex] > 100
}
