package ta

import (
	"math"
	"sync"
)

// Package ta 提供技术分析指标的计算功能

// TaMacd MACD指标结构体(Moving Average Convergence/Divergence)
// 说明：
//
//	MACD是由Gerald Appel于1979年提出的，用于判断股票价格趋势的技术分析指标
//	它通过比较两条不同速度的移动平均线，来对价格的波动趋势进行预测
//	主要应用场景：
//	1. 趋势确认：MACD能有效确认市场趋势的形成和转折
//	2. 背离判断：价格与MACD的背离可以预示趋势的可能转折
//	3. 超买超卖：MACD柱状图的极值可以指示市场的超买超卖状态
type TaMacd struct {
	Macd         []float64 `json:"macd"`          // MACD线，表示DIF与DEA的差值，用于判断买卖信号
	Dif          []float64 `json:"dif"`           // 差离值，快速与慢速移动平均线的差
	Dea          []float64 `json:"dea"`           // 讯号线，DIF的移动平均，也称为MACD线
	ShortPeriod  int       `json:"short_period"`  // 快线周期
	LongPeriod   int       `json:"long_period"`   // 慢线周期
	SignalPeriod int       `json:"signal_period"` // 信号线周期
}

// CalculateMACD 计算移动平均趋势指标(Moving Average Convergence/Divergence)
// 参数：
//   - prices: 价格序列
//   - shortPeriod: 快线周期，通常为12
//   - longPeriod: 慢线周期，通常为26
//   - signalPeriod: 信号线周期，通常为9
//
// 返回值：
//   - *TaMacd: 包含MACD、Signal和Histogram的结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	MACD是一种趋势跟踪动量指标，用于判断股票价格趋势
//	计算步骤：
//	1. 计算快速EMA(shortPeriod)和慢速EMA(longPeriod)
//	2. 计算DIF = 快速EMA - 慢速EMA
//	3. 计算DEA = DIF的signalPeriod周期EMA
//	4. 计算MACD = 2 * (DIF - DEA)
//
// 示例：
//
//	macd, err := CalculateMACD(prices, 12, 26, 9)
func CalculateMACD(prices []float64, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {
	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 5) // [shortEMA, longEMA, dif, dea, macd]
	shortEMA, longEMA, dif, dea, macd := slices[0], slices[1], slices[2], slices[3], slices[4]

	// 并行计算EMA
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		calculateEMA(prices, shortPeriod, shortEMA)
	}()

	go func() {
		defer wg.Done()
		calculateEMA(prices, longPeriod, longEMA)
	}()

	wg.Wait()

	// 计算DIF
	for i := longPeriod - 1; i < length; i++ {
		dif[i] = shortEMA[i] - longEMA[i]
	}

	// 计算DEA
	calculateEMA(dif, signalPeriod, dea)

	// 计算MACD
	for i := 0; i < length; i++ {
		macd[i] = 2 * (dif[i] - dea[i])
	}

	return &TaMacd{
		Macd:         macd,
		Dif:          dif,
		Dea:          dea,
		ShortPeriod:  shortPeriod,
		LongPeriod:   longPeriod,
		SignalPeriod: signalPeriod,
	}, nil
}

// calculateEMA 优化的EMA计算函数
func calculateEMA(prices []float64, period int, result []float64) {
	if len(prices) < period {
		return
	}

	// 计算第一个EMA值
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// 计算乘数
	multiplier := 2.0 / float64(period+1)
	oneMinusMultiplier := 1.0 - multiplier

	// 使用递推公式计算后续的EMA值
	for i := period; i < len(prices); i++ {
		result[i] = prices[i]*multiplier + result[i-1]*oneMinusMultiplier
	}
}

// MACD 计算K线数据的MACD指标
// 参数：
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//   - shortPeriod: 快速EMA周期，通常为12日
//   - longPeriod: 慢速EMA周期，通常为26日
//   - signalPeriod: DEA信号线周期，通常为9日
//
// 返回值：
//   - *TaMacd: 包含MACD、DIF、DEA三个指标数据的结构体
//   - error: 计算过程中可能出现的错误
//
// 说明：
//  1. 建议使用收盘价(close)作为计算基础
//  2. 指标需要足够的历史数据才能产生有效信号
//  3. 应结合其他技术指标和基本面分析使用
//
// 示例：
//
//	macd, err := kline.MACD("close", 12, 26, 9)
func (k *KlineDatas) MACD(source string, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {
	prices, err := k._ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
}

// MACD_ 获取最新的MACD值
// 参数：
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//   - shortPeriod: 快速EMA周期，通常为12日
//   - longPeriod: 慢速EMA周期，通常为26日
//   - signalPeriod: DEA信号线周期，通常为9日
//
// 返回值：
//   - macd: 最新的MACD值
//   - dif: 最新的DIF值
//   - dea: 最新的DEA值
//
// 示例：
//
//	macd, dif, dea := kline.MACD_("close", 12, 26, 9)
func (k *KlineDatas) MACD_(source string, shortPeriod, longPeriod, signalPeriod int) (macd, dif, dea float64) {
	_k, err := k._Keep((longPeriod + signalPeriod) * 2)
	if err != nil {
		_k = *k
	}
	prices, err := _k._ExtractSlice(source)
	if err != nil {
		return 0, 0, 0
	}
	m, err := CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
	if err != nil {
		return 0, 0, 0
	}
	return m.Value()
}

// Value 返回最新的MACD值
// 返回值：
//   - macd: 最新的MACD值
//   - dif: 最新的DIF值
//   - dea: 最新的DEA值
//
// 示例：
//
//	macd, dif, dea := m.Value()
func (t *TaMacd) Value() (macd, dif, dea float64) {
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex], t.Dif[lastIndex], t.Dea[lastIndex]
}

// IsGoldenCross 判断是否出现金叉信号
// 返回值：
//   - bool: 如果出现金叉返回true，否则返回false
//
// 说明：
//
//	金叉信号出现在DIF从下向上穿越DEA时
//
// 示例：
//
//	isGolden := macd.IsGoldenCross()
func (t *TaMacd) IsGoldenCross() bool {
	if len(t.Dif) < 2 || len(t.Dea) < 2 {
		return false
	}
	lastIndex := len(t.Dif) - 1
	return t.Dif[lastIndex-1] <= t.Dea[lastIndex-1] && t.Dif[lastIndex] > t.Dea[lastIndex]
}

// IsDeathCross 判断是否出现死叉信号
// 返回值：
//   - bool: 如果出现死叉返回true，否则返回false
//
// 说明：
//
//	死叉信号出现在DIF从上向下穿越DEA时
//
// 示例：
//
//	isDeath := macd.IsDeathCross()
func (t *TaMacd) IsDeathCross() bool {
	if len(t.Dif) < 2 || len(t.Dea) < 2 {
		return false
	}
	lastIndex := len(t.Dif) - 1
	return t.Dif[lastIndex-1] >= t.Dea[lastIndex-1] && t.Dif[lastIndex] < t.Dea[lastIndex]
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
//	多头背离指MACD创新高而价格创新低的情况，通常是底部反转信号
//
// 示例：
//
//	isBullish := macd.IsBullishDivergence(prices)
func (t *TaMacd) IsBullishDivergence(prices []float64) bool {
	if len(t.Macd) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	macdLow := t.Macd[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Macd[lastIndex-i] < macdLow {
			macdLow = t.Macd[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.Macd[lastIndex] > macdLow && prices[lastIndex] < priceLow
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
//	空头背离指MACD创新低而价格创新高的情况，通常是顶部反转信号
//
// 示例：
//
//	isBearish := macd.IsBearishDivergence(prices)
func (t *TaMacd) IsBearishDivergence(prices []float64) bool {
	if len(t.Macd) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	macdHigh := t.Macd[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Macd[lastIndex-i] > macdHigh {
			macdHigh = t.Macd[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.Macd[lastIndex] < macdHigh && prices[lastIndex] > priceHigh
}

// IsZeroCross 判断是否穿越零轴
// 返回值：
//   - up: 是否向上穿越零轴
//   - down: 是否向下穿越零轴
//
// 说明：
//
//	零轴上方为多头市场，零轴下方为空头市场
//
// 示例：
//
//	up, down := macd.IsZeroCross()
func (t *TaMacd) IsZeroCross() (up, down bool) {
	if len(t.Macd) < 2 {
		return false, false
	}
	lastIndex := len(t.Macd) - 1
	up = t.Macd[lastIndex-1] <= 0 && t.Macd[lastIndex] > 0
	down = t.Macd[lastIndex-1] >= 0 && t.Macd[lastIndex] < 0
	return
}

// GetTrend 获取MACD趋势
// 返回值：
//   - int: 趋势值，1=强势上涨，2=上涨，0=盘整，-1=强势下跌，-2=下跌
//
// 示例：
//
//	trend := macd.GetTrend()
func (t *TaMacd) GetTrend() int {
	if len(t.Macd) < t.SignalPeriod {
		return 0
	}
	lastIndex := len(t.Macd) - 1
	if t.Macd[lastIndex] > 0 && t.Dif[lastIndex] > t.Dea[lastIndex] {
		return 1 // 强势上涨
	} else if t.Macd[lastIndex] > 0 && t.Dif[lastIndex] < t.Dea[lastIndex] {
		return 2 // 上涨减速
	} else if t.Macd[lastIndex] < 0 && t.Dif[lastIndex] < t.Dea[lastIndex] {
		return -1 // 强势下跌
	} else if t.Macd[lastIndex] < 0 && t.Dif[lastIndex] > t.Dea[lastIndex] {
		return -2 // 下跌减速
	}
	return 0 // 盘整
}

// GetHistogramWidth 获取MACD柱状图宽度
// 返回值：
//   - float64: MACD柱状图的宽度（绝对值）
//
// 示例：
//
//	width := macd.GetHistogramWidth()
func (t *TaMacd) GetHistogramWidth() float64 {
	return t.Macd[len(t.Macd)-1]
}

// IsHistogramIncreasing 判断MACD柱状图是否在增加
// 返回值：
//   - bool: 如果柱状图在增加返回true，否则返回false
//
// 示例：
//
//	isIncreasing := macd.IsHistogramIncreasing()
func (t *TaMacd) IsHistogramIncreasing() bool {
	if len(t.Macd) < 2 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex] > t.Macd[lastIndex-1]
}

// IsHistogramDecreasing 判断MACD柱状图是否在减少
// 返回值：
//   - bool: 如果柱状图在减少返回true，否则返回false
//
// 示例：
//
//	isDecreasing := macd.IsHistogramDecreasing()
func (t *TaMacd) IsHistogramDecreasing() bool {
	if len(t.Macd) < 2 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex] < t.Macd[lastIndex-1]
}

// GetConvergence 获取MACD收敛度
// 返回值：
//   - float64: DIF与DEA之间的收敛程度
//
// 示例：
//
//	convergence := macd.GetConvergence()
func (t *TaMacd) GetConvergence() float64 {
	if len(t.Dif) < 1 || len(t.Dea) < 1 {
		return 0
	}
	lastIndex := len(t.Dif) - 1
	return math.Abs(t.Dif[lastIndex] - t.Dea[lastIndex])
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
//	isConfirmed := macd.IsDivergenceConfirmed(prices, 0.1)
func (t *TaMacd) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.Macd) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	macdChange := (t.Macd[lastIndex] - t.Macd[lastIndex-1]) / math.Abs(t.Macd[lastIndex-1]) * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(macdChange-priceChange) > threshold
}
