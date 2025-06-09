package ta

import (
	"fmt"
	"math"
)

// TaADX ADX指标结构体(Average Directional Index)
type TaADX struct {
	ADX     []float64 `json:"adx"`      // ADX值序列
	PlusDI  []float64 `json:"plus_di"`  // +DI值序列
	MinusDI []float64 `json:"minus_di"` // -DI值序列
	Period  int       `json:"period"`   // 计算周期
}

// CalculateADX 计算平均趋向指标(Average Directional Index)
// 参数：
//   - klineData: K线数据集合
//   - period: 计算周期，通常为14
//
// 返回值：
//   - *TaADX: ADX值序列，+DI值序列，-DI值序列
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	ADX是一个技术分析指标，用于衡量趋势的强度而不考虑趋势方向
//	计算步骤：
//	1. 计算方向变动：+DM（上升动向）和-DM（下降动向）
//	2. 计算真实波幅TR
//	3. 对+DM、-DM和TR进行period周期的平滑处理
//	4. 计算方向指标：+DI = (+DM/TR)*100，-DI = (-DM/TR)*100
//	5. 计算方向指数：DX = |+DI - -DI| / |+DI + -DI| * 100
//	6. 计算ADX：对DX进行period周期的平滑处理
//
// 示例：
//
//	adx, err := CalculateADX(klineData, 14)
func CalculateADX(klineData KlineDatas, period int) (*TaADX, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 6) // [plusDM, minusDM, trueRange, plusDI, minusDI, adx]
	plusDM, minusDM, trueRange, plusDI, minusDI, adx := slices[0], slices[1], slices[2], slices[3], slices[4], slices[5]

	// 计算+DM、-DM和TR
	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevHigh := klineData[i-1].High
		prevLow := klineData[i-1].Low

		// 计算方向移动
		upMove := high - prevHigh
		downMove := prevLow - low

		// 计算+DM和-DM
		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}

		// 计算真实波幅
		tr1 := high - low
		tr2 := math.Abs(high - klineData[i-1].Close)
		tr3 := math.Abs(low - klineData[i-1].Close)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	// 计算平滑值
	var smoothPlusDM, smoothMinusDM, smoothTR float64

	// 初始化第一个周期的平滑值
	for i := 1; i <= period; i++ {
		smoothPlusDM += plusDM[i]
		smoothMinusDM += minusDM[i]
		smoothTR += trueRange[i]
	}

	// 计算第一个周期的DI值
	if smoothTR > 0 {
		plusDI[period] = 100 * smoothPlusDM / smoothTR
		minusDI[period] = 100 * smoothMinusDM / smoothTR
	}

	// 计算后续周期的DI值
	for i := period + 1; i < length; i++ {
		// 更新平滑值
		smoothPlusDM = smoothPlusDM - (smoothPlusDM / float64(period)) + plusDM[i]
		smoothMinusDM = smoothMinusDM - (smoothMinusDM / float64(period)) + minusDM[i]
		smoothTR = smoothTR - (smoothTR / float64(period)) + trueRange[i]

		// 计算DI值
		if smoothTR > 0 {
			plusDI[i] = 100 * smoothPlusDM / smoothTR
			minusDI[i] = 100 * smoothMinusDM / smoothTR
		}

		// 计算DX值
		diSum := math.Abs(plusDI[i] - minusDI[i])
		diDiff := plusDI[i] + minusDI[i]
		if diDiff > 0 {
			adx[i] = 100 * diSum / diDiff
		}
	}

	// 平滑ADX值
	var smoothADX float64
	for i := period * 2; i < length; i++ {
		if i == period*2 {
			// 初始化平滑值
			for j := period; j <= i; j++ {
				smoothADX += adx[j]
			}
			adx[i] = smoothADX / float64(period+1)
		} else {
			// 计算平滑ADX
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
//   - period: 计算周期，通常使用14日
//
// 返回值：
//   - *TaADX: ADX值序列，+DI值序列，-DI值序列
//   - error: 可能的错误信息
func (k *KlineDatas) ADX(period int) (*TaADX, error) {
	return CalculateADX(*k, period)
}

// ADX_ 获取最新的ADX值
// 参数：
//   - period: 计算周期，通常使用14日
//
// 返回值：
//   - float64: 最新的ADX值，如果计算出错则返回-1
func (k *KlineDatas) ADX_(period int) (adx, plusDI, minusDI float64) {
	_k, err := k._Keep(period * 14)
	if err != nil {
		_k = *k
	}
	adxData, err := _k.ADX(period)
	if err != nil {
		return 0, 0, 0
	}
	return adxData.Value()
}

// Value 返回最新的ADX值
func (t *TaADX) Value() (adx, plusDI, minusDI float64) {
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex], t.PlusDI[lastIndex], t.MinusDI[lastIndex]
}

// IsTrendStrong 判断趋势是否强势
func (t *TaADX) IsTrendStrong(threshold ...float64) bool {
	th := 25.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.ADX[len(t.ADX)-1] > th
}

// IsTrendWeak 判断趋势是否弱势
func (t *TaADX) IsTrendWeak(threshold ...float64) bool {
	th := 20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.ADX[len(t.ADX)-1] < th
}

// IsTrendStrengthening 判断趋势是否在增强
func (t *TaADX) IsTrendStrengthening() bool {
	if len(t.ADX) < 2 {
		return false
	}
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex] > t.ADX[lastIndex-1]
}

// IsTrendWeakening 判断趋势是否在减弱
func (t *TaADX) IsTrendWeakening() bool {
	if len(t.ADX) < 2 {
		return false
	}
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex] < t.ADX[lastIndex-1]
}

// GetTrend 获取趋势方向
func (t *TaADX) GetTrend() int {
	lastIndex := len(t.ADX) - 1
	if t.PlusDI[lastIndex] > t.MinusDI[lastIndex] {
		return 1 // 上涨趋势
	} else if t.PlusDI[lastIndex] < t.MinusDI[lastIndex] {
		return -1 // 下跌趋势
	}
	return 0 // 无明显趋势
}

// IsBullishCrossover 判断是否发生多头趋势转换
func (t *TaADX) IsBullishCrossover() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex-1] <= t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] > t.MinusDI[lastIndex]
}

// IsBearishCrossover 判断是否发生空头趋势转换
func (t *TaADX) IsBearishCrossover() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex-1] >= t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] < t.MinusDI[lastIndex]
}

// GetDISpread 获取DI差值
func (t *TaADX) GetDISpread() float64 {
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex] - t.MinusDI[lastIndex]
}

// IsDIConverging 判断DI是否收敛
func (t *TaADX) IsDIConverging() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	currentSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	previousSpread := math.Abs(t.PlusDI[lastIndex-1] - t.MinusDI[lastIndex-1])
	return currentSpread < previousSpread
}

// IsDIDiverging 判断DI是否发散
func (t *TaADX) IsDIDiverging() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	currentSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	previousSpread := math.Abs(t.PlusDI[lastIndex-1] - t.MinusDI[lastIndex-1])
	return currentSpread > previousSpread
}

// GetTrendStrength 获取趋势强度
func (t *TaADX) GetTrendStrength() float64 {
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex]
}

// IsExtremeTrend 判断是否为极端趋势
func (t *TaADX) IsExtremeTrend(threshold ...float64) bool {
	th := 50.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.ADX[len(t.ADX)-1] > th
}

// GetTrendQuality 获取趋势质量
func (t *TaADX) GetTrendQuality() float64 {
	lastIndex := len(t.ADX) - 1
	diSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	return t.ADX[lastIndex] * diSpread / 100
}
