package ta

import (
	"fmt"
	"math"
)

// TaStochRSI StochRSI指标结构体(Stochastic Relative Strength Index)
// 说明：
//
//	StochRSI是由Tushar Chande和Stanley Kroll于1994年提出的技术指标，
//	它结合了RSI和随机指标的特点，用于判断市场的超买超卖状态。
//
// 主要应用场景：
//  1. 超买超卖：K值和D值在80以上为超买，20以下为超卖
//  2. 趋势反转：K线与D线的交叉可以作为交易信号
//  3. 背离判断：价格与StochRSI的背离可以预示趋势的可能转折
//
// 计算公式：
//  1. 计算RSI
//  2. StochRSI = (RSI - 最低RSI) / (最高RSI - 最低RSI)
//  3. K = StochRSI的N日移动平均
//  4. D = K的M日移动平均
type TaStochRSI struct {
	K           []float64 `json:"k"`            // K值，快速线
	D           []float64 `json:"d"`            // D值，慢速线
	RsiPeriod   int       `json:"rsi_period"`   // RSI计算周期
	StochPeriod int       `json:"stoch_period"` // StochRSI计算周期
	KPeriod     int       `json:"k_period"`     // K值平滑周期
	DPeriod     int       `json:"d_period"`     // D值平滑周期
}

// CalculateStochRSI 计算随机相对强弱指标(Stochastic RSI)
// 参数：
//   - prices: 价格序列
//   - rsiPeriod: RSI计算周期，通常为14
//   - stochPeriod: StochRSI计算周期，通常为14
//   - kPeriod: K值平滑周期，通常为3
//   - dPeriod: D值平滑周期，通常为3
//
// 返回值：
//   - *TaStochRSI: StochRSI指标结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	StochRSI结合了RSI的动量特性和KDJ的随机特性，能更好地识别超买超卖机会
//	计算步骤：
//	1. 首先计算指定周期的RSI值
//	2. 将RSI值代入随机指标公式计算StochRSI
//	3. 对StochRSI进行平滑处理得到K值
//	4. 对K值进行平滑处理得到D值
//
// 示例：
//
//	stochRsi, err := CalculateStochRSI(prices, 14, 14, 3, 3)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前K值：%.2f，D值：%.2f\n", stochRsi.K[len(stochRsi.K)-1], stochRsi.D[len(stochRsi.D)-1])
func CalculateStochRSI(prices []float64, rsiPeriod, stochPeriod, kPeriod, dPeriod int) (*TaStochRSI, error) {
	if len(prices) < rsiPeriod+stochPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 计算RSI
	rsi, err := CalculateRSI(prices, rsiPeriod)
	if err != nil {
		return nil, err
	}

	length := len(prices)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 3) // [stochRsi, k, d]
	stochRsi, k, d := slices[0], slices[1], slices[2]

	// 计算StochRSI
	for i := stochPeriod - 1; i < length; i++ {
		// 计算周期内的最高和最低RSI
		var highestRsi, lowestRsi = rsi.Values[i], rsi.Values[i]
		for j := 0; j < stochPeriod; j++ {
			idx := i - j
			if rsi.Values[idx] > highestRsi {
				highestRsi = rsi.Values[idx]
			}
			if rsi.Values[idx] < lowestRsi {
				lowestRsi = rsi.Values[idx]
			}
		}

		// 计算StochRSI
		if highestRsi != lowestRsi {
			stochRsi[i] = (rsi.Values[i] - lowestRsi) / (highestRsi - lowestRsi) * 100
		} else {
			stochRsi[i] = 50 // 当最高RSI等于最低RSI时，取中值
		}
	}

	// 计算K值（%K的移动平均）
	var sumK float64
	for i := 0; i < kPeriod && i < length; i++ {
		sumK += stochRsi[i]
	}
	k[kPeriod-1] = sumK / float64(kPeriod)

	// 使用滑动窗口计算后续的K值
	for i := kPeriod; i < length; i++ {
		sumK = sumK - stochRsi[i-kPeriod] + stochRsi[i]
		k[i] = sumK / float64(kPeriod)
	}

	// 计算D值（%K的移动平均）
	var sumD float64
	for i := 0; i < dPeriod && i < length; i++ {
		sumD += k[i]
	}
	d[dPeriod-1] = sumD / float64(dPeriod)

	// 使用滑动窗口计算后续的D值
	for i := dPeriod; i < length; i++ {
		sumD = sumD - k[i-dPeriod] + k[i]
		d[i] = sumD / float64(dPeriod)
	}

	return &TaStochRSI{
		K:           k,
		D:           d,
		RsiPeriod:   rsiPeriod,
		StochPeriod: stochPeriod,
		KPeriod:     kPeriod,
		DPeriod:     dPeriod,
	}, nil
}

// StochRSI 计算K线数据的StochRSI指标
// 参数：
//   - rsiPeriod: RSI计算周期，通常为14
//   - stochPeriod: StochRSI计算周期，通常为14
//   - kPeriod: K值平滑周期，通常为3
//   - dPeriod: D值平滑周期，通常为3
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaStochRSI: StochRSI指标结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	StochRSI是一个复合指标，结合了RSI和随机指标的优点
//	主要用于：
//	1. 判断超买超卖：K值和D值同时高于80为超买，低于20为超卖
//	2. 交叉信号：K线穿越D线产生交易信号
//	3. 背离分析：价格与指标的背离可预示趋势转折
//
// 示例：
//
//	stochRsi, err := kline.StochRSI(14, 14, 3, 3, "close")
//	if err != nil {
//	    return err
//	}
//	// 判断是否出现超买信号
//	if stochRsi.IsOverbought() {
//	    // 执行卖出逻辑
//	}
func (k *KlineDatas) StochRSI(rsiPeriod, stochPeriod, kPeriod, dPeriod int, source string) (*TaStochRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateStochRSI(prices, rsiPeriod, stochPeriod, kPeriod, dPeriod)
}

// StochRSI_ 获取最新的StochRSI值
// 参数：
//   - rsiPeriod: RSI计算周期，通常为14
//   - stochPeriod: StochRSI计算周期，通常为14
//   - kPeriod: K值平滑周期，通常为3
//   - dPeriod: D值平滑周期，通常为3
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - kValue: 最新的K值
//   - dValue: 最新的D值
//
// 示例：
//
//	k, d := kline.StochRSI_(14, 14, 3, 3, "close")
func (k *KlineDatas) StochRSI_(rsiPeriod, stochPeriod, kPeriod, dPeriod int, source string) (kValue, dValue float64) {
	_k, err := k.Keep((rsiPeriod + stochPeriod) * 2)
	if err != nil {
		_k = *k
	}
	stochRsi, err := _k.StochRSI(rsiPeriod, stochPeriod, kPeriod, dPeriod, source)
	if err != nil {
		return 0, 0
	}
	return stochRsi.Value()
}

// Value 返回最新的K值和D值
// 返回值：
//   - kValue: 最新的K值
//   - dValue: 最新的D值
//
// 示例：
//
//	k, d := stochRsi.Value()
func (t *TaStochRSI) Value() (kValue, dValue float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex]
}

// IsOverbought 判断是否处于超买状态
// 参数：
//   - threshold: 超买阈值，可选参数，默认为80
//
// 返回值：
//   - bool: 如果K值和D值同时大于阈值返回true，否则返回false
//
// 示例：
//
//	isOverbought := stochRsi.IsOverbought(80)
func (t *TaStochRSI) IsOverbought(threshold ...float64) bool {
	th := 80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] > th && t.D[len(t.D)-1] > th
}

// IsOversold 判断是否处于超卖状态
// 参数：
//   - threshold: 超卖阈值，可选参数，默认为20
//
// 返回值：
//   - bool: 如果K值和D值同时小于阈值返回true，否则返回false
//
// 示例：
//
//	isOversold := stochRsi.IsOversold(20)
func (t *TaStochRSI) IsOversold(threshold ...float64) bool {
	th := 20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] < th && t.D[len(t.D)-1] < th
}

// IsGoldenCross 判断是否出现金叉信号
// 返回值：
//   - bool: 如果K线从下向上穿越D线返回true，否则返回false
//
// 说明：
//
//	金叉信号出现在K线从下向上穿越D线时，通常是买入信号
//
// 示例：
//
//	isGolden := stochRsi.IsGoldenCross()
func (t *TaStochRSI) IsGoldenCross() bool {
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
//	死叉信号出现在K线从上向下穿越D线时，通常是卖出信号
//
// 示例：
//
//	isDeath := stochRsi.IsDeathCross()
func (t *TaStochRSI) IsDeathCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] >= t.D[lastIndex-1] && t.K[lastIndex] < t.D[lastIndex]
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
//	多头背离指K值在超卖区创新高而价格创新低，是底部反转信号
//
// 示例：
//
//	isBullish := stochRsi.IsBullishDivergence(prices)
func (t *TaStochRSI) IsBullishDivergence(prices []float64) bool {
	if len(t.K) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	kLow := t.K[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.K[lastIndex-i] < kLow {
			kLow = t.K[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.K[lastIndex] > kLow && prices[lastIndex] < priceLow
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
//	空头背离指K值在超买区创新低而价格创新高，是顶部反转信号
//
// 示例：
//
//	isBearish := stochRsi.IsBearishDivergence(prices)
func (t *TaStochRSI) IsBearishDivergence(prices []float64) bool {
	if len(t.K) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	kHigh := t.K[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.K[lastIndex-i] > kHigh {
			kHigh = t.K[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.K[lastIndex] < kHigh && prices[lastIndex] > priceHigh
}

// IsCenterCross 判断是否穿越中轴线
// 返回值：
//   - up: 是否向上穿越50线
//   - down: 是否向下穿越50线
//
// 说明：
//
//	穿越中轴线(50)可以作为趋势转换的早期信号
//
// 示例：
//
//	up, down := stochRsi.IsCenterCross()
func (t *TaStochRSI) IsCenterCross() (up, down bool) {
	if len(t.K) < 2 {
		return false, false
	}
	lastIndex := len(t.K) - 1
	up = t.K[lastIndex-1] <= 50 && t.K[lastIndex] > 50
	down = t.K[lastIndex-1] >= 50 && t.K[lastIndex] < 50
	return
}

// GetTrend 获取趋势方向
// 返回值：
//   - int: 趋势值，1=强势上涨，2=上涨，0=盘整，-1=强势下跌，-2=下跌
//
// 说明：
//
//	根据K值的位置判断当前趋势状态
//
// 示例：
//
//	trend := stochRsi.GetTrend()
func (t *TaStochRSI) GetTrend() int {
	lastIndex := len(t.K) - 1
	if t.K[lastIndex] > 80 {
		return 1 // 强势上涨
	} else if t.K[lastIndex] > 50 {
		return 2 // 上涨
	} else if t.K[lastIndex] < 20 {
		return -1 // 强势下跌
	} else if t.K[lastIndex] < 50 {
		return -2 // 下跌
	}
	return 0 // 盘整
}

// GetStrength 获取指标强度
// 返回值：
//   - float64: K值与中轴线(50)的距离
//
// 说明：
//
//	指标强度反映了当前趋势的强弱程度
//
// 示例：
//
//	strength := stochRsi.GetStrength()
func (t *TaStochRSI) GetStrength() float64 {
	lastIndex := len(t.K) - 1
	return math.Abs(t.K[lastIndex] - 50)
}

// IsStrengthening 判断指标是否在增强
// 返回值：
//   - bool: 如果指标强度在增加返回true，否则返回false
//
// 说明：
//
//	通过比较当前和前一个周期的强度来判断趋势是否在增强
//
// 示例：
//
//	isStrengthening := stochRsi.IsStrengthening()
func (t *TaStochRSI) IsStrengthening() bool {
	if len(t.K) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return math.Abs(t.K[lastIndex]-50) > math.Abs(t.K[lastIndex-1]-50)
}

// IsWeakening 判断指标是否在减弱
// 返回值：
//   - bool: 如果指标强度在减弱返回true，否则返回false
//
// 说明：
//
//	通过比较当前和前一个周期的强度来判断趋势是否在减弱
//
// 示例：
//
//	isWeakening := stochRsi.IsWeakening()
func (t *TaStochRSI) IsWeakening() bool {
	if len(t.K) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return math.Abs(t.K[lastIndex]-50) < math.Abs(t.K[lastIndex-1]-50)
}

// GetKDSpread 获取K-D差值
// 返回值：
//   - float64: K值与D值的差值
//
// 说明：
//
//	K-D差值可以用来判断趋势的强度和潜在的反转信号
//
// 示例：
//
//	spread := stochRsi.GetKDSpread()
func (t *TaStochRSI) GetKDSpread() float64 {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex] - t.D[lastIndex]
}

// IsKDConverging 判断K-D是否收敛
// 返回值：
//   - bool: 如果K线和D线距离在缩小返回true，否则返回false
//
// 说明：
//
//	K线和D线的收敛通常预示着趋势可能即将改变
//
// 示例：
//
//	isConverging := stochRsi.IsKDConverging()
func (t *TaStochRSI) IsKDConverging() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	currentSpread := math.Abs(t.K[lastIndex] - t.D[lastIndex])
	previousSpread := math.Abs(t.K[lastIndex-1] - t.D[lastIndex-1])
	return currentSpread < previousSpread
}

// IsKDDiverging 判断K-D是否发散
// 返回值：
//   - bool: 如果K线和D线距离在扩大返回true，否则返回false
//
// 说明：
//
//	K线和D线的发散通常表示当前趋势正在增强
//
// 示例：
//
//	isDiverging := stochRsi.IsKDDiverging()
func (t *TaStochRSI) IsKDDiverging() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	currentSpread := math.Abs(t.K[lastIndex] - t.D[lastIndex])
	previousSpread := math.Abs(t.K[lastIndex-1] - t.D[lastIndex-1])
	return currentSpread > previousSpread
}

// GetMomentum 获取动量
// 返回值：
//   - float64: K值的变化量
//
// 说明：
//
//	动量值反映了指标变化的速度，可用于判断趋势强度
//
// 示例：
//
//	momentum := stochRsi.GetMomentum()
func (t *TaStochRSI) GetMomentum() float64 {
	if len(t.K) < 2 {
		return 0
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex] - t.K[lastIndex-1]
}

// IsDivergenceConfirmed 判断背离是否确认
// 参数：
//   - prices: 价格序列
//   - threshold: 确认阈值
//
// 返回值：
//   - bool: 如果背离得到确认返回true，否则返回false
//
// 说明：
//
//	通过比较K值和价格的变化率来确认背离信号的有效性
//
// 示例：
//
//	isConfirmed := stochRsi.IsDivergenceConfirmed(prices, 0.1)
func (t *TaStochRSI) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.K) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	kChange := (t.K[lastIndex] - t.K[lastIndex-1]) / math.Abs(t.K[lastIndex-1]) * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(kChange-priceChange) > threshold
}

// GetZonePosition 获取区域位置
// 返回值：
//   - int: 区域位置，1=超买区，2=上方区域，0=中性区域，-1=超卖区，-2=下方区域
//
// 说明：
//
//	根据K值所处的位置来判断当前的市场状态
//
// 示例：
//
//	position := stochRsi.GetZonePosition()
func (t *TaStochRSI) GetZonePosition() int {
	value := t.K[len(t.K)-1]
	if value > 80 {
		return 1 // 超买区
	} else if value > 50 {
		return 2 // 上方区域
	} else if value < 20 {
		return -1 // 超卖区
	} else if value < 50 {
		return -2 // 下方区域
	}
	return 0 // 中性区域
}
