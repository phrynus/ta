package ta

import (
	"fmt"
	"math"
)

// TaWilliamsR Williams %R指标结构体(Williams Percent Range)
// 说明：
//
//	Williams %R是一个动量指标，用于衡量市场的超买超卖状态
//	取值范围为0到-100，通常-20以上为超买区，-80以下为超卖区
//	主要用于：
//	1. 判断超买超卖状态
//	2. 寻找趋势反转信号
//	3. 确认价格背离
//
// 计算公式：
//
//	Williams %R = (最高价 - 收盘价) / (最高价 - 最低价) * -100
type TaWilliamsR struct {
	Values []float64 `json:"values"` // Williams %R值序列，取值范围[-100,0]
	Period int       `json:"period"` // 计算周期
}

// CalculateWilliamsR 计算威廉指标(Calculate Williams %R)
// 参数：
//   - high: 最高价数组
//   - low: 最低价数组
//   - close: 收盘价数组
//   - period: 计算周期
//
// 返回值：
//   - *TaWilliamsR: 威廉指标结构体指针
//   - error: 错误信息
//
// 说明：
//
//	计算指定周期的Williams %R值
//
// 示例：
//
//	wr, err := CalculateWilliamsR(high, low, close, 14)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前Williams %%R值：%v\n", wr.Value())
func CalculateWilliamsR(high, low, close []float64, period int) (*TaWilliamsR, error) {
	if len(high) < period || len(low) < period || len(close) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 1) // [wr]
	wr := slices[0]

	// 计算Williams %R
	for i := period - 1; i < length; i++ {
		// 计算周期内的最高价和最低价
		var highestHigh, lowestLow = high[i], low[i]
		for j := 0; j < period; j++ {
			idx := i - j
			if high[idx] > highestHigh {
				highestHigh = high[idx]
			}
			if low[idx] < lowestLow {
				lowestLow = low[idx]
			}
		}

		// 计算Williams %R
		if highestHigh != lowestLow {
			wr[i] = ((highestHigh - close[i]) / (highestHigh - lowestLow)) * -100
		} else {
			wr[i] = -50 // 当最高价等于最低价时，取中值
		}
	}

	return &TaWilliamsR{
		Values: wr,
		Period: period,
	}, nil
}

// WilliamsR 计算K线数据的威廉指标(Calculate Williams %R from Kline)
// 参数：
//   - period: 计算周期(默认14)
//
// 返回值：
//   - *TaWilliamsR: 威廉指标结构体指针
//   - error: 错误信息
//
// 示例：
//
//	wr, err := kline.WilliamsR(14)
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) WilliamsR(period int) (*TaWilliamsR, error) {
	high, err := k._ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k._ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	close, err := k._ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	return CalculateWilliamsR(high, low, close, period)
}

// WilliamsR_ 获取最新的威廉指标值(Get Latest Williams %R Value)
// 参数：
//   - period: 计算周期(默认14)
//
// 返回值：
//   - float64: 最新的Williams %R值
//
// 示例：
//
//	wrValue := kline.WilliamsR_(14)
func (k *KlineDatas) WilliamsR_(period int) float64 {
	// 只保留必要的计算数据
	_k, err := k._Keep(period * 2)
	if err != nil {
		_k = *k
	}
	wr, err := _k.WilliamsR(period)
	if err != nil {
		return 0
	}
	return wr.Value()
}

// IsWilliamsROverbought 判断是否处于超买区域(Check Overbought)
// 参数：
//   - wr: Williams %R值
//   - threshold: 超买阈值(可选，默认-20)
//
// 返回值：
//   - bool: 是否处于超买状态
//
// 说明：
//
//	当Williams %R值高于阈值时，表示市场处于超买状态
//
// 示例：
//
//	if IsWilliamsROverbought(wrValue, -20) {
//	    fmt.Println("市场处于超买状态")
//	}
func IsWilliamsROverbought(wr float64, threshold ...float64) bool {
	th := -20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return wr > th
}

// IsWilliamsROversold 判断是否处于超卖区域(Check Oversold)
// 参数：
//   - wr: Williams %R值
//   - threshold: 超卖阈值(可选，默认-80)
//
// 返回值：
//   - bool: 是否处于超卖状态
//
// 说明：
//
//	当Williams %R值低于阈值时，表示市场处于超卖状态
//
// 示例：
//
//	if IsWilliamsROversold(wrValue, -80) {
//	    fmt.Println("市场处于超卖状态")
//	}
func IsWilliamsROversold(wr float64, threshold ...float64) bool {
	th := -80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return wr < th
}

// IsWilliamsRCrossOver 判断是否向上穿越指定水平(Check Cross Over)
// 参数：
//   - current: 当前值
//   - previous: 前一个值
//   - level: 水平线值
//
// 返回值：
//   - bool: 是否向上穿越
//
// 说明：
//
//	判断Williams %R是否从下向上穿越指定水平线
//
// 示例：
//
//	if IsWilliamsRCrossOver(current, previous, -80) {
//	    fmt.Println("Williams %R向上穿越-80，可能是买入信号")
//	}
func IsWilliamsRCrossOver(current, previous, level float64) bool {
	return previous < level && current >= level
}

// IsWilliamsRCrossUnder 判断是否向下穿越指定水平(Check Cross Under)
// 参数：
//   - current: 当前值
//   - previous: 前一个值
//   - level: 水平线值
//
// 返回值：
//   - bool: 是否向下穿越
//
// 说明：
//
//	判断Williams %R是否从上向下穿越指定水平线
//
// 示例：
//
//	if IsWilliamsRCrossUnder(current, previous, -20) {
//	    fmt.Println("Williams %R向下穿越-20，可能是卖出信号")
//	}
func IsWilliamsRCrossUnder(current, previous, level float64) bool {
	return previous > level && current <= level
}

// Value 返回最新的Williams %R值(Get Latest Value)
// 返回值：
//   - float64: 最新的Williams %R值
//
// 说明：
//
//	获取计算结果中的最新值
//
// 示例：
//
//	latestValue := wr.Value()
func (t *TaWilliamsR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsOverbought 判断是否处于超买状态(Check Overbought State)
// 参数：
//   - threshold: 超买阈值(可选，默认-20)
//
// 返回值：
//   - bool: 是否处于超买状态
//
// 说明：
//
//	判断当前Williams %R值是否处于超买区域
//
// 示例：
//
//	if wr.IsOverbought(-20) {
//	    fmt.Println("当前处于超买状态")
//	}
func (t *TaWilliamsR) IsOverbought(threshold ...float64) bool {
	th := -20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() > th
}

// IsOversold 判断是否处于超卖状态(Check Oversold State)
// 参数：
//   - threshold: 超卖阈值(可选，默认-80)
//
// 返回值：
//   - bool: 是否处于超卖状态
//
// 说明：
//
//	判断当前Williams %R值是否处于超卖区域
//
// 示例：
//
//	if wr.IsOversold(-80) {
//	    fmt.Println("当前处于超卖状态")
//	}
func (t *TaWilliamsR) IsOversold(threshold ...float64) bool {
	th := -80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() < th
}

// IsBuySignal 判断是否出现买入信号(Check Buy Signal)
// 参数：
//   - threshold: 超卖阈值(可选，默认-80)
//
// 返回值：
//   - bool: 是否出现买入信号
//
// 说明：
//
//	当Williams %R从超卖区域向上突破时，产生买入信号
//
// 示例：
//
//	if wr.IsBuySignal(-80) {
//	    fmt.Println("出现买入信号")
//	}
func (t *TaWilliamsR) IsBuySignal(threshold ...float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	th := -80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] < th && t.Values[lastIndex] > th
}

// IsSellSignal 判断是否出现卖出信号(Check Sell Signal)
// 参数：
//   - threshold: 超买阈值(可选，默认-20)
//
// 返回值：
//   - bool: 是否出现卖出信号
//
// 说明：
//
//	当Williams %R从超买区域向下突破时，产生卖出信号
//
// 示例：
//
//	if wr.IsSellSignal(-20) {
//	    fmt.Println("出现卖出信号")
//	}
func (t *TaWilliamsR) IsSellSignal(threshold ...float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	th := -20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] > th && t.Values[lastIndex] < th
}

// IsBullishDivergence 判断是否出现多头背离(Check Bullish Divergence)
// 参数：
//   - prices: 价格数组
//
// 返回值：
//   - bool: 是否出现多头背离
//
// 说明：
//
//	当价格创新低而Williams %R未创新低时，形成多头背离
//
// 示例：
//
//	if wr.IsBullishDivergence(prices) {
//	    fmt.Println("出现多头背离信号")
//	}
func (t *TaWilliamsR) IsBullishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	wrLow := t.Values[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] < wrLow {
			wrLow = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] > wrLow && prices[lastIndex] < priceLow
}

// IsBearishDivergence 判断是否出现空头背离(Check Bearish Divergence)
// 参数：
//   - prices: 价格数组
//
// 返回值：
//   - bool: 是否出现空头背离
//
// 说明：
//
//	当价格创新高而Williams %R未创新高时，形成空头背离
//
// 示例：
//
//	if wr.IsBearishDivergence(prices) {
//	    fmt.Println("出现空头背离信号")
//	}
func (t *TaWilliamsR) IsBearishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	wrHigh := t.Values[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] > wrHigh {
			wrHigh = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] < wrHigh && prices[lastIndex] > priceHigh
}

// IsCenterCross 判断是否穿越中轴线(Check Center Line Cross)
// 返回值：
//   - up: 向上穿越
//   - down: 向下穿越
//
// 说明：
//
//	判断Williams %R是否穿越-50中轴线
//
// 示例：
//
//	up, down := wr.IsCenterCross()
//	if up {
//	    fmt.Println("向上穿越中轴线")
//	}
func (t *TaWilliamsR) IsCenterCross() (up, down bool) {
	if len(t.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	up = t.Values[lastIndex-1] < -50 && t.Values[lastIndex] > -50
	down = t.Values[lastIndex-1] > -50 && t.Values[lastIndex] < -50
	return
}

// GetTrend 获取趋势方向(Get Trend Direction)
// 返回值：
//   - int: 1表示上升趋势，-1表示下降趋势，0表示横盘
//
// 说明：
//
//	根据Williams %R的位置和变化判断趋势方向
//
// 示例：
//
//	trend := wr.GetTrend()
//	if trend > 0 {
//	    fmt.Println("当前处于上升趋势")
//	}
func (t *TaWilliamsR) GetTrend() int {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	if t.Values[lastIndex] > -20 {
		return 1 // 强势上涨
	} else if t.Values[lastIndex] > -50 {
		return 2 // 上涨
	} else if t.Values[lastIndex] < -80 {
		return -1 // 强势下跌
	} else if t.Values[lastIndex] < -50 {
		return -2 // 下跌
	}
	return 0 // 盘整
}

// GetStrength 获取趋势强度(Get Trend Strength)
// 返回值：
//   - float64: 趋势强度值，范围[0,1]
//
// 说明：
//
//	计算当前趋势的强度
//
// 示例：
//
//	strength := wr.GetStrength()
//	fmt.Printf("当前趋势强度：%.2f\n", strength)
func (t *TaWilliamsR) GetStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex] + 50)
}

// IsStrengthening 判断趋势是否增强(Check if Trend is Strengthening)
// 返回值：
//   - bool: 趋势是否正在增强
//
// 说明：
//
//	通过比较最近几个周期的趋势强度变化判断
//
// 示例：
//
//	if wr.IsStrengthening() {
//	    fmt.Println("趋势正在增强")
//	}
func (t *TaWilliamsR) IsStrengthening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]+50) > math.Abs(t.Values[lastIndex-1]+50)
}

// IsWeakening 判断趋势是否减弱(Check if Trend is Weakening)
// 返回值：
//   - bool: 趋势是否正在减弱
//
// 说明：
//
//	通过比较最近几个周期的趋势强度变化判断
//
// 示例：
//
//	if wr.IsWeakening() {
//	    fmt.Println("趋势正在减弱")
//	}
func (t *TaWilliamsR) IsWeakening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]+50) < math.Abs(t.Values[lastIndex-1]+50)
}

// GetMomentum 获取动量值(Get Momentum)
// 返回值：
//   - float64: 动量值
//
// 说明：
//
//	计算Williams %R的动量值，用于判断趋势的加速或减速
//
// 示例：
//
//	momentum := wr.GetMomentum()
//	fmt.Printf("当前动量：%.2f\n", momentum)
func (t *TaWilliamsR) GetMomentum() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

// IsDivergenceConfirmed 判断背离是否得到确认(Check if Divergence is Confirmed)
// 参数：
//   - prices: 价格数组
//   - threshold: 确认阈值
//
// 返回值：
//   - bool: 背离是否得到确认
//
// 说明：
//
//	通过价格行为和其他条件确认背离的有效性
//
// 示例：
//
//	if wr.IsDivergenceConfirmed(prices, 0.1) {
//	    fmt.Println("背离得到确认")
//	}
func (t *TaWilliamsR) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	wrChange := (t.Values[lastIndex] - t.Values[lastIndex-1]) / math.Abs(t.Values[lastIndex-1]) * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(wrChange-priceChange) > threshold
}

// GetZonePosition 获取所在区域位置(Get Zone Position)
// 返回值：
//   - int: 2表示超买区，1表示中性偏上区，0表示中性区，-1表示中性偏下区，-2表示超卖区
//
// 说明：
//
//	判断Williams %R当前所处的区域位置
//
// 示例：
//
//	position := wr.GetZonePosition()
//	if position == 2 {
//	    fmt.Println("当前处于超买区")
//	}
func (t *TaWilliamsR) GetZonePosition() int {
	value := t.Value()
	if value > -20 {
		return 1 // 超买区
	} else if value > -50 {
		return 2 // 上方区域
	} else if value < -80 {
		return -1 // 超卖区
	} else if value < -50 {
		return -2 // 下方区域
	}
	return 0 // 中性区域
}
