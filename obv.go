package ta

import (
	"fmt"
)

// TaOBV 能量潮指标结构体(On Balance Volume)
// 说明：
//
//	OBV是一种成交量分析指标，通过累计成交量来判断价格趋势
//	它基于这样的理论：成交量变化先于价格变化
//	主要应用场景：
//	1. 趋势确认：通过成交量验证价格趋势
//	2. 背离分析：价格与OBV的背离预示趋势转折
//	3. 支撑阻力：OBV趋势线可作为支撑阻力参考
//
// 计算公式：
//  1. 如果收盘价上涨，OBV = 前一日OBV + 当日成交量
//  2. 如果收盘价下跌，OBV = 前一日OBV - 当日成交量
//  3. 如果收盘价不变，OBV = 前一日OBV
type TaOBV struct {
	Values []float64 `json:"values"` // OBV值序列
}

// CalculateOBV 计算能量潮指标
// 参数：
//   - prices: 价格序列
//   - volumes: 成交量序列
//
// 返回值：
//   - *TaOBV: OBV指标结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	OBV通过累计成交量来反映价格趋势的强弱
//	计算步骤：
//	1. 比较当日收盘价与前一日收盘价
//	2. 根据价格变化方向累加或累减成交量
//	3. 生成OBV值序列
//
// 示例：
//
//	obv, err := CalculateOBV(prices, volumes)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前OBV值：%.2f\n", obv.Values[len(obv.Values)-1])
func CalculateOBV(prices, volumes []float64) (*TaOBV, error) {
	if len(prices) != len(volumes) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(prices) < 2 {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)
	// 预分配结果切片
	obv := make([]float64, length)
	obv[0] = volumes[0] // 第一个值设为当日成交量

	// 并行计算OBV
	parallelProcess(prices, func(data []float64, start, end int) {
		if start == 0 {
			start = 1 // 从第二个元素开始计算
		}
		for i := start; i < end; i++ {
			switch {
			case data[i] > data[i-1]:
				obv[i] = obv[i-1] + volumes[i]
			case data[i] < data[i-1]:
				obv[i] = obv[i-1] - volumes[i]
			default:
				obv[i] = obv[i-1]
			}
		}
	})

	return &TaOBV{
		Values: obv,
	}, nil
}

// OBV 计算K线数据的能量潮指标
// 参数：
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaOBV: OBV指标结构体指针
//   - error: 可能的错误信息
//
// 说明：
//
//	OBV的使用要点：
//	1. OBV上升表示买方力量强于卖方
//	2. OBV下降表示卖方力量强于买方
//	3. OBV趋势线突破可能预示价格突破
//
// 示例：
//
//	obv, err := kline.OBV("close")
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) OBV(source string) (*TaOBV, error) {
	close, err := k._ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	volume, err := k._ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateOBV(close, volume)
}

// OBV_ 获取最新的OBV值
// 参数：
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的OBV值
//
// 示例：
//
//	value := kline.OBV_("close")
func (k *KlineDatas) OBV_(source string) float64 {
	// 只保留必要的计算数据
	_k, err := k._Keep(50) // 保留足够的数据用于趋势判断
	if err != nil {
		_k = *k
	}
	obv, err := _k.OBV(source)
	if err != nil {
		return 0
	}
	return obv.Value()
}

// Value 返回最新的OBV值
// 返回值：
//   - float64: 最新的OBV值
//
// 示例：
//
//	value := obv.Value()
func (t *TaOBV) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsTrendUp 判断OBV趋势是否向上
// 返回值：
//   - bool: 如果OBV趋势向上返回true，否则返回false
//
// 说明：
//
//	OBV趋势向上表示买方力量占优
//
// 示例：
//
//	isUp := obv.IsTrendUp()
func (t *TaOBV) IsTrendUp() bool {
	if len(t.Values) < 5 {
		return false
	}
	lastIndex := len(t.Values) - 1
	// 使用线性回归斜率判断趋势
	var sumX, sumY, sumXY, sumX2 float64
	n := 5
	for i := 0; i < n; i++ {
		x := float64(i)
		y := t.Values[lastIndex-i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	return slope < 0 // 因为i是倒序的，所以斜率为负表示上升趋势
}

// IsTrendDown 判断OBV趋势是否向下
// 返回值：
//   - bool: 如果OBV趋势向下返回true，否则返回false
//
// 说明：
//
//	OBV趋势向下表示卖方力量占优
//
// 示例：
//
//	isDown := obv.IsTrendDown()
func (t *TaOBV) IsTrendDown() bool {
	if len(t.Values) < 5 {
		return false
	}
	lastIndex := len(t.Values) - 1
	// 使用线性回归斜率判断趋势
	var sumX, sumY, sumXY, sumX2 float64
	n := 5
	for i := 0; i < n; i++ {
		x := float64(i)
		y := t.Values[lastIndex-i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	return slope > 0 // 因为i是倒序的，所以斜率为正表示下降趋势
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
//	当价格创新低而OBV创新高时，可能出现多头背离
//
// 示例：
//
//	isBullish := obv.IsBullishDivergence(prices)
func (t *TaOBV) IsBullishDivergence(prices []float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	previousLow := prices[lastIndex-1]
	for i := 2; i < len(prices); i++ {
		if prices[i] < previousLow && t.Values[i] > t.Values[i-1] {
			return true
		}
	}
	return false
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
//	当价格创新高而OBV创新低时，可能出现空头背离
//
// 示例：
//
//	isBearish := obv.IsBearishDivergence(prices)
func (t *TaOBV) IsBearishDivergence(prices []float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	previousHigh := prices[lastIndex-1]
	for i := 2; i < len(prices); i++ {
		if prices[i] > previousHigh && t.Values[i] < t.Values[i-1] {
			return true
		}
	}
	return false
}

// GetTrendStrength 获取趋势强度
// 返回值：
//   - float64: 趋势强度，正值表示上升趋势，负值表示下降趋势
//
// 说明：
//
//	通过计算OBV的变化率来衡量趋势强度
//
// 示例：
//
//	strength := obv.GetTrendStrength()
func (t *TaOBV) GetTrendStrength() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	previousIndex := len(t.Values) - 2
	return t.Values[lastIndex] - t.Values[previousIndex]
}

// IsBreakout 判断是否出现突破
// 参数：
//   - level: 突破水平
//
// 返回值：
//   - bool: 如果OBV突破指定水平返回true，否则返回false
//
// 说明：
//
//	OBV突破重要水平可能预示价格突破
//
// 示例：
//
//	isBreakout := obv.IsBreakout(1000000)
func (t *TaOBV) IsBreakout(level float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > level
}

// GetAccumulation 获取累积量
// 返回值：
//   - float64: 累积成交量
//
// 说明：
//
//	累积量反映了市场的参与度
//
// 示例：
//
//	accum := obv.GetAccumulation()
func (t *TaOBV) GetAccumulation() float64 {
	return t.Values[len(t.Values)-1]
}

// IsVolumeExpanding 判断成交量是否在放大
// 返回值：
//   - bool: 如果成交量在放大返回true，否则返回false
//
// 说明：
//
//	成交量放大通常伴随着趋势的加强
//
// 示例：
//
//	isExpanding := obv.IsVolumeExpanding()
func (t *TaOBV) IsVolumeExpanding() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > t.Values[len(t.Values)-2]
}

// IsVolumeContracting 判断成交量是否在萎缩
// 返回值：
//   - bool: 如果成交量在萎缩返回true，否则返回false
//
// 说明：
//
//	成交量萎缩可能预示趋势即将改变
//
// 示例：
//
//	isContracting := obv.IsVolumeContracting()
func (t *TaOBV) IsVolumeContracting() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] < t.Values[len(t.Values)-2]
}

// GetVolumeForce 获取量能强度
// 返回值：
//   - float64: 量能强度，正值表示买方力量，负值表示卖方力量
//
// 说明：
//
//	量能强度反映了市场的买卖力量对比
//
// 示例：
//
//	force := obv.GetVolumeForce()
func (t *TaOBV) GetVolumeForce() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	return t.Values[len(t.Values)-1] - t.Values[len(t.Values)-2]
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
//	通过比较OBV和价格的变化率来确认背离信号
//
// 示例：
//
//	isConfirmed := obv.IsDivergenceConfirmed(prices, 0.1)
func (t *TaOBV) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	previousIndex := len(prices) - 2
	obvChange := t.Values[lastIndex] - t.Values[previousIndex]
	priceChange := prices[lastIndex] - prices[previousIndex]
	return obvChange > threshold*priceChange
}
