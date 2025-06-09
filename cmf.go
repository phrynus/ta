package ta

import (
	"fmt"
	"math"
)

// TaCMF 钱德动量指标结构体(Chaikin Money Flow)
// 说明：
//
//	CMF是由Marc Chaikin开发的技术指标，用于衡量资金流向
//	它结合了价格和成交量来判断市场的买卖压力
//	主要应用场景：
//	1. 资金流向：判断资金是流入还是流出
//	2. 趋势确认：通过资金流向验证价格趋势
//	3. 背离分析：价格与资金流向的背离预示趋势转折
//
// 计算公式：
//  1. 计算资金流量乘数(Money Flow Multiplier)：
//     MFM = [(Close - Low) - (High - Close)] / (High - Low)
//  2. 计算资金流量(Money Flow Volume)：
//     MFV = MFM * Volume
//  3. 计算N周期的CMF：
//     CMF = Sum(MFV) / Sum(Volume)
type TaCMF struct {
	Values []float64 `json:"values"` // CMF值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateCMF 计算钱德动量指标
// 参数：
//   - high: 最高价序列
//   - low: 最低价序列
//   - close: 收盘价序列
//   - volume: 成交量序列
//   - period: 计算周期
//
// 返回值：
//   - *TaCMF: CMF指标结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	CMF通过资金流量来判断市场趋势
//	计算步骤：
//	1. 计算每个周期的资金流量乘数
//	2. 计算资金流量
//	3. 计算N周期的CMF值
//
// 示例：
//
//	cmf, err := CalculateCMF(high, low, close, volume, 20)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前CMF值：%.2f\n", cmf.Values[len(cmf.Values)-1])
func CalculateCMF(high, low, close, volume []float64, period int) (*TaCMF, error) {
	if len(high) != len(low) || len(high) != len(close) || len(high) != len(volume) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(high) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(high)
	mfv := make([]float64, length) // 货币流量值
	cmf := make([]float64, length) // CMF值

	// 计算货币流量值
	for i := 0; i < length; i++ {
		if high[i] == low[i] {
			mfv[i] = 0
		} else {
			// 计算货币流量乘数
			mfm := ((close[i] - low[i]) - (high[i] - close[i])) / (high[i] - low[i])
			mfv[i] = mfm * volume[i]
		}
	}

	// 计算CMF
	for i := period - 1; i < length; i++ {
		sumMFV := 0.0
		sumVolume := 0.0
		for j := 0; j < period; j++ {
			sumMFV += mfv[i-j]
			sumVolume += volume[i-j]
		}
		if sumVolume != 0 {
			cmf[i] = sumMFV / sumVolume
		}
	}

	return &TaCMF{
		Values: cmf,
		Period: period,
	}, nil
}

// CMF 计算K线数据的钱德动量指标
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaCMF: CMF指标结构体指针
//   - error: 可能的错误信息
//
// 说明：
//
//	CMF的使用要点：
//	1. CMF > 0 表示资金流入，看涨信号
//	2. CMF < 0 表示资金流出，看跌信号
//	3. CMF的背离可能预示趋势反转
//
// 示例：
//
//	cmf, err := kline.CMF(20, "close")
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) CMF(period int, source string) (*TaCMF, error) {
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
	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateCMF(high, low, close, volume, period)
}

// Value 返回最新的CMF值
// 返回值：
//   - float64: 最新的CMF值
//
// 示例：
//
//	value := cmf.Value()
func (t *TaCMF) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsPositive 判断CMF是否为正
// 返回值：
//   - bool: 如果CMF为正返回true，否则返回false
//
// 说明：
//
//	CMF为正表示资金净流入
//
// 示例：
//
//	isPositive := cmf.IsPositive()
func (t *TaCMF) IsPositive() bool {
	return t.Values[len(t.Values)-1] > 0
}

// IsNegative 判断CMF是否为负
// 返回值：
//   - bool: 如果CMF为负返回true，否则返回false
//
// 说明：
//
//	CMF为负表示资金净流出
//
// 示例：
//
//	isNegative := cmf.IsNegative()
func (t *TaCMF) IsNegative() bool {
	return t.Values[len(t.Values)-1] < 0
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
//	当价格创新低而CMF创新高时，可能出现多头背离
//
// 示例：
//
//	isBullish := cmf.IsBullishDivergence(prices)
func (t *TaCMF) IsBullishDivergence(prices []float64) bool {
	if len(prices) < 20 {
		return false
	}
	lastIndex := len(prices) - 1
	previousLow := prices[lastIndex-1]
	for i := 2; i < 20; i++ {
		if prices[lastIndex-i] < previousLow {
			previousLow = prices[lastIndex-i]
		}
	}
	return prices[lastIndex] > previousLow && t.Values[lastIndex] < 0
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
//	当价格创新高而CMF创新低时，可能出现空头背离
//
// 示例：
//
//	isBearish := cmf.IsBearishDivergence(prices)
func (t *TaCMF) IsBearishDivergence(prices []float64) bool {
	if len(prices) < 20 {
		return false
	}
	lastIndex := len(prices) - 1
	previousHigh := prices[lastIndex-1]
	for i := 2; i < 20; i++ {
		if prices[lastIndex-i] > previousHigh {
			previousHigh = prices[lastIndex-i]
		}
	}
	return prices[lastIndex] < previousHigh && t.Values[lastIndex] > 0
}

// CMF_ 获取最新的CMF值
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的CMF值
//
// 示例：
//
//	value := kline.CMF_(20, "close")
func (k *KlineDatas) CMF_(period int, source string) float64 {
	// 只保留必要的计算数据
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	cmf, err := _k.CMF(period, source)
	if err != nil {
		return 0
	}
	return cmf.Value()
}

// GetStrength 获取资金流强度
// 返回值：
//   - float64: 资金流强度，正值表示流入，负值表示流出
//
// 说明：
//
//	资金流强度反映了市场的买卖力量对比
//
// 示例：
//
//	strength := cmf.GetStrength()
func (t *TaCMF) GetStrength() float64 {
	return t.Values[len(t.Values)-1]
}

// IsCrossZero 判断是否穿越零线
// 返回值：
//   - up: 是否向上穿越零线
//   - down: 是否向下穿越零线
//
// 说明：
//
//	穿越零线可能预示趋势的转换
//
// 示例：
//
//	up, down := cmf.IsCrossZero()
func (t *TaCMF) IsCrossZero() (bool, bool) {
	lastValue := t.Values[len(t.Values)-1]
	return lastValue > 0, lastValue < 0
}

// GetAccumulation 获取累积资金流
// 返回值：
//   - float64: 累积资金流量
//
// 说明：
//
//	累积资金流反映了一段时期内的资金流向
//
// 示例：
//
//	accum := cmf.GetAccumulation()
func (t *TaCMF) GetAccumulation() float64 {
	sum := 0.0
	for _, value := range t.Values {
		sum += value
	}
	return sum
}

// IsStrengthening 判断资金流是否在增强
// 返回值：
//   - bool: 如果资金流在增强返回true，否则返回false
//
// 说明：
//
//	资金流增强可能预示趋势的加强
//
// 示例：
//
//	isStrengthening := cmf.IsStrengthening()
func (t *TaCMF) IsStrengthening() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > t.Values[len(t.Values)-2]
}

// IsWeakening 判断资金流是否在减弱
// 返回值：
//   - bool: 如果资金流在减弱返回true，否则返回false
//
// 说明：
//
//	资金流减弱可能预示趋势的减弱
//
// 示例：
//
//	isWeakening := cmf.IsWeakening()
func (t *TaCMF) IsWeakening() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] < t.Values[len(t.Values)-2]
}

// GetMoneyFlowZone 获取资金流区间
// 返回值：
//   - string: 资金流区间，"strong_inflow"=强流入，"inflow"=流入，"neutral"=中性，"outflow"=流出，"strong_outflow"=强流出
//
// 说明：
//
//	根据CMF值的位置判断当前的资金流向状态
//
// 示例：
//
//	zone := cmf.GetMoneyFlowZone()
func (t *TaCMF) GetMoneyFlowZone() string {
	lastValue := t.Values[len(t.Values)-1]
	if lastValue > 0 {
		return "strong_inflow"
	} else if lastValue > -0.1 && lastValue < 0.1 {
		return "neutral"
	} else {
		return "strong_outflow"
	}
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
//	通过比较CMF和价格的变化率来确认背离信号
//
// 示例：
//
//	isConfirmed := cmf.IsDivergenceConfirmed(prices, 0.1)
func (t *TaCMF) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	lastCMF := t.Values[lastIndex]
	previousCMF := t.Values[lastIndex-1]
	priceChange := prices[lastIndex] - prices[lastIndex-1]
	cmfChange := lastCMF - previousCMF
	return math.Abs(cmfChange/priceChange) > threshold
}

// GetOptimalPeriod 获取最优计算周期
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - int: 建议的最优计算周期
//
// 说明：
//
//	根据价格波动特征自动计算最优的CMF周期
//
// 示例：
//
//	optPeriod := cmf.GetOptimalPeriod(prices)
func (t *TaCMF) GetOptimalPeriod(prices []float64) int {
	if len(prices) < 2 {
		return 1
	}
	period := 1
	for i := 1; i < len(prices); i++ {
		if prices[i] != prices[i-1] {
			period = i + 1
		}
	}
	return period
}
