package ta

import (
	"fmt"
)

type TaT3 struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
	VFact  float64   `json:"vfact"`
}

func CalculateT3(prices []float64, period int, vfact float64) (*TaT3, error) {
	if len(prices) < period*6 {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 7)
	ema1, ema2, ema3, ema4, ema5, ema6, t3 := slices[0], slices[1], slices[2], slices[3], slices[4], slices[5], slices[6]

	k := 2.0 / float64(period+1)
	ema1[0] = prices[0]
	for i := 1; i < length; i++ {
		ema1[i] = prices[i]*k + ema1[i-1]*(1-k)
	}

	for i := 1; i < length; i++ {
		ema2[i] = ema1[i]*k + ema2[i-1]*(1-k)
		ema3[i] = ema2[i]*k + ema3[i-1]*(1-k)
		ema4[i] = ema3[i]*k + ema4[i-1]*(1-k)
		ema5[i] = ema4[i]*k + ema5[i-1]*(1-k)
		ema6[i] = ema5[i]*k + ema6[i-1]*(1-k)
	}

	b := vfact
	c1 := -b * b * b
	c2 := 3*b*b + 3*b*b*b
	c3 := -6*b*b - 3*b - 3*b*b*b
	c4 := 1 + 3*b + b*b*b + 3*b*b

	for i := period * 6; i < length; i++ {
		t3[i] = c1*ema6[i] + c2*ema5[i] + c3*ema4[i] + c4*ema3[i]
	}

	return &TaT3{
		Values: t3,
		Period: period,
		VFact:  vfact,
	}, nil
}

func (k *KlineDatas) T3(period int, vfact float64, source string) (*TaT3, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateT3(prices, period, vfact)
}

func (k *KlineDatas) T3_(period int, vfact float64, source string) float64 {
	_k, err := k.Keep(period * 10)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	t3, err := CalculateT3(prices, period, vfact)
	if err != nil {
		return 0
	}
	return t3.Value()
}

func (t *TaT3) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaT3) IsCrossOver(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] <= price && t.Values[lastIndex] > price
}

func (t *TaT3) IsCrossUnder(price float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] >= price && t.Values[lastIndex] < price
}

func (t *TaT3) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

func (t *TaT3) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

func (t *TaT3) GetSlope() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

func (t *TaT3) IsCrossOverT3(other *TaT3) (bool, bool) {
	if len(t.Values) < 2 || len(other.Values) < 2 {
		return false, false
	}
	lastT3 := t.Values[len(t.Values)-1]
	prevT3 := t.Values[len(t.Values)-2]
	lastOther := other.Values[len(other.Values)-1]
	prevOther := other.Values[len(other.Values)-2]

	return lastT3 > lastOther && prevT3 <= prevOther, lastT3 < lastOther && prevT3 >= prevOther
}

func (t *TaT3) IsAccelerating() bool {
	if len(t.Values) < 4 {
		return false
	}
	lastIndex := len(t.Values) - 1
	diff1 := t.Values[lastIndex] - t.Values[lastIndex-1]
	diff2 := t.Values[lastIndex-1] - t.Values[lastIndex-2]
	diff3 := t.Values[lastIndex-2] - t.Values[lastIndex-3]
	return (diff1 > diff2 && diff2 > diff3) || (diff1 < diff2 && diff2 < diff3)
}

func (t *TaT3) GetTrendStrength() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	diff := t.Values[lastIndex] - t.Values[0]
	return (diff / t.Values[0]) * 100
}

func (t *TaT3) GetDeviation(price float64) float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastValue := t.Values[len(t.Values)-1]
	diff := lastValue - price
	return (diff / price) * 100
}

func (t *TaT3) IsOverbought(price float64, threshold float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	deviation := t.GetDeviation(price)
	return deviation > threshold
}

func (t *TaT3) IsOversold(price float64, threshold float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	deviation := t.GetDeviation(price)
	return deviation < -threshold
}

func (t *TaT3) GetVolumeFactorEffect() float64 {
	if len(t.Values) < 2 {
		return 0
	}

	return t.VFact * 100
}

func (t *TaT3) GetOptimalPeriod(prices []float64) int {
	if len(prices) < 2 {
		return 0
	}
	bestPeriod := 0
	bestSlope := 0.0
	for period := 5; period <= 20; period++ {
		t3, _ := CalculateT3(prices, period, t.VFact)
		slope := t3.GetSlope()
		if slope > bestSlope {
			bestSlope = slope
			bestPeriod = period
		}
	}
	return bestPeriod
}

func (t *TaT3) GetOptimalVolumeFactor(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}
	bestVFact := 0.0
	bestSlope := 0.0
	for vfact := 0.1; vfact <= 1.0; vfact += 0.1 {
		t3, _ := CalculateT3(prices, t.Period, vfact)
		slope := t3.GetSlope()
		if slope > bestSlope {
			bestSlope = slope
			bestVFact = vfact
		}
	}
	return bestVFact
}
