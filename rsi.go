package ta

import (
	"fmt"
	"math"
)

// TaRSI RSI指标结构体(Relative Strength Index)
type TaRSI struct {
	Values []float64 `json:"values"` // RSI值序列
	Period int       `json:"period"` // 计算周期
	Gains  []float64 `json:"gains"`  // 上涨幅度序列
	Losses []float64 `json:"losses"` // 下跌幅度序列
}

// CalculateRSI 计算相对强弱指标(Relative Strength Index)
// 参数：
//   - prices: 价格序列
//   - period: 计算周期
//
// 返回值：
//   - *TaRSI: RSI指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	RSI是一种动量指标，用于衡量价格变动的强度
//	计算步骤：
//	1. 计算每日价格变动的涨跌幅
//	2. 分别计算上涨和下跌的平均值
//	3. 计算相对强度(RS) = 平均上涨点数 / 平均下跌点数
//	4. 计算RSI = 100 - (100 / (1 + RS))
//
// 示例：
//
//	rsi, err := CalculateRSI(prices, 14)
func CalculateRSI(prices []float64, period int) (*TaRSI, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 3) // [rsi, gains, losses]
	rsi, gains, losses := slices[0], slices[1], slices[2]

	// 计算每日的涨跌
	for i := 1; i < length; i++ {
		change := prices[i] - prices[i-1]
		gains[i] = math.Max(0, change)
		losses[i] = math.Max(0, -change)
	}

	// 计算初始平均涨幅和跌幅
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// 计算RSI
	for i := period; i < length; i++ {
		if i > period {
			avgGain = (avgGain*(float64(period)-1) + gains[i]) / float64(period)
			avgLoss = (avgLoss*(float64(period)-1) + losses[i]) / float64(period)
		}

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return &TaRSI{
		Values: rsi,
		Period: period,
		Gains:  gains,
		Losses: losses,
	}, nil
}

// RSI 计算指定周期的相对强弱指标(RSI)序列
// 参数：
//   - period: RSI计算周期，常用值为14日
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaRSI: RSI指标结构体
//   - error: 可能的错误信息
//
// 示例：
//
//	rsi, err := k.RSI(14, "close")
func (k *KlineDatas) RSI(period int, source string) (*TaRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRSI(prices, period)
}

// RSI_ 获取最新的RSI值
// 参数：
//   - period: RSI计算周期，常用值为14日
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的RSI值，取值范围0-100，如果计算失败则返回0
//
// 示例：
//
//	value := k.RSI_(14, "close")
func (k *KlineDatas) RSI_(period int, source string) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	rsi, err := CalculateRSI(prices, period)
	if err != nil {
		return 0
	}
	return rsi.Value()
}

// Value 返回最新的RSI值
// 返回值：
//   - float64: 最新的RSI值
//
// 示例：
//
//	value := rsi.Value()
func (t *TaRSI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsOverbought 判断是否处于超买状态
// 参数：
//   - threshold: 超买阈值，可选参数，默认为70
//
// 返回值：
//   - bool: 如果RSI值超过阈值返回true，否则返回false
//
// 示例：
//
//	isOverbought := rsi.IsOverbought(80)
func (t *TaRSI) IsOverbought(threshold ...float64) bool {
	th := 70.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() > th
}

// IsOversold 判断是否处于超卖状态
// 参数：
//   - threshold: 超卖阈值，可选参数，默认为30
//
// 返回值：
//   - bool: 如果RSI值低于阈值返回true，否则返回false
//
// 示例：
//
//	isOversold := rsi.IsOversold(20)
func (t *TaRSI) IsOversold(threshold ...float64) bool {
	th := 30.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() < th
}

// IsBullishDivergence 判断是否出现多头背离
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - bool: 如果出现多头背离返回true，否则返回false
//
// 说明：
//
//	多头背离指RSI创新高而价格创新低的情况，通常是底部反转信号
//
// 示例：
//
//	isBullish := rsi.IsBullishDivergence(prices)
func (t *TaRSI) IsBullishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	rsiLow := t.Values[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] < rsiLow {
			rsiLow = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] > rsiLow && prices[lastIndex] < priceLow
}

// IsBearishDivergence 判断是否出现空头背离
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - bool: 如果出现空头背离返回true，否则返回false
//
// 说明：
//
//	空头背离指RSI创新低而价格创新高的情况，通常是顶部反转信号
//
// 示例：
//
//	isBearish := rsi.IsBearishDivergence(prices)
func (t *TaRSI) IsBearishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	rsiHigh := t.Values[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] > rsiHigh {
			rsiHigh = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] < rsiHigh && prices[lastIndex] > priceHigh
}

// IsCenterCross 判断是否穿越中轴线
// 返回值：
//   - up: 是否向上穿越50线
//   - down: 是否向下穿越50线
//
// 示例：
//
//	up, down := rsi.IsCenterCross()
func (t *TaRSI) IsCenterCross() (up, down bool) {
	if len(t.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	up = t.Values[lastIndex-1] <= 50 && t.Values[lastIndex] > 50
	down = t.Values[lastIndex-1] >= 50 && t.Values[lastIndex] < 50
	return
}

// GetTrend 获取RSI趋势
// 返回值：
//   - int: 趋势值，1=强势上涨，2=上涨，0=盘整，-1=强势下跌，-2=下跌
//
// 示例：
//
//	trend := rsi.GetTrend()
func (t *TaRSI) GetTrend() int {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	if t.Values[lastIndex] > 70 {
		return 1 // 强势上涨
	} else if t.Values[lastIndex] > 50 {
		return 2 // 上涨
	} else if t.Values[lastIndex] < 30 {
		return -1 // 强势下跌
	} else if t.Values[lastIndex] < 50 {
		return -2 // 下跌
	}
	return 0 // 盘整
}

// GetStrength 获取RSI强度
// 返回值：
//   - float64: RSI值与中轴线(50)的距离
//
// 示例：
//
//	strength := rsi.GetStrength()
func (t *TaRSI) GetStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex] - 50)
}

// IsStrengthening 判断RSI是否在增强
// 返回值：
//   - bool: 如果RSI强度在增加返回true，否则返回false
//
// 示例：
//
//	isStrengthening := rsi.IsStrengthening()
func (t *TaRSI) IsStrengthening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]-50) > math.Abs(t.Values[lastIndex-1]-50)
}

// IsWeakening 判断RSI是否在减弱
// 返回值：
//   - bool: 如果RSI强度在减弱返回true，否则返回false
//
// 示例：
//
//	isWeakening := rsi.IsWeakening()
func (t *TaRSI) IsWeakening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]-50) < math.Abs(t.Values[lastIndex-1]-50)
}

// GetGainLossRatio 获取涨跌比率
// 返回值：
//   - float64: 最新的涨幅与跌幅之比
//
// 示例：
//
//	ratio := rsi.GetGainLossRatio()
func (t *TaRSI) GetGainLossRatio() float64 {
	lastIndex := len(t.Values) - 1
	if t.Losses[lastIndex] == 0 {
		return math.Inf(1)
	}
	return t.Gains[lastIndex] / t.Losses[lastIndex]
}

// IsDivergenceConfirmed 判断背离是否得到确认
// 参数：
//   - prices: 价格序列
//   - threshold: 确认阈值
//
// 返回值：
//   - bool: 如果背离得到确认返回true，否则返回false
//
// 示例：
//
//	isConfirmed := rsi.IsDivergenceConfirmed(prices, 0.1)
func (t *TaRSI) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	rsiChange := (t.Values[lastIndex] - t.Values[lastIndex-1]) / t.Values[lastIndex-1] * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(rsiChange-priceChange) > threshold
}
