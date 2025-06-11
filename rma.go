package ta

import (
	"fmt"
)

type TaRMA struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateRMA(prices []float64, period int) (*TaRMA, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 1)
	rma := slices[0]

	alpha := 1.0 / float64(period)
	rma[0] = prices[0]

	for i := 1; i < length; i++ {
		rma[i] = alpha*prices[i] + (1-alpha)*rma[i-1]
	}

	return &TaRMA{
		Values: rma,
		Period: period,
	}, nil
}

func (k *KlineDatas) RMA(period int, source string) (*TaRMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRMA(prices, period)
}

func (k *KlineDatas) RMA_(period int, source string) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	rma, err := CalculateRMA(prices, period)
	if err != nil {
		return 0
	}
	return rma.Value()
}

func (t *TaRMA) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaRMA) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

func (t *TaRMA) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

func (t *TaRMA) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

func (t *TaRMA) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

func (t *TaRMA) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

func (t *TaRMA) IsCrossOverRMA(other *TaRMA) (golden, death bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	golden = t.Values[lastIndex-1] <= other.Values[lastIndex-1] && t.Values[lastIndex] > other.Values[lastIndex]
	death = t.Values[lastIndex-1] >= other.Values[lastIndex-1] && t.Values[lastIndex] < other.Values[lastIndex]
	return
}

func (t *TaRMA) IsAccelerating() bool {
	if len(t.Values) < 3 {
		return false
	}
	lastIndex := len(t.Values) - 1
	diff1 := t.Values[lastIndex] - t.Values[lastIndex-1]
	diff2 := t.Values[lastIndex-1] - t.Values[lastIndex-2]
	return (diff1 > 0 && diff1 > diff2) || (diff1 < 0 && diff1 < diff2)
}

func (t *TaRMA) GetTrendStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	startIndex := lastIndex - t.Period + 1
	startValue := t.Values[startIndex]
	endValue := t.Values[lastIndex]
	return (endValue - startValue) / startValue * 100
}

func (t *TaRMA) GetDeviation(price float64) float64 {
	lastValue := t.Value()
	return (price - lastValue) / lastValue * 100
}

func (t *TaRMA) IsOverbought(price float64, threshold float64) bool {
	return t.GetDeviation(price) > threshold
}

func (t *TaRMA) IsOversold(price float64, threshold float64) bool {
	return t.GetDeviation(price) < -threshold
}

func (t *TaRMA) GetSmoothing() float64 {
	return 1.0 / float64(t.Period)
}
