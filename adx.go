package ta

import (
	"fmt"
	"math"
)

// TaADX 平均趋向指标(ADX)的计算结果结构体
// 说明：
//
//	ADX指标用于衡量市场趋势的强度，不论趋势是上涨还是下跌
//
// 字段：
//   - ADX: ADX指标值数组，表示趋势强度
//   - PlusDI: +DI指标值数组，表示上升趋势的强度
//   - MinusDI: -DI指标值数组，表示下降趋势的强度
//   - Period: 计算周期
type TaADX struct {
	ADX     []float64 `json:"adx"`
	PlusDI  []float64 `json:"plus_di"`
	MinusDI []float64 `json:"minus_di"`
	Period  int       `json:"period"`
}

// CalculateADX 计算给定K线数据的ADX、+DI和-DI指标
// 参数：
//   - klineData: K线数据数组，包含OHLC价格数据
//   - period: 计算周期
//
// 返回值：
//   - *TaADX: 包含计算结果的TaADX结构体指针
//   - error: 计算过程中的错误，如数据不足等
//
// 说明：
//
//	该函数实现了ADX指标的完整计算过程，包括：
//	1. 计算+DM和-DM
//	2. 计算真实波幅(TR)
//	3. 计算平滑后的+DI和-DI
//	4. 最终计算ADX值
//	计算过程采用Wilder平滑方法
//
// 示例：
//
//	adx, err := CalculateADX(klineData, 14)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("ADX: %v, +DI: %v, -DI: %v\n", adx.ADX[len(adx.ADX)-1], adx.PlusDI[len(adx.PlusDI)-1], adx.MinusDI[len(adx.MinusDI)-1])
func CalculateADX(klineData KlineDatas, period int) (*TaADX, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)

	slices := preallocateSlices(length, 6)
	plusDM, minusDM, trueRange, plusDI, minusDI, adx := slices[0], slices[1], slices[2], slices[3], slices[4], slices[5]

	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevHigh := klineData[i-1].High
		prevLow := klineData[i-1].Low

		upMove := high - prevHigh
		downMove := prevLow - low

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}

		tr1 := high - low
		tr2 := math.Abs(high - klineData[i-1].Close)
		tr3 := math.Abs(low - klineData[i-1].Close)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	var smoothPlusDM, smoothMinusDM, smoothTR float64

	for i := 1; i <= period; i++ {
		smoothPlusDM += plusDM[i]
		smoothMinusDM += minusDM[i]
		smoothTR += trueRange[i]
	}

	if smoothTR > 0 {
		plusDI[period] = 100 * smoothPlusDM / smoothTR
		minusDI[period] = 100 * smoothMinusDM / smoothTR
	}

	for i := period + 1; i < length; i++ {

		smoothPlusDM = smoothPlusDM - (smoothPlusDM / float64(period)) + plusDM[i]
		smoothMinusDM = smoothMinusDM - (smoothMinusDM / float64(period)) + minusDM[i]
		smoothTR = smoothTR - (smoothTR / float64(period)) + trueRange[i]

		if smoothTR > 0 {
			plusDI[i] = 100 * smoothPlusDM / smoothTR
			minusDI[i] = 100 * smoothMinusDM / smoothTR
		}

		diSum := math.Abs(plusDI[i] - minusDI[i])
		diDiff := plusDI[i] + minusDI[i]
		if diDiff > 0 {
			adx[i] = 100 * diSum / diDiff
		}
	}

	var smoothADX float64
	for i := period * 2; i < length; i++ {
		if i == period*2 {

			for j := period; j <= i; j++ {
				smoothADX += adx[j]
			}
			adx[i] = smoothADX / float64(period+1)
		} else {

			adx[i] = (adx[i-1]*float64(period-1) + adx[i]) / float64(period)
		}
	}

	return &TaADX{
		ADX:     adx,
		PlusDI:  plusDI,
		MinusDI: minusDI,
		Period:  period,
	}, nil
}

// ADX 计算K线数据的ADX指标
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - *TaADX: ADX计算结果
//   - error: 计算过程中的错误
func (k *KlineDatas) ADX(period int) (*TaADX, error) {
	return CalculateADX(*k, period)
}

// ADX_ 计算并返回最新的ADX、+DI和-DI值
// 参数：
//   - period: 计算周期
//
// 返回值：
//   - adx: 最新的ADX值
//   - plusDI: 最新的+DI值
//   - minusDI: 最新的-DI值
func (k *KlineDatas) ADX_(period int) (adx, plusDI, minusDI float64) {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	adxData, err := _k.ADX(period)
	if err != nil {
		return 0, 0, 0
	}
	return adxData.Value()
}

// Value 获取最新的ADX、+DI和-DI值
// 返回值：
//   - adx: 最新的ADX值
//   - plusDI: 最新的+DI值
//   - minusDI: 最新的-DI值
func (t *TaADX) Value() (adx, plusDI, minusDI float64) {
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex], t.PlusDI[lastIndex], t.MinusDI[lastIndex]
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// IsTrendStrengthening 判断趋势是否正在增强
// 返回值：
//   - bool: 如果当前ADX值大于前一个ADX值则返回true，表示趋势正在增强
func (t *TaADX) IsTrendStrengthening() bool {
	if len(t.ADX) < 2 {
		return false
	}
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex] > t.ADX[lastIndex-1]
}

// IsTrendWeakening 判断趋势是否正在减弱
// 返回值：
//   - bool: 如果当前ADX值小于前一个ADX值则返回true，表示趋势正在减弱
func (t *TaADX) IsTrendWeakening() bool {
	if len(t.ADX) < 2 {
		return false
	}
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex] < t.ADX[lastIndex-1]
}

// IsBullishCrossover 判断是否出现多头交叉
// 返回值：
//   - bool: 如果+DI从下向上穿过-DI则返回true，表示出现多头交叉
func (t *TaADX) IsBullishCrossover() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex-1] <= t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] > t.MinusDI[lastIndex]
}

// IsBearishCrossover 判断是否出现空头交叉
// 返回值：
//   - bool: 如果+DI从上向下穿过-DI则返回true，表示出现空头交叉
func (t *TaADX) IsBearishCrossover() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex-1] >= t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] < t.MinusDI[lastIndex]
}

// IsDIConverging 判断DI是否正在收敛
// 返回值：
//   - bool: 如果+DI和-DI之间的距离正在缩小则返回true
func (t *TaADX) IsDIConverging() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	currentSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	previousSpread := math.Abs(t.PlusDI[lastIndex-1] - t.MinusDI[lastIndex-1])
	return currentSpread < previousSpread
}

// IsDIDiverging 判断DI是否正在发散
// 返回值：
//   - bool: 如果+DI和-DI之间的距离正在增大则返回true
func (t *TaADX) IsDIDiverging() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	currentSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	previousSpread := math.Abs(t.PlusDI[lastIndex-1] - t.MinusDI[lastIndex-1])
	return currentSpread > previousSpread
}
