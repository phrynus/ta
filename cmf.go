package ta

import (
	"fmt"
)

// TaCMF 用于计算资金流量指标（Chaikin Money Flow）的结构体
// 说明：
//
//	该结构体存储了计算得到的 CMF 值以及计算时使用的周期
//
// 字段：
//   - Values: 存储计算得到的 CMF 值的切片 (float64 类型)
//   - Period: 计算 CMF 时使用的周期 (int 类型)
type TaCMF struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

// CalculateCMF 计算资金流量指标（Chaikin Money Flow）
// 参数：
//   - high: 最高价数组
//   - low: 最低价数组
//   - close: 收盘价数组
//   - volume: 成交量数组
//   - period: 计算周期
//
// 返回值：
//   - *TaCMF: 存储计算结果的 TaCMF 结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	输入的 high、low、close 和 volume 数组长度必须一致
//	输入数据长度必须大于等于计算周期
//
// 示例：
//
//	cmf, err := CalculateCMF(highPrices, lowPrices, closePrices, volumes, 20)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateCMF(high, low, close, volume []float64, period int) (*TaCMF, error) {
	if len(high) != len(low) || len(high) != len(close) || len(high) != len(volume) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(high) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(high)
	mfv := make([]float64, length)
	cmf := make([]float64, length)

	for i := 0; i < length; i++ {
		if high[i] == low[i] {
			mfv[i] = 0
		} else {
			mfm := ((close[i] - low[i]) - (high[i] - close[i])) / (high[i] - low[i])
			mfv[i] = mfm * volume[i]
		}
	}

	for i := period - 1; i < length; i++ {
		sumMFV := 0.0
		sumVolume := 0.0
		for j := 0; j < period; j++ {
			sumMFV += mfv[i-j]
			sumVolume += volume[i-j]
		}
		if sumVolume != 0 {
			cmf[i] = sumMFV / sumVolume
		}
	}

	return &TaCMF{
		Values: cmf,
		Period: period,
	}, nil
}

// CMF 从 KlineDatas 结构体中提取数据并计算资金流量指标（Chaikin Money Flow）
// 参数：
//   - period: 计算周期
//   - source: 数据源标识
//
// 返回值：
//   - *TaCMF: 存储计算结果的 TaCMF 结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法会调用 KlineDatas 的 ExtractSlice 方法提取数据
//	提取数据过程中可能会出现错误
//
// 示例：
//
//	cmf, err := klineDatas.CMF(20, "source")
//	if err != nil {
//	    // 处理错误
//	}
func (k *KlineDatas) CMF(period int, source string) (*TaCMF, error) {
	high, err := k.ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k.ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateCMF(high, low, close, volume, period)
}

// Value 获取 TaCMF 结构体中最后一个 CMF 值
// 返回值：
//   - float64: 最后一个 CMF 值
//
// 示例：
//
//	value := cmf.Value()
func (t *TaCMF) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsPositive 判断 TaCMF 结构体中最后一个 CMF 值是否为正数
// 返回值：
//   - bool: 如果最后一个 CMF 值为正数，返回 true；否则返回 false
//
// 示例：
//
//	isPositive := cmf.IsPositive()
func (t *TaCMF) IsPositive() bool {
	return t.Values[len(t.Values)-1] > 0
}

// IsNegative 判断 TaCMF 结构体中最后一个 CMF 值是否为负数
// 返回值：
//   - bool: 如果最后一个 CMF 值为负数，返回 true；否则返回 false
//
// 示例：
//
//	isNegative := cmf.IsNegative()
func (t *TaCMF) IsNegative() bool {
	return t.Values[len(t.Values)-1] < 0
}

// IsBullishDivergence 判断是否存在看涨背离
// 参数：
//   - prices: 价格数组
//
// 返回值：
//   - bool: 如果存在看涨背离，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	价格数组长度必须大于等于 20
//
// 示例：
//
//	isBullish := cmf.IsBullishDivergence(prices)
func (t *TaCMF) IsBullishDivergence(prices []float64) bool {
	if len(prices) < 20 {
		return false
	}
	lastIndex := len(prices) - 1
	previousLow := prices[lastIndex-1]
	for i := 2; i < 20; i++ {
		if prices[lastIndex-i] < previousLow {
			previousLow = prices[lastIndex-i]
		}
	}
	return prices[lastIndex] > previousLow && t.Values[lastIndex] < 0
}

// IsBearishDivergence 判断是否存在看跌背离
// 参数：
//   - prices: 价格数组
//
// 返回值：
//   - bool: 如果存在看跌背离，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	价格数组长度必须大于等于 20
//
// 示例：
//
//	isBearish := cmf.IsBearishDivergence(prices)
func (t *TaCMF) IsBearishDivergence(prices []float64) bool {
	if len(prices) < 20 {
		return false
	}
	lastIndex := len(prices) - 1
	previousHigh := prices[lastIndex-1]
	for i := 2; i < 20; i++ {
		if prices[lastIndex-i] > previousHigh {
			previousHigh = prices[lastIndex-i]
		}
	}
	return prices[lastIndex] < previousHigh && t.Values[lastIndex] > 0
}

// CMF_ 从 KlineDatas 结构体中提取数据并计算资金流量指标（Chaikin Money Flow）的简化版本
// 参数：
//   - period: 计算周期
//   - source: 数据源标识
//
// 返回值：
//   - float64: 计算得到的 CMF 值
//
// 说明/注意事项：
//
//	该方法会调用 KlineDatas 的 Keep 方法和 CMF 方法
//	可能会出现错误，错误发生时返回 0
//
// 示例：
//
//	cmfValue := klineDatas.CMF_(20, "source")
func (k *KlineDatas) CMF_(period int, source string) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	cmf, err := _k.CMF(period, source)
	if err != nil {
		return 0
	}
	return cmf.Value()
}

// GetStrength 获取 TaCMF 结构体中最后一个 CMF 值，作为强度指标
// 返回值：
//   - float64: 最后一个 CMF 值
//
// 示例：
//
//	strength := cmf.GetStrength()
func (t *TaCMF) GetStrength() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossZero 判断 TaCMF 结构体中最后一个 CMF 值是否穿过零轴
// 返回值：
//   - bool: 如果最后一个 CMF 值大于 0，返回 true；否则返回 false
//   - bool: 如果最后一个 CMF 值小于 0，返回 true；否则返回 false
//
// 示例：
//
//	isPositive, isNegative := cmf.IsCrossZero()
func (t *TaCMF) IsCrossZero() (bool, bool) {
	lastValue := t.Values[len(t.Values)-1]
	return lastValue > 0, lastValue < 0
}

// IsStrengthening 判断 TaCMF 结构体中最后一个 CMF 值是否比前一个值大
// 返回值：
//   - bool: 如果最后一个 CMF 值比前一个值大，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	结构体中 CMF 值的切片长度必须大于等于 2
//
// 示例：
//
//	isStrengthening := cmf.IsStrengthening()
func (t *TaCMF) IsStrengthening() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > t.Values[len(t.Values)-2]
}

// IsWeakening 判断 TaCMF 结构体中最后一个 CMF 值是否比前一个值小
// 返回值：
//   - bool: 如果最后一个 CMF 值比前一个值小，返回 true；否则返回 false
//
// 说明/注意事项：
//
//	结构体中 CMF 值的切片长度必须大于等于 2
//
// 示例：
//
//	isWeakening := cmf.IsWeakening()
func (t *TaCMF) IsWeakening() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] < t.Values[len(t.Values)-2]
}

// GetMoneyFlowZone 根据最后一个 CMF 值判断资金流向区域
// 返回值：
//   - string: 资金流向区域的标识，可能为 "流入", "中性", "外流"
//
// 示例：
//
//	zone := cmf.GetMoneyFlowZone()
func (t *TaCMF) GetMoneyFlowZone() string {
	lastValue := t.Values[len(t.Values)-1]
	if lastValue > 0 {
		return "流入"
	} else if lastValue > -0.1 && lastValue < 0.1 {
		return "中性"
	} else {
		return "外流"
	}
}
