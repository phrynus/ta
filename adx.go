package ta

import (
	"fmt"
	"math"
)

type TaADX struct {
	ADX     []float64 `json:"adx"`
	PlusDI  []float64 `json:"plus_di"`
	MinusDI []float64 `json:"minus_di"`
	Period  int       `json:"period"`
}

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

func (k *KlineDatas) ADX(period int) (*TaADX, error) {
	return CalculateADX(*k, period)
}

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

func (t *TaADX) Value() (adx, plusDI, minusDI float64) {
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex], t.PlusDI[lastIndex], t.MinusDI[lastIndex]
}

func (t *TaADX) IsTrendStrong(threshold ...float64) bool {
	th := 25.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.ADX[len(t.ADX)-1] > th
}

func (t *TaADX) IsTrendWeak(threshold ...float64) bool {
	th := 20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.ADX[len(t.ADX)-1] < th
}

func (t *TaADX) IsTrendStrengthening() bool {
	if len(t.ADX) < 2 {
		return false
	}
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex] > t.ADX[lastIndex-1]
}

func (t *TaADX) IsTrendWeakening() bool {
	if len(t.ADX) < 2 {
		return false
	}
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex] < t.ADX[lastIndex-1]
}

func (t *TaADX) GetTrend() int {
	lastIndex := len(t.ADX) - 1
	if t.PlusDI[lastIndex] > t.MinusDI[lastIndex] {
		return 1
	} else if t.PlusDI[lastIndex] < t.MinusDI[lastIndex] {
		return -1
	}
	return 0
}

func (t *TaADX) IsBullishCrossover() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex-1] <= t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] > t.MinusDI[lastIndex]
}

func (t *TaADX) IsBearishCrossover() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex-1] >= t.MinusDI[lastIndex-1] && t.PlusDI[lastIndex] < t.MinusDI[lastIndex]
}

func (t *TaADX) GetDISpread() float64 {
	lastIndex := len(t.PlusDI) - 1
	return t.PlusDI[lastIndex] - t.MinusDI[lastIndex]
}

func (t *TaADX) IsDIConverging() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	currentSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	previousSpread := math.Abs(t.PlusDI[lastIndex-1] - t.MinusDI[lastIndex-1])
	return currentSpread < previousSpread
}

func (t *TaADX) IsDIDiverging() bool {
	if len(t.PlusDI) < 2 || len(t.MinusDI) < 2 {
		return false
	}
	lastIndex := len(t.PlusDI) - 1
	currentSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	previousSpread := math.Abs(t.PlusDI[lastIndex-1] - t.MinusDI[lastIndex-1])
	return currentSpread > previousSpread
}

func (t *TaADX) GetTrendStrength() float64 {
	lastIndex := len(t.ADX) - 1
	return t.ADX[lastIndex]
}

func (t *TaADX) IsExtremeTrend(threshold ...float64) bool {
	th := 50.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.ADX[len(t.ADX)-1] > th
}

func (t *TaADX) GetTrendQuality() float64 {
	lastIndex := len(t.ADX) - 1
	diSpread := math.Abs(t.PlusDI[lastIndex] - t.MinusDI[lastIndex])
	return t.ADX[lastIndex] * diSpread / 100
}
