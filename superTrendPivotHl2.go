package ta

import (
	"fmt"
)

// TaSuperTrendPivotHl2 超级趋势指标结构体
type TaSuperTrendPivotHl2 struct {
	Values     []float64 `json:"values"`     // 超级趋势值序列
	Direction  []int     `json:"direction"`  // 趋势方向 1: 上涨, -1: 下跌, 0: 不确定
	UpperBand  []float64 `json:"upper_band"` // 上轨
	LowerBand  []float64 `json:"lower_band"` // 下轨
	Period     int       `json:"period"`     // ATR周期
	Multiplier float64   `json:"multiplier"` // ATR乘数
}

// CalculateSuperTrendPivotHl2 计算超级趋势指标
// 参数：
//   - klineData: K线数据集合
//   - period: ATR计算周期
//   - multiplier: ATR乘数
//
// 返回值：
//   - *TaSuperTrendPivotHl2: 超级趋势指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	超级趋势指标是一个趋势跟踪指标，结合了ATR和中轴价格
//	计算步骤：
//	1. 计算中轴价格 HL2 = (High + Low) / 2
//	2. 计算ATR
//	3. 计算动态通道边界
//	4. 根据价格突破情况确定趋势方向
//
// 示例：
//
//	st, err := CalculateSuperTrendPivotHl2(klineData, 14, 3.0)
func CalculateSuperTrendPivotHl2(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	length := len(klineData)
	if length < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 计算ATR
	atr, err := CalculateATR(klineData, period)
	if err != nil {
		return nil, err
	}

	// 预分配切片
	slices := preallocateSlices(length, 4) // [values, direction, upperBand, lowerBand]
	values, direction, upperBand, lowerBand := slices[0], make([]int, length), slices[2], slices[3]

	// 计算基础数据
	for i := 0; i < length; i++ {
		// 计算中轴价格 HL2
		hl2 := (klineData[i].High + klineData[i].Low) / 2

		if i < period {
			// 初始化阶段
			upperBand[i] = hl2 + multiplier*atr.Values[i]
			lowerBand[i] = hl2 - multiplier*atr.Values[i]
			direction[i] = 0
			values[i] = hl2
			continue
		}

		// 计算动态通道边界
		basicUpperBand := hl2 + multiplier*atr.Values[i]
		basicLowerBand := hl2 - multiplier*atr.Values[i]

		// 更新通道边界
		// LowerBand只上升不下降，直到收盘价下破LowerBand
		if basicLowerBand > lowerBand[i-1] || klineData[i-1].Close < lowerBand[i-1] {
			lowerBand[i] = basicLowerBand
		} else {
			lowerBand[i] = lowerBand[i-1]
		}

		// UpperBand只下降不上升，直到收盘价上破UpperBand
		if basicUpperBand < upperBand[i-1] || klineData[i-1].Close > upperBand[i-1] {
			upperBand[i] = basicUpperBand
		} else {
			upperBand[i] = upperBand[i-1]
		}

		// 确定趋势方向
		if direction[i-1] <= 0 { // 之前是空头或不确定
			if klineData[i].Close > upperBand[i] {
				direction[i] = 1 // 转为多头
			} else {
				direction[i] = -1 // 保持空头
			}
		} else { // 之前是多头
			if klineData[i].Close < lowerBand[i] {
				direction[i] = -1 // 转为空头
			} else {
				direction[i] = 1 // 保持多头
			}
		}

		// 根据趋势方向确定超级趋势值
		if direction[i] == 1 {
			values[i] = lowerBand[i]
		} else {
			values[i] = upperBand[i]
		}
	}

	return &TaSuperTrendPivotHl2{
		Values:     values,
		Direction:  direction,
		UpperBand:  upperBand,
		LowerBand:  lowerBand,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

// Value 返回最新的超级趋势值
// 返回值：
//   - float64: 最新的超级趋势值
//
// 示例：
//
//	value := st.Value()
func (t *TaSuperTrendPivotHl2) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// GetDirection 返回最新的趋势方向
// 返回值：
//   - int: 趋势方向（1: 上涨, -1: 下跌, 0: 不确定）
//
// 示例：
//
//	direction := st.GetDirection()
func (t *TaSuperTrendPivotHl2) GetDirection() int {
	return t.Direction[len(t.Direction)-1]
}

// GetBands 返回最新的通道边界值
// 返回值：
//   - upper: 上轨值
//   - lower: 下轨值
//
// 示例：
//
//	upper, lower := st.GetBands()
func (t *TaSuperTrendPivotHl2) GetBands() (upper, lower float64) {
	lastIndex := len(t.Values) - 1
	return t.UpperBand[lastIndex], t.LowerBand[lastIndex]
}

// SuperTrendPivotHl2 计算K线数据的超级趋势指标
// 参数：
//   - period: ATR计算周期
//   - multiplier: ATR乘数
//
// 返回值：
//   - *TaSuperTrendPivotHl2: 超级趋势指标结构体
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	st, err := k.SuperTrendPivotHl2(14, 3.0)
func (k *KlineDatas) SuperTrendPivotHl2(period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	return CalculateSuperTrendPivotHl2(*k, period, multiplier)
}

// SuperTrendPivotHl2_ 计算最新的超级趋势值
// 参数：
//   - period: ATR计算周期
//   - multiplier: ATR乘数
//
// 返回值：
//   - float64: 最新的超级趋势值
//
// 示例：
//
//	value := k.SuperTrendPivotHl2_(14, 3.0)
func (k *KlineDatas) SuperTrendPivotHl2_(period int, multiplier float64) float64 {
	_k, err := k.Keep(period * 2)
	if err != nil {
		_k = *k
	}
	st, err := CalculateSuperTrendPivotHl2(_k, period, multiplier)
	if err != nil {
		return 0
	}
	return st.Value()
}
