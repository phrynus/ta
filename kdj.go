package ta

import (
	"fmt"
)

// TaKDJ 表示 KDJ 指标的计算结果结构体
// 说明：
//
//	该结构体用于存储 KDJ 指标的计算结果，包含 K、D、J 三条线的值
//
// 字段：
//   - K: K 线的值数组 (float64 类型)
//   - D: D 线的值数组 (float64 类型)
//   - J: J 线的值数组 (float64 类型)
type TaKDJ struct {
	K []float64 `json:"k"`
	D []float64 `json:"d"`
	J []float64 `json:"j"`
}

// CalculateKDJ 计算 KDJ 指标
// 参数：
//   - high: 最高价数组
//   - low: 最低价数组
//   - close: 收盘价数组
//   - rsvPeriod: RSV 的计算周期
//   - kPeriod: K 线的计算周期
//   - dPeriod: D 线的计算周期
//
// 返回值：
//   - *TaKDJ: 包含 KDJ 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	输入的 high、low、close 数组长度必须不小于 rsvPeriod，否则将返回错误
//
// 示例：
//
//	high := []float64{...}
//	low := []float64{...}
//	close := []float64{...}
//	kdj, err := CalculateKDJ(high, low, close, 9, 3, 3)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateKDJ(high, low, close []float64, rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
	if len(high) < rsvPeriod || len(low) < rsvPeriod || len(close) < rsvPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)

	slices := preallocateSlices(length, 4)
	rsv, k, d, j := slices[0], slices[1], slices[2], slices[3]

	for i := rsvPeriod - 1; i < length; i++ {

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

		if highestHigh != lowestLow {
			rsv[i] = (close[i] - lowestLow) / (highestHigh - lowestLow) * 100
		} else {
			rsv[i] = 50
		}
	}

	k[rsvPeriod-1] = rsv[rsvPeriod-1]
	d[rsvPeriod-1] = rsv[rsvPeriod-1]
	j[rsvPeriod-1] = rsv[rsvPeriod-1]

	for i := rsvPeriod; i < length; i++ {

		k[i] = (2.0*k[i-1] + rsv[i]) / 3.0

		d[i] = (2.0*d[i-1] + k[i]) / 3.0

		j[i] = 3.0*k[i] - 2.0*d[i]
	}

	return &TaKDJ{
		K: k,
		D: d,
		J: j,
	}, nil
}

// KDJ 计算 K 线数据的 KDJ 指标
// 参数：
//   - rsvPeriod: RSV 的计算周期
//   - kPeriod: K 线的计算周期
//   - dPeriod: D 线的计算周期
//
// 返回值：
//   - *TaKDJ: 包含 KDJ 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//
//	该方法会从 KlineDatas 中提取 high、low、close 数据进行计算
//
// 示例：
//
//	klineData := &KlineDatas{...}
//	kdj, err := klineData.KDJ(9, 3, 3)
//	if err != nil {
//	    // 处理错误
//	}
func (k *KlineDatas) KDJ(rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
	high, err := k.ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k.ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	return CalculateKDJ(high, low, close, rsvPeriod, kPeriod, dPeriod)
}

// KDJ_ 计算 K 线数据的 KDJ 指标的最后一个值
// 参数：
//   - rsvPeriod: RSV 的计算周期
//   - kPeriod: K 线的计算周期
//   - dPeriod: D 线的计算周期
//
// 返回值：
//   - kValue: K 线的最后一个值
//   - dValue: D 线的最后一个值
//   - jValue: J 线的最后一个值
//
// 说明/注意事项：
//
//	该方法会截取最近的 rsvPeriod * 2 条数据进行计算，如果截取失败则使用全部数据
//
// 示例：
//
//	klineData := &KlineDatas{...}
//	k, d, j := klineData.KDJ_(9, 3, 3)
func (k *KlineDatas) KDJ_(rsvPeriod, kPeriod, dPeriod int) (kValue, dValue, jValue float64) {
	_k, err := k.Keep(rsvPeriod * 2)
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

// Value 获取 TaKDJ 结构体中 K、D、J 线的最后一个值
// 返回值：
//   - k: K 线的最后一个值
//   - d: D 线的最后一个值
//   - j: J 线的最后一个值
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	k, d, j := kdj.Value()
func (t *TaKDJ) Value() (k, d, j float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex], t.J[lastIndex]
}

// IsGoldenCross 判断是否出现金叉
// 返回值：
//   - bool: 如果出现金叉则返回 true，否则返回 false
//
// 说明/注意事项：
//
//	金叉定义为 K 线从下向上穿过 D 线
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	if kdj.IsGoldenCross() {
//	    // 处理金叉事件
//	}
func (t *TaKDJ) IsGoldenCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] <= t.D[lastIndex-1] && t.K[lastIndex] > t.D[lastIndex]
}

// IsDeathCross 判断是否出现死叉
// 返回值：
//   - bool: 如果出现死叉则返回 true，否则返回 false
//
// 说明/注意事项：
//
//	死叉定义为 K 线从上向下穿过 D 线
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	if kdj.IsDeathCross() {
//	    // 处理死叉事件
//	}
func (t *TaKDJ) IsDeathCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] >= t.D[lastIndex-1] && t.K[lastIndex] < t.D[lastIndex]
}

// IsOverbought 判断是否超买
// 参数：
//   - threshold: 可选参数，超买阈值，默认为 80.0
//
// 返回值：
//   - bool: 如果 K 线和 D 线都超过阈值则返回 true，否则返回 false
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	if kdj.IsOverbought() {
//	    // 处理超买事件
//	}
func (t *TaKDJ) IsOverbought(threshold ...float64) bool {
	th := 80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] > th && t.D[len(t.D)-1] > th
}

// IsOversold 判断是否超卖
// 参数：
//   - threshold: 可选参数，超卖阈值，默认为 20.0
//
// 返回值：
//   - bool: 如果 K 线和 D 线都低于阈值则返回 true，否则返回 false
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	if kdj.IsOversold() {
//	    // 处理超卖事件
//	}
func (t *TaKDJ) IsOversold(threshold ...float64) bool {
	th := 20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] < th && t.D[len(t.D)-1] < th
}

// IsBullishDivergence 判断是否出现底背离
// 返回值：
//   - bool: 如果出现底背离则返回 true，否则返回 false
//
// 说明/注意事项：
//
//	底背离定义为 K 线在低位形成的低点逐步抬高
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	if kdj.IsBullishDivergence() {
//	    // 处理底背离事件
//	}
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

// IsBearishDivergence 判断是否出现顶背离
// 返回值：
//   - bool: 如果出现顶背离则返回 true，否则返回 false
//
// 说明/注意事项：
//
//	顶背离定义为 K 线在高位形成的高点逐步降低
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	if kdj.IsBearishDivergence() {
//	    // 处理顶背离事件
//	}
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

// IsTrendStrengthening 判断趋势是否加强
// 返回值：
//   - bool: 如果趋势加强则返回 true，否则返回 false
//
// 说明/注意事项：
//
//	趋势加强定义为 K 线的差值呈递增或递减趋势
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	if kdj.IsTrendStrengthening() {
//	    // 处理趋势加强事件
//	}
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

// IsJCrossK 判断 J 线是否穿过 K 线
// 返回值：
//   - bullish: 如果 J 线从下向上穿过 K 线则返回 true，否则返回 false
//   - bearish: 如果 J 线从上向下穿过 K 线则返回 true，否则返回 false
//
// 示例：
//
//	kdj := &TaKDJ{...}
//	bullish, bearish := kdj.IsJCrossK()
//	if bullish {
//	    // 处理 J 线上穿 K 线事件
//	}
//	if bearish {
//	    // 处理 J 线下穿 K 线事件
//	}
func (t *TaKDJ) IsJCrossK() (bullish, bearish bool) {
	if len(t.K) < 2 || len(t.J) < 2 {
		return false, false
	}
	lastIndex := len(t.K) - 1
	bullish = t.J[lastIndex-1] <= t.K[lastIndex-1] && t.J[lastIndex] > t.K[lastIndex]
	bearish = t.J[lastIndex-1] >= t.K[lastIndex-1] && t.J[lastIndex] < t.K[lastIndex]
	return
}
