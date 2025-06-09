package ta

import (
	"fmt"
)

// TaKDJ KDJ指标结构体(Stochastic Oscillator)
// 说明：
//
//	KDJ是一个超买超卖指标，也称为随机指标，由乔治·莱恩(George Lane)在1950年代开发
//	它通过分析价格在一定周期内的最高价、最低价及收盘价之间的关系，来判断市场的超买超卖状态
type TaKDJ struct {
	K []float64 `json:"k"` // K值，快速线
	D []float64 `json:"d"` // D值，慢速线
	J []float64 `json:"j"` // J值，离差值
}

// CalculateKDJ 计算KDJ指标
// 参数：
//   - high: 最高价序列
//   - low: 最低价序列
//   - close: 收盘价序列
//   - rsvPeriod: RSV计算周期，通常为9
//   - kPeriod: K值计算周期，通常为3
//   - dPeriod: D值计算周期，通常为3
//
// 返回值：
//   - *TaKDJ: KDJ指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	计算步骤：
//	1. 计算RSV = (收盘价 - 最低价) / (最高价 - 最低价) * 100
//	2. K = 2/3 * 前一日K值 + 1/3 * 当日RSV
//	3. D = 2/3 * 前一日D值 + 1/3 * 当日K值
//	4. J = 3 * K - 2 * D
//
// 示例：
//
//	kdj, err := CalculateKDJ(high, low, close, 9, 3, 3)
func CalculateKDJ(high, low, close []float64, rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
	if len(high) < rsvPeriod || len(low) < rsvPeriod || len(close) < rsvPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 4) // [rsv, k, d, j]
	rsv, k, d, j := slices[0], slices[1], slices[2], slices[3]

	// 计算RSV
	for i := rsvPeriod - 1; i < length; i++ {
		// 计算周期内的最高价和最低价
		var highestHigh, lowestLow = high[i], low[i]
		for j := 0; j < rsvPeriod; j++ {
			idx := i - j
			if high[idx] > highestHigh {
				highestHigh = high[idx]
			}
			if low[idx] < lowestLow {
				lowestLow = low[idx]
			}
		}

		// 计算RSV
		if highestHigh != lowestLow {
			rsv[i] = (close[i] - lowestLow) / (highestHigh - lowestLow) * 100
		} else {
			rsv[i] = 50 // 当最高价等于最低价时，RSV取中值
		}
	}

	// 使用递推公式计算KDJ值
	k[rsvPeriod-1] = rsv[rsvPeriod-1]
	d[rsvPeriod-1] = rsv[rsvPeriod-1]
	j[rsvPeriod-1] = rsv[rsvPeriod-1]

	for i := rsvPeriod; i < length; i++ {
		// 计算K值
		k[i] = (2.0*k[i-1] + rsv[i]) / 3.0

		// 计算D值
		d[i] = (2.0*d[i-1] + k[i]) / 3.0

		// 计算J值
		j[i] = 3.0*k[i] - 2.0*d[i]
	}

	return &TaKDJ{
		K: k,
		D: d,
		J: j,
	}, nil
}

// KDJ 计算K线数据的KDJ指标
// 参数：
//   - rsvPeriod: RSV计算周期，通常为9
//   - kPeriod: K值计算周期，通常为3
//   - dPeriod: D值计算周期，通常为3
//
// 返回值：
//   - *TaKDJ: KDJ指标结构体
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	kdj, err := k.KDJ(9, 3, 3)
func (k *KlineDatas) KDJ(rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
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
	return CalculateKDJ(high, low, close, rsvPeriod, kPeriod, dPeriod)
}

// KDJ_ 获取最新的KDJ值
// 参数：
//   - rsvPeriod: RSV计算周期，通常为9
//   - kPeriod: K值计算周期，通常为3
//   - dPeriod: D值计算周期，通常为3
//
// 返回值：
//   - kValue: 最新的K值
//   - dValue: 最新的D值
//   - jValue: 最新的J值
//
// 示例：
//
//	k, d, j := k.KDJ_(9, 3, 3)
func (k *KlineDatas) KDJ_(rsvPeriod, kPeriod, dPeriod int) (kValue, dValue, jValue float64) {
	// 只保留必要的计算数据
	_k, err := k._Keep(rsvPeriod * 2)
	if err != nil {
		_k = *k
	}
	kdj, err := _k.KDJ(rsvPeriod, kPeriod, dPeriod)
	if err != nil {
		return 0, 0, 0
	}
	lastIndex := len(kdj.K) - 1
	return kdj.K[lastIndex], kdj.D[lastIndex], kdj.J[lastIndex]
}

// Value 返回最新的KDJ值
// 返回值：
//   - k: 最新的K值
//   - d: 最新的D值
//   - j: 最新的J值
//
// 示例：
//
//	k, d, j := kdj.Value()
func (t *TaKDJ) Value() (k, d, j float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex], t.J[lastIndex]
}

// IsGoldenCross 判断是否出现金叉信号
// 返回值：
//   - bool: 如果K线从下向上穿越D线返回true，否则返回false
//
// 说明：
//
//	金叉信号是买入信号，表示可能开始上涨
//
// 示例：
//
//	isGolden := kdj.IsGoldenCross()
func (t *TaKDJ) IsGoldenCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] <= t.D[lastIndex-1] && t.K[lastIndex] > t.D[lastIndex]
}

// IsDeathCross 判断是否出现死叉信号
// 返回值：
//   - bool: 如果K线从上向下穿越D线返回true，否则返回false
//
// 说明：
//
//	死叉信号是卖出信号，表示可能开始下跌
//
// 示例：
//
//	isDeath := kdj.IsDeathCross()
func (t *TaKDJ) IsDeathCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] >= t.D[lastIndex-1] && t.K[lastIndex] < t.D[lastIndex]
}

// IsOverbought 判断是否处于超买区域
// 参数：
//   - threshold: 超买阈值，可选参数，默认为80
//
// 返回值：
//   - bool: 如果K值和D值同时大于阈值返回true，否则返回false
//
// 说明：
//
//	在超买区域出现死叉时，卖出信号的可靠性较高
//
// 示例：
//
//	isOverbought := kdj.IsOverbought(80)
func (t *TaKDJ) IsOverbought(threshold ...float64) bool {
	th := 80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] > th && t.D[len(t.D)-1] > th
}

// IsOversold 判断是否处于超卖区域
// 参数：
//   - threshold: 超卖阈值，可选参数，默认为20
//
// 返回值：
//   - bool: 如果K值和D值同时小于阈值返回true，否则返回false
//
// 说明：
//
//	在超卖区域出现金叉时，买入信号的可靠性较高
//
// 示例：
//
//	isOversold := kdj.IsOversold(20)
func (t *TaKDJ) IsOversold(threshold ...float64) bool {
	th := 20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] < th && t.D[len(t.D)-1] < th
}

// IsBullishDivergence 判断是否出现多头背离
// 返回值：
//   - bool: 如果出现多头背离返回true，否则返回false
//
// 说明：
//
//	多头背离指K值在超卖区创新高而价格创新低，是底部反转信号
//
// 示例：
//
//	isBullish := kdj.IsBullishDivergence()
func (t *TaKDJ) IsBullishDivergence() bool {
	if len(t.K) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	previousLow := t.K[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.K[lastIndex-i] < previousLow {
			previousLow = t.K[lastIndex-i]
		}
	}
	return t.K[lastIndex] > previousLow && t.K[lastIndex] < 20
}

// IsBearishDivergence 判断是否出现空头背离
// 返回值：
//   - bool: 如果出现空头背离返回true，否则返回false
//
// 说明：
//
//	空头背离指K值在超买区创新低而价格创新高，是顶部反转信号
//
// 示例：
//
//	isBearish := kdj.IsBearishDivergence()
func (t *TaKDJ) IsBearishDivergence() bool {
	if len(t.K) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	previousHigh := t.K[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.K[lastIndex-i] > previousHigh {
			previousHigh = t.K[lastIndex-i]
		}
	}
	return t.K[lastIndex] < previousHigh && t.K[lastIndex] > 80
}

// IsExtremeBought 判断是否处于极度超买状态
// 返回值：
//   - bool: 如果K值和D值同时大于90返回true，否则返回false
//
// 示例：
//
//	isExtreme := kdj.IsExtremeBought()
func (t *TaKDJ) IsExtremeBought() bool {
	return t.K[len(t.K)-1] > 90 && t.D[len(t.D)-1] > 90
}

// IsExtremeSold 判断是否处于极度超卖状态
// 返回值：
//   - bool: 如果K值和D值同时小于10返回true，否则返回false
//
// 示例：
//
//	isExtreme := kdj.IsExtremeSold()
func (t *TaKDJ) IsExtremeSold() bool {
	return t.K[len(t.K)-1] < 10 && t.D[len(t.D)-1] < 10
}

// IsTrendStrengthening 判断趋势是否在增强
// 返回值：
//   - bool: 如果趋势在增强返回true，否则返回false
//
// 说明：
//
//	通过比较连续三个K值的变化率来判断趋势强度
//
// 示例：
//
//	isStrengthening := kdj.IsTrendStrengthening()
func (t *TaKDJ) IsTrendStrengthening() bool {
	if len(t.K) < 4 {
		return false
	}
	lastIndex := len(t.K) - 1
	diff1 := t.K[lastIndex] - t.K[lastIndex-1]
	diff2 := t.K[lastIndex-1] - t.K[lastIndex-2]
	diff3 := t.K[lastIndex-2] - t.K[lastIndex-3]
	return (diff1 > diff2 && diff2 > diff3) || (diff1 < diff2 && diff2 < diff3)
}

// IsJCrossK 判断J线是否穿越K线
// 返回值：
//   - bullish: J线是否从下向上穿越K线
//   - bearish: J线是否从上向下穿越K线
//
// 说明：
//
//	J线穿越K线可以作为辅助判断趋势转折的信号
//
// 示例：
//
//	bullish, bearish := kdj.IsJCrossK()
func (t *TaKDJ) IsJCrossK() (bullish, bearish bool) {
	if len(t.K) < 2 || len(t.J) < 2 {
		return false, false
	}
	lastIndex := len(t.K) - 1
	bullish = t.J[lastIndex-1] <= t.K[lastIndex-1] && t.J[lastIndex] > t.K[lastIndex]
	bearish = t.J[lastIndex-1] >= t.K[lastIndex-1] && t.J[lastIndex] < t.K[lastIndex]
	return
}
