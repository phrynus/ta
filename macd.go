package ta

import "math"

// TaMacd 用于计算和存储 MACD 指标相关数据的结构体
// 说明：
//
//	该结构体存储了 MACD 指标的计算结果，包括 MACD 线、DIF 线、DEA 线以及计算所需的周期参数
//
// 字段：
//   - Macd: MACD 线的数据数组 (float64 类型)
//   - Dif: DIF 线的数据数组 (float64 类型)
//   - Dea: DEA 线的数据数组 (float64 类型)
//   - ShortPeriod: 短期 EMA 计算的周期 (int 类型)
//   - LongPeriod: 长期 EMA 计算的周期 (int 类型)
//   - SignalPeriod: 信号线计算的周期 (int 类型)
type TaMacd struct {
	Macd         []float64 `json:"macd"`
	Dif          []float64 `json:"dif"`
	Dea          []float64 `json:"dea"`
	ShortPeriod  int       `json:"short_period"`
	LongPeriod   int       `json:"long_period"`
	SignalPeriod int       `json:"signal_period"`
}

// CalculateMACD 根据给定的价格数据和周期参数计算 MACD 指标
// 参数：
//   - prices: 价格数据数组 (float64 类型)
//   - shortPeriod: 短期 EMA 计算的周期 (int 类型)
//   - longPeriod: 长期 EMA 计算的周期 (int 类型)
//   - signalPeriod: 信号线计算的周期 (int 类型)
//
// 返回值：
//   - *TaMacd: 存储 MACD 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该函数依赖 CalculateEMA 函数计算 EMA 值，若计算 EMA 时出错将直接返回错误
//
// 示例：
//
//	macdResult, err := CalculateMACD(prices, 12, 26, 9)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateMACD(prices []float64, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {

	shortEMA, err := CalculateEMA(prices, shortPeriod)
	if err != nil {
		return nil, err
	}
	longEMA, err := CalculateEMA(prices, longPeriod)
	if err != nil {
		return nil, err
	}

	dif := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		if i < longPeriod-1 {
			dif[i] = 0
		} else {
			dif[i] = shortEMA.Values[i] - longEMA.Values[i]
		}
	}

	dea, err := CalculateEMA(dif, signalPeriod)
	if err != nil {
		return nil, err
	}

	macd := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		macd[i] = 2 * (dif[i] - dea.Values[i]) / 2
	}
	return &TaMacd{
		Macd:         macd,
		Dif:          dif,
		Dea:          dea.Values,
		ShortPeriod:  shortPeriod,
		LongPeriod:   longPeriod,
		SignalPeriod: signalPeriod,
	}, nil
}

// MACD 从 KlineDatas 中提取指定来源的价格数据并计算 MACD 指标
// 参数：
//   - source: 价格数据的来源 (string 类型)
//   - shortPeriod: 短期 EMA 计算的周期 (int 类型)
//   - longPeriod: 长期 EMA 计算的周期 (int 类型)
//   - signalPeriod: 信号线计算的周期 (int 类型)
//
// 返回值：
//   - *TaMacd: 存储 MACD 指标计算结果的结构体指针
//   - error: 提取价格数据或计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法依赖 KlineDatas 的 ExtractSlice 方法提取价格数据，若提取失败将直接返回错误
//
// 示例：
//
//	macdResult, err := k.MACD("close", 12, 26, 9)
//	if err != nil {
//	    // 处理错误
//	}
func (k *KlineDatas) MACD(source string, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
}

// MACD_ 从 KlineDatas 中提取指定来源的价格数据并计算 MACD 指标的最后一个值
// 参数：
//   - source: 价格数据的来源 (string 类型)
//   - shortPeriod: 短期 EMA 计算的周期 (int 类型)
//   - longPeriod: 长期 EMA 计算的周期 (int 类型)
//   - signalPeriod: 信号线计算的周期 (int 类型)
//
// 返回值：
//   - macd: MACD 线的最后一个值 (float64 类型)
//   - dif: DIF 线的最后一个值 (float64 类型)
//   - dea: DEA 线的最后一个值 (float64 类型)
//
// 说明/注意事项：
//
//	该方法会先截取部分数据进行计算，若截取失败则使用原始数据
//	若提取价格数据或计算过程中出现错误，将返回 0
//
// 示例：
//
//	macd, dif, dea := k.MACD_("close", 12, 26, 9)
func (k *KlineDatas) MACD_(source string, shortPeriod, longPeriod, signalPeriod int) (macd, dif, dea float64) {
	_k, err := k.Keep((longPeriod + signalPeriod) * 2)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0, 0
	}
	m, err := CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
	if err != nil {
		return 0, 0, 0
	}
	return m.Value()
}

// Value 获取 TaMacd 结构体中 MACD、DIF 和 DEA 线的最后一个值
// 参数：无
// 返回值：
//   - macd: MACD 线的最后一个值 (float64 类型)
//   - dif: DIF 线的最后一个值 (float64 类型)
//   - dea: DEA 线的最后一个值 (float64 类型)
//
// 说明/注意事项：无
// 示例：
//
//	macd, dif, dea := t.Value()
func (t *TaMacd) Value() (macd, dif, dea float64) {
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex], t.Dif[lastIndex], t.Dea[lastIndex]
}

// IsGoldenCross 判断是否发生金叉
// 参数：无
// 返回值：
//   - bool: 是否发生金叉
//
// 说明/注意事项：
//
//	当 DIF 线从下向上穿过 DEA 线时，认为发生金叉
//	若 DIF 或 DEA 线的数据长度小于 2，则直接返回 false
//
// 示例：
//
//	isCross := t.IsGoldenCross()
func (t *TaMacd) IsGoldenCross() bool {
	if len(t.Dif) < 2 || len(t.Dea) < 2 {
		return false
	}
	lastIndex := len(t.Dif) - 1
	return t.Dif[lastIndex-1] <= t.Dea[lastIndex-1] && t.Dif[lastIndex] > t.Dea[lastIndex]
}

// IsDeathCross 判断是否发生死叉
// 参数：无
// 返回值：
//   - bool: 是否发生死叉
//
// 说明/注意事项：
//
//	当 DIF 线从上向下穿过 DEA 线时，认为发生死叉
//	若 DIF 或 DEA 线的数据长度小于 2，则直接返回 false
//
// 示例：
//
//	isCross := t.IsDeathCross()
func (t *TaMacd) IsDeathCross() bool {
	if len(t.Dif) < 2 || len(t.Dea) < 2 {
		return false
	}
	lastIndex := len(t.Dif) - 1
	return t.Dif[lastIndex-1] >= t.Dea[lastIndex-1] && t.Dif[lastIndex] < t.Dea[lastIndex]
}

// CheckDivergence 检查是否出现底背离或顶背离
// 参数：
//   - prices: 价格数据数组 (float64 类型)
//
// 返回值：
//   - bullish: 是否出现底背离
//   - bearish: 是否出现顶背离
func (t *TaMacd) CheckDivergence(prices []float64) (bullish, bearish bool) {
	if len(t.Macd) < 20 || len(prices) < 20 {
		return false, false
	}
	lastIndex := len(t.Macd) - 1

	// 寻找 MACD 和价格的高低点
	macdMin, macdMax := math.MaxFloat64, math.SmallestNonzeroFloat64
	priceMin, priceMax := math.MaxFloat64, math.SmallestNonzeroFloat64
	var macdMinIndex, macdMaxIndex, priceMinIndex, priceMaxIndex int

	for i := 0; i < 20; i++ {
		if t.Macd[lastIndex-i] < macdMin {
			macdMin = t.Macd[lastIndex-i]
			macdMinIndex = lastIndex - i
		}
		if t.Macd[lastIndex-i] > macdMax {
			macdMax = t.Macd[lastIndex-i]
			macdMaxIndex = lastIndex - i
		}
		if prices[lastIndex-i] < priceMin {
			priceMin = prices[lastIndex-i]
			priceMinIndex = lastIndex - i
		}
		if prices[lastIndex-i] > priceMax {
			priceMax = prices[lastIndex-i]
			priceMaxIndex = lastIndex - i
		}
	}

	// 判断底背离和顶背离
	bullish = t.Macd[lastIndex] > macdMin && prices[lastIndex] < priceMin && macdMinIndex < priceMinIndex
	bearish = t.Macd[lastIndex] < macdMax && prices[lastIndex] > priceMax && macdMaxIndex < priceMaxIndex

	return
}

// IsZeroCross 判断 MACD 线是否发生零轴穿越
// 参数：无
// 返回值：
//   - up: 是否向上穿越零轴
//   - down: 是否向下穿越零轴
//
// 说明/注意事项：
//
//	若 MACD 线的数据长度小于 2，则直接返回 false
//
// 示例：
//
//	up, down := t.IsZeroCross()
func (t *TaMacd) IsZeroCross() (up, down bool) {
	if len(t.Macd) < 2 {
		return false, false
	}
	lastIndex := len(t.Macd) - 1
	up = t.Macd[lastIndex-1] <= 0 && t.Macd[lastIndex] > 0
	down = t.Macd[lastIndex-1] >= 0 && t.Macd[lastIndex] < 0
	return
}

// IsHistogramIncreasing 判断 MACD 柱状图是否在增长
// 参数：无
// 返回值：
//   - bool: MACD 柱状图是否在增长
//
// 说明/注意事项：
//
//	若 MACD 线的数据长度小于 2，则直接返回 false
//
// 示例：
//
//	isIncreasing := t.IsHistogramIncreasing()
func (t *TaMacd) IsHistogramIncreasing() bool {
	if len(t.Macd) < 2 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex] > t.Macd[lastIndex-1]
}

// IsHistogramDecreasing 判断 MACD 柱状图是否在减小
// 参数：无
// 返回值：
//   - bool: MACD 柱状图是否在减小
//
// 说明/注意事项：
//
//	若 MACD 线的数据长度小于 2，则直接返回 false
//
// 示例：
//
//	isDecreasing := t.IsHistogramDecreasing()
func (t *TaMacd) IsHistogramDecreasing() bool {
	if len(t.Macd) < 2 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex] < t.Macd[lastIndex-1]
}

// 移除原有的 IsBullishDivergence 和 IsBearishDivergence 方法
// func (t *TaMacd) IsBullishDivergence(prices []float64) bool {
//     ...
// }
// func (t *TaMacd) IsBearishDivergence(prices []float64) bool {
//     ...
// }
