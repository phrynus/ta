package ta

import (
	"fmt"
	"math"
)

// TaBoll 表示布林带指标的计算结果
// 说明：
//
//	布林带指标由上轨、中轨和下轨组成，用于衡量价格波动的范围和趋势
//
// 字段：
//   - Upper: 布林带上轨的值数组
//   - Mid: 布林带中轨的值数组
//   - Lower: 布林带下轨的值数组
type TaBoll struct {
	Upper []float64 `json:"upper"`
	Mid   []float64 `json:"mid"`
	Lower []float64 `json:"lower"`
}

// CalculateBoll 计算布林带指标
// 参数：
//   - prices: 价格数据数组
//   - period: 计算周期
//   - stdDev: 标准差倍数
//
// 返回值：
//   - *TaBoll: 布林带指标的计算结果指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	当输入的价格数据长度小于计算周期时，会返回错误
//
// 示例：
//
//	boll, err := CalculateBoll(prices, 20, 2)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateBoll(prices []float64, period int, stdDev float64) (*TaBoll, error) {

	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 3)
	upper, mid, lower := slices[0], slices[1], slices[2]

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}

	mid[period-1] = sum / float64(period)

	for i := period; i < length; i++ {

		sum = sum - prices[i-period] + prices[i]

		mid[i] = sum / float64(period)
	}

	for i := period - 1; i < length; i++ {

		var sumSquares float64

		for j := 0; j < period; j++ {
			diff := prices[i-j] - mid[i]
			sumSquares += diff * diff
		}

		sd := math.Sqrt(sumSquares / float64(period))

		band := sd * stdDev

		upper[i] = mid[i] + band

		lower[i] = mid[i] - band
	}

	return &TaBoll{
		Upper: upper,
		Mid:   mid,
		Lower: lower,
	}, nil
}

// Boll 为 KlineDatas 类型计算布林带指标
// 参数：
//   - period: 计算周期
//   - stdDev: 标准差倍数
//   - source: 价格数据来源
//
// 返回值：
//   - *TaBoll: 布林带指标的计算结果指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	会调用 ExtractSlice 方法提取价格数据，若提取失败会返回错误
func (k *KlineDatas) Boll(period int, stdDev float64, source string) (*TaBoll, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateBoll(prices, period, stdDev)
}

// Boll_ 为 KlineDatas 类型计算布林带指标并返回最后一个值
// 参数：
//   - period: 计算周期
//   - stdDev: 标准差倍数
//   - source: 价格数据来源
//
// 返回值：
//   - upper: 布林带上轨的最后一个值
//   - mid: 布林带中轨的最后一个值
//   - lower: 布林带下轨的最后一个值
//
// 说明/注意事项：
//
//	会先截取前 period * 14 个数据，若截取失败则使用全部数据
//	调用 ExtractSlice 方法提取价格数据，若提取失败会返回 0
func (k *KlineDatas) Boll_(period int, stdDev float64, source string) (upper, mid, lower float64) {

	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0, 0
	}
	boll, err := CalculateBoll(prices, period, stdDev)
	if err != nil {
		return 0, 0, 0
	}
	return boll.Value()
}

// Value 返回布林带指标的最后一个值
// 返回值：
//   - upper: 布林带上轨的最后一个值
//   - mid: 布林带中轨的最后一个值
//   - lower: 布林带下轨的最后一个值
func (t *TaBoll) Value() (upper, mid, lower float64) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Mid[lastIndex], t.Lower[lastIndex]
}

// IsBollCross 判断价格是否穿过布林带上下轨
// 参数：
//   - prices: 价格数据数组
//
// 返回值：
//   - upperCross: 价格是否穿过上轨
//   - lowerCross: 价格是否穿过下轨
//
// 说明/注意事项：
//
//	当价格数据、上轨数据或下轨数据长度小于 2 时，返回 false
func (t *TaBoll) IsBollCross(prices []float64) (upperCross, lowerCross bool) {
	if len(prices) < 2 || len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false, false
	}
	lastIndex := len(prices) - 1
	upperCross = prices[lastIndex-1] <= t.Upper[lastIndex-1] && prices[lastIndex] > t.Upper[lastIndex]
	lowerCross = prices[lastIndex-1] >= t.Lower[lastIndex-1] && prices[lastIndex] < t.Lower[lastIndex]
	return
}


// IsTrendUp 判断布林带中轨是否向上
// 返回值：
//   - bool: 布林带中轨是否向上
//
// 说明/注意事项：
//
//	当布林带中轨数据长度小于 2 时，返回 false
func (t *TaBoll) IsTrendUp() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] > t.Mid[lastIndex-1]
}

// IsTrendDown 判断布林带中轨是否向下
// 返回值：
//   - bool: 布林带中轨是否向下
//
// 说明/注意事项：
//
//	当布林带中轨数据长度小于 2 时，返回 false
func (t *TaBoll) IsTrendDown() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] < t.Mid[lastIndex-1]
}
