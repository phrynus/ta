package ta

import (
	"fmt"
	"math"
)

type TaSuperTrendPivot struct {
	Upper       []float64 `json:"upper"`
	Lower       []float64 `json:"lower"`
	Trend       []int     `json:"trend"`
	PivotPeriod int       `json:"pivot_period"`
	Factor      float64   `json:"factor"`
	AtrPeriod   int       `json:"atr_period"`
}

func FindPivotHighPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].High > klineData[index].High {
			return math.NaN()
		}
	}
	return klineData[index].High
}

func FindPivotLowPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].Low < klineData[index].Low {
			return math.NaN()
		}
	}
	return klineData[index].Low
}

func CalculateSuperTrendPivot(klineData KlineDatas, pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {

	dataLen := len(klineData)
	if dataLen < pivotPeriod || dataLen < atrPeriod {
		return nil, fmt.Errorf("计算数据不足: 数据长度%d, 需要轴点周期%d和ATR周期%d", dataLen, pivotPeriod, atrPeriod)
	}

	trendUp := make([]float64, dataLen)
	trendDown := make([]float64, dataLen)
	trend := make([]int, dataLen)

	atr, err := klineData.ATR(atrPeriod)
	if err != nil {
		return nil, fmt.Errorf("计算ATR失败: %v", err)
	}

	var center float64
	var centerCount int

	for i := pivotPeriod; i < dataLen; i++ {

		pivotHigh := FindPivotHighPoint(klineData, i, pivotPeriod)
		pivotLow := FindPivotLowPoint(klineData, i, pivotPeriod)

		if !math.IsNaN(pivotHigh) || !math.IsNaN(pivotLow) {
			newCenter := 0.0
			if !math.IsNaN(pivotHigh) && !math.IsNaN(pivotLow) {

				newCenter = (pivotHigh + pivotLow) / 2
			} else if !math.IsNaN(pivotHigh) {
				newCenter = pivotHigh
			} else {
				newCenter = pivotLow
			}

			if centerCount == 0 {
				center = newCenter
			} else {
				center = (center*2 + newCenter) / 3
			}
			centerCount++
		}

		if centerCount == 0 {
			center = (klineData[i].High + klineData[i].Low) / 2
		}

		band := factor * atr.Values[i]
		upperBand := center + band
		lowerBand := center - band

		if i > 0 {

			if klineData[i-1].Close > trendUp[i-1] {
				trendUp[i] = math.Max(lowerBand, trendUp[i-1])
			} else {
				trendUp[i] = lowerBand
			}

			if klineData[i-1].Close < trendDown[i-1] {
				trendDown[i] = math.Min(upperBand, trendDown[i-1])
			} else {
				trendDown[i] = upperBand
			}

			if klineData[i].Close > trendDown[i-1] {
				trend[i] = 1
			} else if klineData[i].Close < trendUp[i-1] {
				trend[i] = -1
			} else {
				trend[i] = trend[i-1]
			}
		} else {

			trendUp[i] = lowerBand
			trendDown[i] = upperBand
			trend[i] = 0
		}
	}

	return &TaSuperTrendPivot{
		Upper:       trendDown,
		Lower:       trendUp,
		Trend:       trend,
		PivotPeriod: pivotPeriod,
		Factor:      factor,
		AtrPeriod:   atrPeriod,
	}, nil
}

func (k *KlineDatas) SuperTrendPivot(pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {
	return CalculateSuperTrendPivot(*k, pivotPeriod, factor, atrPeriod)
}

func (k *KlineDatas) SuperTrendPivot_IsUp(pivotPeriod int, factor float64, atrPeriod int) bool {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return false
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return false
	}
	return superTrend.Trend[len(superTrend.Trend)-1] == 1
}

func (k *KlineDatas) SuperTrendPivot_IsDown(pivotPeriod int, factor float64, atrPeriod int) bool {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return false
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return false
	}
	return superTrend.Trend[len(superTrend.Trend)-1] == -1
}

func (k *KlineDatas) SuperTrendPivot_GetUpper(pivotPeriod int, factor float64, atrPeriod int) float64 {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return -1
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return -1
	}
	return superTrend.Upper[len(superTrend.Upper)-1]
}

func (k *KlineDatas) SuperTrendPivot_GetLower(pivotPeriod int, factor float64, atrPeriod int) float64 {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return -1
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return -1
	}
	return superTrend.Lower[len(superTrend.Lower)-1]
}

func (t *TaSuperTrendPivot) Value() (upper, lower float64, trend int) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

func (t *TaSuperTrendPivot) IsUp() bool {
	return t.Trend[len(t.Trend)-1] == 1
}

func (t *TaSuperTrendPivot) IsDown() bool {
	return t.Trend[len(t.Trend)-1] == -1
}

func (t *TaSuperTrendPivot) IsSideways() bool {
	return t.Trend[len(t.Trend)-1] == 0
}

func (t *TaSuperTrendPivot) GetUpper() float64 {
	return t.Upper[len(t.Upper)-1]
}

func (t *TaSuperTrendPivot) GetLower() float64 {
	return t.Lower[len(t.Lower)-1]
}

func (t *TaSuperTrendPivot) IsTrendChange() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex] != t.Trend[lastIndex-1]
}

func (t *TaSuperTrendPivot) IsBullishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex-1] <= 0 && t.Trend[lastIndex] == 1
}

func (t *TaSuperTrendPivot) IsBearishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex-1] >= 0 && t.Trend[lastIndex] == -1
}

func (t *TaSuperTrendPivot) GetTrendStrength() float64 {
	lastIndex := len(t.Upper) - 1
	if t.Trend[lastIndex] == 1 {
		return t.Upper[lastIndex] - t.Lower[lastIndex]
	} else if t.Trend[lastIndex] == -1 {
		return t.Lower[lastIndex] - t.Upper[lastIndex]
	}
	return 0
}

func (t *TaSuperTrendPivot) IsTrendStrengthening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength > previousStrength
}

func (t *TaSuperTrendPivot) IsTrendWeakening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength < previousStrength
}

func (t *TaSuperTrendPivot) GetTrendDuration() int {
	if len(t.Trend) < 2 {
		return 0
	}
	lastIndex := len(t.Trend) - 1
	currentTrend := t.Trend[lastIndex]
	duration := 1
	for i := lastIndex - 1; i >= 0; i-- {
		if t.Trend[i] != currentTrend {
			break
		}
		duration++
	}
	return duration
}

func (t *TaSuperTrendPivot) GetBandwidth() float64 {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex] - t.Lower[lastIndex]
}

func (t *TaSuperTrendPivot) IsBreakoutPossible(threshold ...float64) bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	th := 0.1
	if len(threshold) > 0 {
		th = threshold[0]
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return math.Abs(currentBandwidth-previousBandwidth)/previousBandwidth > th
}

func (t *TaSuperTrendPivot) GetTrendQuality() float64 {
	duration := t.GetTrendDuration()
	strength := t.GetTrendStrength()
	return float64(duration) * strength
}

func (t *TaSuperTrendPivot) IsPivotBreakout(klineData KlineDatas) bool {
	if len(t.Trend) < t.PivotPeriod {
		return false
	}
	lastIndex := len(t.Trend) - 1
	pivotHigh := FindPivotHighPoint(klineData, lastIndex, t.PivotPeriod)
	pivotLow := FindPivotLowPoint(klineData, lastIndex, t.PivotPeriod)

	if !math.IsNaN(pivotHigh) && klineData[lastIndex].Close > pivotHigh {
		return true
	}
	if !math.IsNaN(pivotLow) && klineData[lastIndex].Close < pivotLow {
		return true
	}
	return false
}

func (t *TaSuperTrendPivot) GetPivotStrength(klineData KlineDatas) float64 {
	if len(t.Trend) < t.PivotPeriod {
		return 0
	}
	lastIndex := len(t.Trend) - 1
	pivotHigh := FindPivotHighPoint(klineData, lastIndex, t.PivotPeriod)
	pivotLow := FindPivotLowPoint(klineData, lastIndex, t.PivotPeriod)

	if !math.IsNaN(pivotHigh) && !math.IsNaN(pivotLow) {
		return pivotHigh - pivotLow
	}
	return 0
}
