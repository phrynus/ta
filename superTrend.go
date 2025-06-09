package ta

import (
	"fmt"
	"math"
)

// TaSuperTrend 超级趋势指标结构体(SuperTrend Indicator)
// 说明：
//
//	SuperTrend是一个趋势跟踪指标，结合ATR来确定价格波动范围和趋势方向
//	主要用于：
//	1. 识别市场趋势方向
//	2. 确定止损止盈位置
//	3. 捕捉趋势转换信号
//
// 计算公式：
//
//	上轨 = 中点价格 + multiplier * ATR
//	下轨 = 中点价格 - multiplier * ATR
//	中点价格 = (最高价 + 最低价) / 2
type TaSuperTrend struct {
	Upper      []float64 `json:"upper"`      // 上轨线(Upper Band)，价格突破此线视为转多
	Lower      []float64 `json:"lower"`      // 下轨线(Lower Band)，价格跌破此线视为转空
	Trend      []bool    `json:"trend"`      // 趋势方向(Trend Direction)：true表示上涨趋势，false表示下跌趋势
	Period     int       `json:"period"`     // ATR计算周期(ATR Period)
	Multiplier float64   `json:"multiplier"` // ATR乘数(ATR Multiplier)
}

// CalculateSuperTrend 计算超级趋势指标(Calculate SuperTrend)
// 参数：
//   - klineData: K线数据集合
//   - period: ATR计算周期，通常为10
//   - multiplier: ATR乘数，通常为3
//
// 返回值：
//   - *TaSuperTrend: 超级趋势指标结构体指针
//   - error: 错误信息
//
// 说明：
//
//	计算SuperTrend指标，包括上轨、下轨和趋势方向
//
// 示例：
//
//	st, err := CalculateSuperTrend(klineData, 10, 3)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前趋势：%v\n", st.IsUp())
func CalculateSuperTrend(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrend, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 计算ATR
	atr, err := klineData.ATR(period)
	if err != nil {
		return nil, err
	}

	length := len(klineData)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 2) // [upperBand, lowerBand]
	upperBand, lowerBand := slices[0], slices[1]
	trend := make([]bool, length)

	// 计算上下轨
	for i := period; i < length; i++ {
		midpoint := (klineData[i].High + klineData[i].Low) / 2
		atrValue := atr.Values[i]
		upperBand[i] = midpoint + multiplier*atrValue
		lowerBand[i] = midpoint - multiplier*atrValue
	}

	// 初始化趋势
	trend[period] = klineData[period].Close > lowerBand[period]

	// 计算趋势和调整上下轨
	for i := period + 1; i < length; i++ {
		if trend[i-1] { // 当前为多头趋势
			if klineData[i].Close < lowerBand[i] {
				trend[i] = false
				upperBand[i] = upperBand[i-1]
			} else {
				trend[i] = true
				lowerBand[i] = math.Max(lowerBand[i], lowerBand[i-1])
			}
		} else { // 当前为空头趋势
			if klineData[i].Close > upperBand[i] {
				trend[i] = true
				lowerBand[i] = lowerBand[i-1]
			} else {
				trend[i] = false
				upperBand[i] = math.Min(upperBand[i], upperBand[i-1])
			}
		}
	}

	return &TaSuperTrend{
		Upper:      upperBand,
		Lower:      lowerBand,
		Trend:      trend,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

// SuperTrend 计算K线数据的SuperTrend指标(Calculate SuperTrend from Kline)
// 参数：
//   - period: ATR计算周期，通常为7-14日
//   - multiplier: ATR乘数，通常为2-3
//
// 返回值：
//   - *TaSuperTrend: SuperTrend指标结构体指针
//   - error: 错误信息
//
// 示例：
//
//	st, err := kline.SuperTrend(10, 3)
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) SuperTrend(period int, multiplier float64) (*TaSuperTrend, error) {
	return CalculateSuperTrend(*k, period, multiplier)
}

// SuperTrend_ 获取最新的SuperTrend值(Get Latest SuperTrend Values)
// 参数：
//   - period: ATR计算周期
//   - multiplier: ATR乘数
//
// 返回值：
//   - upper: 最新上轨线值
//   - lower: 最新下轨线值
//   - isUpTrend: 是否处于上涨趋势
//
// 示例：
//
//	upper, lower, isUpTrend := kline.SuperTrend_(10, 3)
func (k *KlineDatas) SuperTrend_(period int, multiplier float64) (upper, lower float64, isUpTrend bool) {
	_k, err := k._Keep(period * 14)
	if err != nil {
		_k = *k
	}
	st, err := _k.SuperTrend(period, multiplier)
	if err != nil {
		return 0, 0, false
	}
	lastIndex := len(st.Upper) - 1
	return st.Upper[lastIndex], st.Lower[lastIndex], st.Trend[lastIndex]
}

// Value 返回最新的SuperTrend值(Get Latest Values)
// 返回值：
//   - upper: 最新上轨线值
//   - lower: 最新下轨线值
//   - isUpTrend: 是否处于上涨趋势
//
// 说明：
//
//	获取计算结果中的最新值
//
// 示例：
//
//	upper, lower, isUpTrend := st.Value()
func (t *TaSuperTrend) Value() (upper, lower float64, isUpTrend bool) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

// IsTrendChange 判断是否发生趋势转换(Check Trend Change)
// 返回值：
//   - bool: 是否发生趋势转换
//
// 说明：
//
//	判断最近两个周期的趋势是否发生改变
//
// 示例：
//
//	if st.IsTrendChange() {
//	    fmt.Println("趋势发生转换")
//	}
func (t *TaSuperTrend) IsTrendChange() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex] != t.Trend[lastIndex-1]
}

// IsBullishCross 判断是否发生多头趋势转换(Check Bullish Cross)
// 返回值：
//   - bool: 是否发生多头趋势转换
//
// 说明：
//
//	判断是否从空头趋势转换为多头趋势
//
// 示例：
//
//	if st.IsBullishCross() {
//	    fmt.Println("转换为多头趋势")
//	}
func (t *TaSuperTrend) IsBullishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return !t.Trend[lastIndex-1] && t.Trend[lastIndex]
}

// IsBearishCross 判断是否发生空头趋势转换(Check Bearish Cross)
// 返回值：
//   - bool: 是否发生空头趋势转换
//
// 说明：
//
//	判断是否从多头趋势转换为空头趋势
//
// 示例：
//
//	if st.IsBearishCross() {
//	    fmt.Println("转换为空头趋势")
//	}
func (t *TaSuperTrend) IsBearishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex-1] && !t.Trend[lastIndex]
}

// IsUp 判断最新趋势是否为上涨(Check Uptrend)
// 返回值：
//   - bool: 是否处于上涨趋势
//
// 说明：
//
//	判断当前是否处于上涨趋势
//
// 示例：
//
//	if st.IsUp() {
//	    fmt.Println("当前处于上涨趋势")
//	}
func (t *TaSuperTrend) IsUp() bool {
	return t.Trend[len(t.Trend)-1]
}

// IsDown 判断最新趋势是否为下跌(Check Downtrend)
// 返回值：
//   - bool: 是否处于下跌趋势
//
// 说明：
//
//	判断当前是否处于下跌趋势
//
// 示例：
//
//	if st.IsDown() {
//	    fmt.Println("当前处于下跌趋势")
//	}
func (t *TaSuperTrend) IsDown() bool {
	return !t.Trend[len(t.Trend)-1]
}

// GetUpper 获取最新的上轨线值(Get Latest Upper Band)
// 返回值：
//   - float64: 最新的上轨线值
//
// 说明：
//
//	获取最新计算的上轨线值
//
// 示例：
//
//	upperBand := st.GetUpper()
func (t *TaSuperTrend) GetUpper() float64 {
	return t.Upper[len(t.Upper)-1]
}

// GetLower 获取最新的下轨线值(Get Latest Lower Band)
// 返回值：
//   - float64: 最新的下轨线值
//
// 说明：
//
//	获取最新计算的下轨线值
//
// 示例：
//
//	lowerBand := st.GetLower()
func (t *TaSuperTrend) GetLower() float64 {
	return t.Lower[len(t.Lower)-1]
}

// GetTrendStrength 获取趋势强度(Get Trend Strength)
// 返回值：
//   - float64: 趋势强度值
//
// 说明：
//
//	计算当前趋势的强度，值越大表示趋势越强
//
// 示例：
//
//	strength := st.GetTrendStrength()
//	fmt.Printf("当前趋势强度：%.2f\n", strength)
func (t *TaSuperTrend) GetTrendStrength() float64 {
	lastIndex := len(t.Upper) - 1
	if t.Trend[lastIndex] {
		return t.Upper[lastIndex] - t.Lower[lastIndex]
	}
	return t.Lower[lastIndex] - t.Upper[lastIndex]
}

// IsTrendStrengthening 判断趋势是否增强(Check if Trend is Strengthening)
// 返回值：
//   - bool: 趋势是否正在增强
//
// 说明：
//
//	通过比较最近几个周期的趋势强度变化判断
//
// 示例：
//
//	if st.IsTrendStrengthening() {
//	    fmt.Println("趋势正在增强")
//	}
func (t *TaSuperTrend) IsTrendStrengthening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength > previousStrength
}

// IsTrendWeakening 判断趋势是否减弱(Check if Trend is Weakening)
// 返回值：
//   - bool: 趋势是否正在减弱
//
// 说明：
//
//	通过比较最近几个周期的趋势强度变化判断
//
// 示例：
//
//	if st.IsTrendWeakening() {
//	    fmt.Println("趋势正在减弱")
//	}
func (t *TaSuperTrend) IsTrendWeakening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength < previousStrength
}

// GetTrendDuration 获取当前趋势持续时间(Get Trend Duration)
// 返回值：
//   - int: 趋势持续的周期数
//
// 说明：
//
//	计算当前趋势已经持续的周期数
//
// 示例：
//
//	duration := st.GetTrendDuration()
//	fmt.Printf("当前趋势已持续%d个周期\n", duration)
func (t *TaSuperTrend) GetTrendDuration() int {
	if len(t.Trend) < 2 {
		return 0
	}
	lastIndex := len(t.Trend) - 1
	currentTrend := t.Trend[lastIndex]
	duration := 1
	for i := lastIndex - 1; i >= 0; i-- {
		if t.Trend[i] != currentTrend {
			break
		}
		duration++
	}
	return duration
}

// GetBandwidth 获取带宽(Get Bandwidth)
// 返回值：
//   - float64: 上下轨之间的带宽
//
// 说明：
//
//	计算上下轨之间的价格区间大小
//
// 示例：
//
//	bandwidth := st.GetBandwidth()
//	fmt.Printf("当前带宽：%.2f\n", bandwidth)
func (t *TaSuperTrend) GetBandwidth() float64 {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex] - t.Lower[lastIndex]
}

// IsBreakoutPossible 判断是否可能发生突破(Check Possible Breakout)
// 参数：
//   - threshold: 突破阈值(可选)
//
// 返回值：
//   - bool: 是否可能发生突破
//
// 说明：
//
//	通过分析价格与轨道的关系判断是否可能发生趋势突破
//
// 示例：
//
//	if st.IsBreakoutPossible(0.1) {
//	    fmt.Println("可能即将发生趋势突破")
//	}
func (t *TaSuperTrend) IsBreakoutPossible(threshold ...float64) bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	th := 0.1
	if len(threshold) > 0 {
		th = threshold[0]
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return math.Abs(currentBandwidth-previousBandwidth)/previousBandwidth > th
}

// GetTrendQuality 获取趋势质量(Get Trend Quality)
// 返回值：
//   - float64: 趋势质量得分，范围[0,1]
//
// 说明：
//
//	评估当前趋势的质量，考虑趋势持续时间、强度等因素
//
// 示例：
//
//	quality := st.GetTrendQuality()
//	fmt.Printf("当前趋势质量：%.2f\n", quality)
func (t *TaSuperTrend) GetTrendQuality() float64 {
	duration := t.GetTrendDuration()
	strength := t.GetTrendStrength()
	return float64(duration) * strength
}
