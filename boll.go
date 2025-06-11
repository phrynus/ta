package ta

import (
	"fmt"
	"math"
)

type TaBoll struct {
	Upper []float64 `json:"upper"`
	Mid   []float64 `json:"mid"`
	Lower []float64 `json:"lower"`
}

func CalculateBoll(prices []float64, period int, stdDev float64) (*TaBoll, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 3)
	upper, mid, lower := slices[0], slices[1], slices[2]

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	mid[period-1] = sum / float64(period)

	for i := period; i < length; i++ {
		sum = sum - prices[i-period] + prices[i]
		mid[i] = sum / float64(period)
	}

	for i := period - 1; i < length; i++ {

		var sumSquares float64
		for j := 0; j < period; j++ {
			diff := prices[i-j] - mid[i]
			sumSquares += diff * diff
		}
		sd := math.Sqrt(sumSquares / float64(period))

		band := sd * stdDev
		upper[i] = mid[i] + band
		lower[i] = mid[i] - band
	}

	return &TaBoll{
		Upper: upper,
		Mid:   mid,
		Lower: lower,
	}, nil
}

func (k *KlineDatas) Boll(period int, stdDev float64, source string) (*TaBoll, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateBoll(prices, period, stdDev)
}

func (k *KlineDatas) Boll_(period int, stdDev float64, source string) (upper, mid, lower float64) {

	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0, 0
	}
	boll, err := CalculateBoll(prices, period, stdDev)
	if err != nil {
		return 0, 0, 0
	}
	return boll.Value()
}

func (t *TaBoll) Value() (upper, mid, lower float64) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Mid[lastIndex], t.Lower[lastIndex]
}

func (t *TaBoll) IsBollCross(prices []float64) (upperCross, lowerCross bool) {
	if len(prices) < 2 || len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false, false
	}
	lastIndex := len(prices) - 1
	upperCross = prices[lastIndex-1] <= t.Upper[lastIndex-1] && prices[lastIndex] > t.Upper[lastIndex]
	lowerCross = prices[lastIndex-1] >= t.Lower[lastIndex-1] && prices[lastIndex] < t.Lower[lastIndex]
	return
}

func (t *TaBoll) IsSqueezing() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentBandwidth < previousBandwidth
}

func (t *TaBoll) IsExpanding() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentBandwidth > previousBandwidth
}

func (t *TaBoll) GetBandwidth() float64 {
	lastIndex := len(t.Upper) - 1
	return (t.Upper[lastIndex] - t.Lower[lastIndex]) / t.Mid[lastIndex] * 100
}

func (t *TaBoll) IsOverBought(price float64) bool {
	lastIndex := len(t.Upper) - 1
	return price > t.Upper[lastIndex]
}

func (t *TaBoll) IsOverSold(price float64) bool {
	lastIndex := len(t.Lower) - 1
	return price < t.Lower[lastIndex]
}

func (t *TaBoll) IsTrendUp() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] > t.Mid[lastIndex-1]
}

func (t *TaBoll) IsTrendDown() bool {
	if len(t.Mid) < 2 {
		return false
	}
	lastIndex := len(t.Mid) - 1
	return t.Mid[lastIndex] < t.Mid[lastIndex-1]
}

func (t *TaBoll) IsBreakoutPossible() bool {
	if len(t.Upper) < 20 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	bandwidth := t.GetBandwidth()
	previousBandwidth := (t.Upper[lastIndex-1] - t.Lower[lastIndex-1]) / t.Mid[lastIndex-1] * 100
	return bandwidth < 2.5 && bandwidth < previousBandwidth
}

func (t *TaBoll) GetPercentB(price float64) float64 {
	lastIndex := len(t.Upper) - 1
	return (price - t.Lower[lastIndex]) / (t.Upper[lastIndex] - t.Lower[lastIndex]) * 100
}
