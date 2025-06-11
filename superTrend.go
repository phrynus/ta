package ta

import (
	"fmt"
	"math"
)

type TaSuperTrend struct {
	Upper      []float64 `json:"upper"`
	Lower      []float64 `json:"lower"`
	Trend      []bool    `json:"trend"`
	Period     int       `json:"period"`
	Multiplier float64   `json:"multiplier"`
}

func CalculateSuperTrend(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrend, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	atr, err := klineData.ATR(period)
	if err != nil {
		return nil, err
	}

	length := len(klineData)

	slices := preallocateSlices(length, 2)
	upperBand, lowerBand := slices[0], slices[1]
	trend := make([]bool, length)

	for i := period; i < length; i++ {
		midpoint := (klineData[i].High + klineData[i].Low) / 2
		atrValue := atr.Values[i]
		upperBand[i] = midpoint + multiplier*atrValue
		lowerBand[i] = midpoint - multiplier*atrValue
	}

	trend[period] = klineData[period].Close > lowerBand[period]

	for i := period + 1; i < length; i++ {
		if trend[i-1] {
			if klineData[i].Close < lowerBand[i] {
				trend[i] = false
				upperBand[i] = upperBand[i-1]
			} else {
				trend[i] = true
				lowerBand[i] = math.Max(lowerBand[i], lowerBand[i-1])
			}
		} else {
			if klineData[i].Close > upperBand[i] {
				trend[i] = true
				lowerBand[i] = lowerBand[i-1]
			} else {
				trend[i] = false
				upperBand[i] = math.Min(upperBand[i], upperBand[i-1])
			}
		}
	}

	return &TaSuperTrend{
		Upper:      upperBand,
		Lower:      lowerBand,
		Trend:      trend,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

func (k *KlineDatas) SuperTrend(period int, multiplier float64) (*TaSuperTrend, error) {
	return CalculateSuperTrend(*k, period, multiplier)
}

func (k *KlineDatas) SuperTrend_(period int, multiplier float64) (upper, lower float64, isUpTrend bool) {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	st, err := _k.SuperTrend(period, multiplier)
	if err != nil {
		return 0, 0, false
	}
	lastIndex := len(st.Upper) - 1
	return st.Upper[lastIndex], st.Lower[lastIndex], st.Trend[lastIndex]
}

func (t *TaSuperTrend) Value() (upper, lower float64, isUpTrend bool) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

func (t *TaSuperTrend) IsTrendChange() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex] != t.Trend[lastIndex-1]
}

func (t *TaSuperTrend) IsBullishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return !t.Trend[lastIndex-1] && t.Trend[lastIndex]
}

func (t *TaSuperTrend) IsBearishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex-1] && !t.Trend[lastIndex]
}

func (t *TaSuperTrend) IsUp() bool {
	return t.Trend[len(t.Trend)-1]
}

func (t *TaSuperTrend) IsDown() bool {
	return !t.Trend[len(t.Trend)-1]
}

func (t *TaSuperTrend) GetUpper() float64 {
	return t.Upper[len(t.Upper)-1]
}

func (t *TaSuperTrend) GetLower() float64 {
	return t.Lower[len(t.Lower)-1]
}

func (t *TaSuperTrend) GetTrendStrength() float64 {
	lastIndex := len(t.Upper) - 1
	if t.Trend[lastIndex] {
		return t.Upper[lastIndex] - t.Lower[lastIndex]
	}
	return t.Lower[lastIndex] - t.Upper[lastIndex]
}

func (t *TaSuperTrend) IsTrendStrengthening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength > previousStrength
}

func (t *TaSuperTrend) IsTrendWeakening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength < previousStrength
}

func (t *TaSuperTrend) GetTrendDuration() int {
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

func (t *TaSuperTrend) GetBandwidth() float64 {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex] - t.Lower[lastIndex]
}

func (t *TaSuperTrend) IsBreakoutPossible(threshold ...float64) bool {
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

func (t *TaSuperTrend) GetTrendQuality() float64 {
	duration := t.GetTrendDuration()
	strength := t.GetTrendStrength()
	return float64(duration) * strength
}
