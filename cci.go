package ta

import (
	"fmt"
	"math"
)

type TaCCI struct {
	Values []float64 `json:"values"`
}

func CalculateCCI(klineData KlineDatas, period int) (*TaCCI, error) {
	if len(klineData) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)

	slices := preallocateSlices(length, 2)
	typicalPrice, cci := slices[0], slices[1]

	for i := 0; i < length; i++ {
		typicalPrice[i] = (klineData[i].High + klineData[i].Low + klineData[i].Close) / 3
	}

	for i := period - 1; i < length; i++ {

		var sumTP float64
		for j := i - period + 1; j <= i; j++ {
			sumTP += typicalPrice[j]
		}
		smaTP := sumTP / float64(period)

		var sumAbsDev float64
		for j := i - period + 1; j <= i; j++ {
			sumAbsDev += math.Abs(typicalPrice[j] - smaTP)
		}
		meanDeviation := sumAbsDev / float64(period)

		if meanDeviation != 0 {
			cci[i] = (typicalPrice[i] - smaTP) / (0.015 * meanDeviation)
		}
	}

	return &TaCCI{
		Values: cci,
	}, nil
}

func (k *KlineDatas) CCI(period int) (*TaCCI, error) {
	return CalculateCCI(*k, period)
}

func (k *KlineDatas) CCI_(period int) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	cci, err := CalculateCCI(_k, period)
	if err != nil {
		return 0
	}
	return cci.Value()
}

func (t *TaCCI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaCCI) IsOverbought() bool {
	return t.Value() > 100
}

func (t *TaCCI) IsOversold() bool {
	return t.Value() < -100
}

func (t *TaCCI) IsExtremeBought() bool {
	return t.Value() > 200
}

func (t *TaCCI) IsExtremeSold() bool {
	return t.Value() < -200
}

func (t *TaCCI) IsBullishDivergence() bool {
	if len(t.Values) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	previousLow := t.Values[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.Values[lastIndex-i] < previousLow {
			previousLow = t.Values[lastIndex-i]
		}
	}
	return t.Values[lastIndex] > previousLow && t.Values[lastIndex] < -100
}

func (t *TaCCI) IsBearishDivergence() bool {
	if len(t.Values) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	previousHigh := t.Values[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.Values[lastIndex-i] > previousHigh {
			previousHigh = t.Values[lastIndex-i]
		}
	}
	return t.Values[lastIndex] < previousHigh && t.Values[lastIndex] > 100
}
