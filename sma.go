package ta

import (
	"fmt"
)

type TaSMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateSMA(prices []float64, period int) (*TaSMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	sma := slices[0]

	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	sma[period-1] = sum / float64(period)

	for i := period; i < length; i++ {
		sum += prices[i] - prices[i-period]
		sma[i] = sum / float64(period)
	}

	return &TaSMA{
		Values: sma,
		Period: period,
	}, nil
}

func (k *KlineDatas) SMA(period int, source string) (*TaSMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateSMA(prices, period)
}

func (k *KlineDatas) SMA_(period int, source string) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	sma, err := CalculateSMA(prices, period)
	if err != nil {
		return 0
	}
	return sma.Value()
}

func (t *TaSMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaSMA) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

func (t *TaSMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

func (t *TaSMA) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

func (t *TaSMA) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

func (t *TaSMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

func (t *TaSMA) IsCrossOverSMA(other *TaSMA) (golden, death bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	golden = t.Values[lastIndex-1] <= other.Values[lastIndex-1] && t.Values[lastIndex] > other.Values[lastIndex]
	death = t.Values[lastIndex-1] >= other.Values[lastIndex-1] && t.Values[lastIndex] < other.Values[lastIndex]
	return
}

func (t *TaSMA) IsAccelerating() bool {
	if len(t.Values) < 3 {
		return false
	}
	lastIndex := len(t.Values) - 1
	diff1 := t.Values[lastIndex] - t.Values[lastIndex-1]
	diff2 := t.Values[lastIndex-1] - t.Values[lastIndex-2]
	return (diff1 > 0 && diff1 > diff2) || (diff1 < 0 && diff1 < diff2)
}

func (t *TaSMA) GetTrendStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	startIndex := lastIndex - t.Period + 1
	startValue := t.Values[startIndex]
	endValue := t.Values[lastIndex]
	return (endValue - startValue) / startValue * 100
}

func (t *TaSMA) GetDeviation(price float64) float64 {
	lastValue := t.Value()
	return (price - lastValue) / lastValue * 100
}

func (t *TaSMA) IsOverbought(price float64, threshold float64) bool {
	return t.GetDeviation(price) > threshold
}

func (t *TaSMA) IsOversold(price float64, threshold float64) bool {
	return t.GetDeviation(price) < -threshold
}
