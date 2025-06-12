package ta

import (
	"fmt"
	"math"
)

// TaBoll 表示布林带指标的计算结果
// 说明：
//
//	布林带指标由上轨、中轨和下轨组成，可用于分析价格波动和趋势。
//
// 字段：
//   - Upper: 布林带上轨数值切片
//   - Mid: 布林带中轨数值切片
//   - Lower: 布林带下轨数值切片
type TaBoll struct {
	Upper []float64 `json:"upper"`
	Mid   []float64 `json:"mid"`
	Lower []float64 `json:"lower"`
}

// CalculateBoll 计算布林带指标，包括上轨、中轨和下轨。
// 参数：
//   - prices: 价格数据切片，包含用于计算布林带的历史价格数据。
//   - period: 计算周期，用于计算移动平均线和标准差的窗口大小。
//   - stdDev: 标准差倍数，用于确定上下轨与中轨的距离。
//
// 返回值：
//   - *TaBoll: 包含布林带上轨、中轨和下轨数值切片的结构体指针。
//   - error: 若传入的价格数据长度小于计算周期，返回错误信息。
//
// 说明/注意事项：
//
//	计算所需的价格数据长度必须大于等于指定的周期，否则会返回错误。
//
// 示例：
//
//	boll, err := CalculateBoll(prices, 20, 2.0)
//	if err != nil {
//	    log.Fatal(err)
//	}
func CalculateBoll(prices []float64, period int, stdDev float64) (*TaBoll, error) {
	// 检查价格数据长度是否足够，如果不足则返回错误
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 获取价格数据的长度
	length := len(prices)

	// 预先分配三个切片，分别用于存储上轨、中轨和下轨的数据
	slices := preallocateSlices(length, 3)
	upper, mid, lower := slices[0], slices[1], slices[2]

	// 计算初始周期内价格数据的总和
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	// 计算初始周期的移动平均线，作为中轨的第一个有效数据
	mid[period-1] = sum / float64(period)

	// 从周期结束位置开始，通过滑动窗口计算后续的移动平均线
	for i := period; i < length; i++ {
		// 减去窗口最旧的数据，加上新的数据，更新总和
		sum = sum - prices[i-period] + prices[i]
		// 计算当前窗口的移动平均线
		mid[i] = sum / float64(period)
	}

	// 从周期结束位置开始，计算每个位置的上轨和下轨
	for i := period - 1; i < length; i++ {
		// 初始化平方差总和
		var sumSquares float64
		// 计算当前周期内每个价格与中轨的平方差之和
		for j := 0; j < period; j++ {
			diff := prices[i-j] - mid[i]
			sumSquares += diff * diff
		}
		// 计算当前周期内价格的标准差
		sd := math.Sqrt(sumSquares / float64(period))

		// 计算带宽，即标准差乘以标准差倍数
		band := sd * stdDev
		// 计算上轨，中轨加上带宽
		upper[i] = mid[i] + band
		// 计算下轨，中轨减去带宽
		lower[i] = mid[i] - band
	}

	// 返回包含上轨、中轨和下轨数据的 TaBoll 结构体指针
	return &TaBoll{
		Upper: upper,
		Mid:   mid,
		Lower: lower,
	}, nil
}

func (k *KlineDatas) Boll(period int, stdDev float64, source string) (*TaBoll, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateBoll(prices, period, stdDev)
}

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

func (t *TaBoll) Value() (upper, mid, lower float64) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Mid[lastIndex], t.Lower[lastIndex]
}

func (t *TaBoll) IsBollCross(prices []float64) (upperCross, lowerCross bool) {
	if len(prices) < 2 || len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false, false
	}
	lastIndex := len(prices) - 1
	upperCross = prices[lastIndex-1] <= t.Upper[lastIndex-1] && prices[lastIndex] > t.Upper[lastIndex]
	lowerCross = prices[lastIndex-1] >= t.Lower[lastIndex-1] && prices[lastIndex] < t.Lower[lastIndex]
	return
}

func (t *TaBoll) IsSqueezing() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentBandwidth < previousBandwidth
}

func (t *TaBoll) IsExpanding() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentBandwidth > previousBandwidth
}

func (t *TaBoll) GetBandwidth() float64 {
	lastIndex := len(t.Upper) - 1
	return (t.Upper[lastIndex] - t.Lower[lastIndex]) / t.Mid[lastIndex] * 100
}

func (t *TaBoll) IsOverBought(price float64) bool {
	lastIndex := len(t.Upper) - 1
	return price > t.Upper[lastIndex]
}

func (t *TaBoll) IsOverSold(price float64) bool {
	lastIndex := len(t.Lower) - 1
	return price < t.Lower[lastIndex]
}

func (t *TaBoll) IsTrendUp() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] > t.Mid[lastIndex-1]
}

func (t *TaBoll) IsTrendDown() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] < t.Mid[lastIndex-1]
}

func (t *TaBoll) IsBreakoutPossible() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	bandwidth := t.GetBandwidth()
	previousBandwidth := (t.Upper[lastIndex-1] - t.Lower[lastIndex-1]) / t.Mid[lastIndex-1] * 100
	return bandwidth < 2.5 && bandwidth < previousBandwidth
}

func (t *TaBoll) GetPercentB(price float64) float64 {
	lastIndex := len(t.Upper) - 1
	return (price - t.Lower[lastIndex]) / (t.Upper[lastIndex] - t.Lower[lastIndex]) * 100
}
