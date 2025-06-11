package ta

import (
	"fmt"
	"math"
)

type TaATR struct {
	Values    []float64 `json:"values"`
	Period    int       `json:"period"`
	TrueRange []float64 `json:"true_range"`
}

func CalculateATR(klineData KlineDatas, period int) (*TaATR, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)

	slices := preallocateSlices(length, 2)
	trueRange, atr := slices[0], slices[1]

	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevClose := klineData[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	var sumTR float64
	for i := 1; i <= period; i++ {
		sumTR += trueRange[i]
	}
	atr[period] = sumTR / float64(period)

	for i := period + 1; i < length; i++ {
		atr[i] = (atr[i-1]*(float64(period)-1) + trueRange[i]) / float64(period)
	}

	return &TaATR{
		Values:    atr,
		Period:    period,
		TrueRange: trueRange,
	}, nil
}

func (k *KlineDatas) ATR(period int) (*TaATR, error) {
	return CalculateATR(*k, period)
}

func (k *KlineDatas) ATR_(period int) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	atr, err := CalculateATR(_k, period)
	if err != nil {
		return 0
	}
	return atr.Value()
}

func (t *TaATR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaATR) GetTrueRange() float64 {
	return t.TrueRange[len(t.TrueRange)-1]
}

func (t *TaATR) IsVolatilityHigh(threshold float64) bool {
	if len(t.Values) < t.Period {
		return false
	}
	lastIndex := len(t.Values) - 1
	avgATR := 0.0
	for i := 0; i < t.Period; i++ {
		avgATR += t.Values[lastIndex-i]
	}
	avgATR /= float64(t.Period)
	return t.Values[lastIndex] > avgATR*threshold
}

func (t *TaATR) IsVolatilityLow(threshold float64) bool {
	if len(t.Values) < t.Period {
		return false
	}
	lastIndex := len(t.Values) - 1
	avgATR := 0.0
	for i := 0; i < t.Period; i++ {
		avgATR += t.Values[lastIndex-i]
	}
	avgATR /= float64(t.Period)
	return t.Values[lastIndex] < avgATR*threshold
}

func (t *TaATR) IsVolatilityIncreasing() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

func (t *TaATR) IsVolatilityDecreasing() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

func (t *TaATR) GetVolatilityChange() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return (t.Values[lastIndex] - t.Values[lastIndex-1]) / t.Values[lastIndex-1] * 100
}

func (t *TaATR) GetStopLoss(currentPrice float64, multiplier float64) (stopLoss float64) {
	atr := t.Value()
	return currentPrice - atr*multiplier
}

func (t *TaATR) GetTakeProfit(currentPrice float64, multiplier float64) (takeProfit float64) {
	atr := t.Value()
	return currentPrice + atr*multiplier
}

func (t *TaATR) GetChannelBounds(currentPrice float64, multiplier float64) (upper, lower float64) {
	atr := t.Value()
	upper = currentPrice + atr*multiplier
	lower = currentPrice - atr*multiplier
	return
}

func (t *TaATR) IsBreakingOut(price, prevPrice float64) bool {
	if len(t.Values) < 1 {
		return false
	}
	atr := t.Value()
	priceChange := math.Abs(price - prevPrice)
	return priceChange > atr
}

func (t *TaATR) GetVolatilityRatio() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	currentATR := t.Values[lastIndex]
	var sumATR float64
	for i := 0; i < t.Period; i++ {
		sumATR += t.Values[lastIndex-i]
	}
	avgATR := sumATR / float64(t.Period)
	return currentATR / avgATR
}
