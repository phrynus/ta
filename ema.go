package ta

import (
	"fmt"
)

type TaEMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateEMA(prices []float64, period int) (*TaEMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	result := slices[0]

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	multiplier := 2.0 / float64(period+1)
	oneMinusMultiplier := 1.0 - multiplier

	for i := period; i < length; i++ {
		result[i] = prices[i]*multiplier + result[i-1]*oneMinusMultiplier
	}

	return &TaEMA{
		Values: result,
		Period: period,
	}, nil
}

func (k *KlineDatas) EMA(period int, source string) (*TaEMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateEMA(prices, period)
}

func (k *KlineDatas) EMA_(period int, source string) float64 {

	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	ema, err := CalculateEMA(prices, period)
	if err != nil {
		return 0
	}
	return ema.Value()
}

func (t *TaEMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaEMA) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

func (t *TaEMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

func (t *TaEMA) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

func (t *TaEMA) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

func (t *TaEMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

func (t *TaEMA) IsCrossOverEMA(other *TaEMA) (golden, death bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	golden = t.Values[lastIndex-1] <= other.Values[lastIndex-1] && t.Values[lastIndex] > other.Values[lastIndex]
	death = t.Values[lastIndex-1] >= other.Values[lastIndex-1] && t.Values[lastIndex] < other.Values[lastIndex]
	return
}

func (t *TaEMA) IsAccelerating() bool {
	if len(t.Values) < 3 {
		return false
	}
	lastIndex := len(t.Values) - 1
	diff1 := t.Values[lastIndex] - t.Values[lastIndex-1]
	diff2 := t.Values[lastIndex-1] - t.Values[lastIndex-2]
	return (diff1 > 0 && diff1 > diff2) || (diff1 < 0 && diff1 < diff2)
}

func (t *TaEMA) GetTrendStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	startIndex := lastIndex - t.Period + 1
	startValue := t.Values[startIndex]
	endValue := t.Values[lastIndex]
	return (endValue - startValue) / startValue * 100
}
