package ta

import (
	"fmt"
	"math"
)

// TaBoll 布林带指标结构体(Bollinger Bands)
// 说明：
//
//	布林带是由John Bollinger在1980年代创建的技术分析工具
//	它由一条中间的简单移动平均线(MA)和上下两条标准差带组成，可以测量价格的相对高低和波动性
//	主要应用场景：
//	1. 判断趋势：价格突破上下轨可能预示趋势形成
//	2. 测量波动性：带宽扩大表示波动加剧，带宽收窄表示波动减弱
//	3. 支撑阻力：上下轨可作为动态支撑阻力位
type TaBoll struct {
	Upper []float64 `json:"upper"` // 上轨线
	Mid   []float64 `json:"mid"`   // 中轨线（简单移动平均线）
	Lower []float64 `json:"lower"` // 下轨线
}

// CalculateBoll 计算布林带指标
// 参数：
//   - prices: 价格序列
//   - period: 计算周期，通常为20
//   - stdDev: 标准差倍数，通常为2
//
// 返回值：
//   - *TaBoll: 布林带指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	计算步骤：
//	1. 中轨 = N日简单移动平均线(SMA)
//	2. 上轨 = 中轨 + K倍标准差
//	3. 下轨 = 中轨 - K倍标准差
//
// 示例：
//
//	boll, err := CalculateBoll(prices, 20, 2)
func CalculateBoll(prices []float64, period int, stdDev float64) (*TaBoll, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 3) // [upper, mid, lower]
	upper, mid, lower := slices[0], slices[1], slices[2]

	// 计算移动平均线（中轨）
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	mid[period-1] = sum / float64(period)

	// 使用滑动窗口计算后续的移动平均值
	for i := period; i < length; i++ {
		sum = sum - prices[i-period] + prices[i]
		mid[i] = sum / float64(period)
	}

	// 计算标准差和布林带
	for i := period - 1; i < length; i++ {
		// 计算标准差
		var sumSquares float64
		for j := 0; j < period; j++ {
			diff := prices[i-j] - mid[i]
			sumSquares += diff * diff
		}
		sd := math.Sqrt(sumSquares / float64(period))

		// 计算上下轨
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

// Boll 计算K线数据的布林带指标
// 参数：
//   - period: 计算周期，通常为20
//   - stdDev: 标准差倍数，通常为2
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaBoll: 布林带指标结构体
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	boll, err := k.Boll(20, 2, "close")
func (k *KlineDatas) Boll(period int, stdDev float64, source string) (*TaBoll, error) {
	prices, err := k._ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateBoll(prices, period, stdDev)
}

// Boll_ 获取最新的布林带值
// 参数：
//   - period: 计算周期，通常为20
//   - stdDev: 标准差倍数，通常为2
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - upper: 最新的上轨值
//   - mid: 最新的中轨值
//   - lower: 最新的下轨值
//
// 示例：
//
//	upper, mid, lower := k.Boll_(20, 2, "close")
func (k *KlineDatas) Boll_(period int, stdDev float64, source string) (upper, mid, lower float64) {
	// 只保留必要的计算数据
	_k, err := k._Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k._ExtractSlice(source)
	if err != nil {
		return 0, 0, 0
	}
	boll, err := CalculateBoll(prices, period, stdDev)
	if err != nil {
		return 0, 0, 0
	}
	return boll.Value()
}

// Value 返回最新的布林带值
// 返回值：
//   - upper: 最新的上轨值
//   - mid: 最新的中轨值
//   - lower: 最新的下轨值
//
// 示例：
//
//	upper, mid, lower := boll.Value()
func (t *TaBoll) Value() (upper, mid, lower float64) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Mid[lastIndex], t.Lower[lastIndex]
}

// IsBollCross 判断是否发生布林带交叉
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - upperCross: 是否向上突破上轨
//   - lowerCross: 是否向下突破下轨
//
// 说明：
//
//	价格突破上轨可能预示上涨趋势，突破下轨可能预示下跌趋势
//
// 示例：
//
//	upperCross, lowerCross := boll.IsBollCross(prices)
func (t *TaBoll) IsBollCross(prices []float64) (upperCross, lowerCross bool) {
	if len(prices) < 2 || len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false, false
	}
	lastIndex := len(prices) - 1
	upperCross = prices[lastIndex-1] <= t.Upper[lastIndex-1] && prices[lastIndex] > t.Upper[lastIndex]
	lowerCross = prices[lastIndex-1] >= t.Lower[lastIndex-1] && prices[lastIndex] < t.Lower[lastIndex]
	return
}

// IsSqueezing 判断是否处于带宽收缩状态
// 返回值：
//   - bool: 如果带宽在收缩返回true，否则返回false
//
// 说明：
//
//	带宽收缩通常预示着行情即将发生大的变动
//
// 示例：
//
//	isSqueezing := boll.IsSqueezing()
func (t *TaBoll) IsSqueezing() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentBandwidth < previousBandwidth
}

// IsExpanding 判断是否处于带宽扩张状态
// 返回值：
//   - bool: 如果带宽在扩张返回true，否则返回false
//
// 说明：
//
//	带宽扩张通常表示市场波动性在增加
//
// 示例：
//
//	isExpanding := boll.IsExpanding()
func (t *TaBoll) IsExpanding() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentBandwidth > previousBandwidth
}

// GetBandwidth 获取布林带带宽
// 返回值：
//   - float64: 带宽百分比值
//
// 说明：
//
//	带宽 = (上轨 - 下轨) / 中轨 * 100
//
// 示例：
//
//	bandwidth := boll.GetBandwidth()
func (t *TaBoll) GetBandwidth() float64 {
	lastIndex := len(t.Upper) - 1
	return (t.Upper[lastIndex] - t.Lower[lastIndex]) / t.Mid[lastIndex] * 100
}

// IsOverBought 判断是否处于超买状态
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - bool: 如果价格高于上轨返回true，否则返回false
//
// 示例：
//
//	isOverbought := boll.IsOverBought(price)
func (t *TaBoll) IsOverBought(price float64) bool {
	lastIndex := len(t.Upper) - 1
	return price > t.Upper[lastIndex]
}

// IsOverSold 判断是否处于超卖状态
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - bool: 如果价格低于下轨返回true，否则返回false
//
// 示例：
//
//	isOversold := boll.IsOverSold(price)
func (t *TaBoll) IsOverSold(price float64) bool {
	lastIndex := len(t.Lower) - 1
	return price < t.Lower[lastIndex]
}

// IsTrendUp 判断是否处于上升趋势
// 返回值：
//   - bool: 如果中轨线向上倾斜返回true，否则返回false
//
// 示例：
//
//	isTrendUp := boll.IsTrendUp()
func (t *TaBoll) IsTrendUp() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] > t.Mid[lastIndex-1]
}

// IsTrendDown 判断是否处于下降趋势
// 返回值：
//   - bool: 如果中轨线向下倾斜返回true，否则返回false
//
// 示例：
//
//	isTrendDown := boll.IsTrendDown()
func (t *TaBoll) IsTrendDown() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] < t.Mid[lastIndex-1]
}

// IsBreakoutPossible 判断是否可能出现突破
// 返回值：
//   - bool: 如果可能出现突破返回true，否则返回false
//
// 说明：
//
//	当带宽小于2.5%且持续收缩时，可能即将出现突破
//
// 示例：
//
//	isBreakoutPossible := boll.IsBreakoutPossible()
func (t *TaBoll) IsBreakoutPossible() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	bandwidth := t.GetBandwidth()
	previousBandwidth := (t.Upper[lastIndex-1] - t.Lower[lastIndex-1]) / t.Mid[lastIndex-1] * 100
	return bandwidth < 2.5 && bandwidth < previousBandwidth
}

// GetPercentB 计算价格在布林带中的位置（%B值）
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - float64: 价格在布林带中的百分比位置（0-100）
//
// 说明：
//
//	%B = (价格 - 下轨) / (上轨 - 下轨) * 100
//
// 示例：
//
//	percentB := boll.GetPercentB(price)
func (t *TaBoll) GetPercentB(price float64) float64 {
	lastIndex := len(t.Upper) - 1
	return (price - t.Lower[lastIndex]) / (t.Upper[lastIndex] - t.Lower[lastIndex]) * 100
}
