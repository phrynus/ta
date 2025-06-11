package ta

import (
	"fmt"
)

type TaOBV struct {
	Values []float64 `json:"values"`
}

func CalculateOBV(prices, volumes []float64) (*TaOBV, error) {
	if len(prices) != len(volumes) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(prices) < 2 {
		return nil, fmt.Errorf("计算数据不足")
	}

	obv := make([]float64, len(prices))
	obv[0] = volumes[0]

	for i := 1; i < len(prices); i++ {
		if prices[i] > prices[i-1] {
			obv[i] = obv[i-1] + volumes[i]
		} else if prices[i] < prices[i-1] {
			obv[i] = obv[i-1] - volumes[i]
		} else {
			obv[i] = obv[i-1]
		}
	}

	return &TaOBV{
		Values: obv,
	}, nil
}

func (k *KlineDatas) OBV(source string) (*TaOBV, error) {
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateOBV(close, volume)
}

func (k *KlineDatas) OBV_(source string) float64 {

	_k, err := k.Keep(50)
	if err != nil {
		_k = *k
	}
	obv, err := _k.OBV(source)
	if err != nil {
		return 0
	}
	return obv.Value()
}

func (t *TaOBV) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaOBV) IsTrendUp() bool {
	if len(t.Values) < 5 {
		return false
	}
	lastIndex := len(t.Values) - 1

	var sumX, sumY, sumXY, sumX2 float64
	n := 5
	for i := 0; i < n; i++ {
		x := float64(i)
		y := t.Values[lastIndex-i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	return slope < 0
}

func (t *TaOBV) IsTrendDown() bool {
	if len(t.Values) < 5 {
		return false
	}
	lastIndex := len(t.Values) - 1

	var sumX, sumY, sumXY, sumX2 float64
	n := 5
	for i := 0; i < n; i++ {
		x := float64(i)
		y := t.Values[lastIndex-i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	return slope > 0
}

func (t *TaOBV) IsBullishDivergence(prices []float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	previousLow := prices[lastIndex-1]
	for i := 2; i < len(prices); i++ {
		if prices[i] < previousLow && t.Values[i] > t.Values[i-1] {
			return true
		}
	}
	return false
}

func (t *TaOBV) IsBearishDivergence(prices []float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	previousHigh := prices[lastIndex-1]
	for i := 2; i < len(prices); i++ {
		if prices[i] > previousHigh && t.Values[i] < t.Values[i-1] {
			return true
		}
	}
	return false
}

func (t *TaOBV) GetTrendStrength() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	previousIndex := len(t.Values) - 2
	return t.Values[lastIndex] - t.Values[previousIndex]
}

func (t *TaOBV) IsBreakout(level float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > level
}

func (t *TaOBV) GetAccumulation() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaOBV) IsVolumeExpanding() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > t.Values[len(t.Values)-2]
}

func (t *TaOBV) IsVolumeContracting() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] < t.Values[len(t.Values)-2]
}

func (t *TaOBV) GetVolumeForce() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	return t.Values[len(t.Values)-1] - t.Values[len(t.Values)-2]
}

func (t *TaOBV) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	previousIndex := len(prices) - 2
	obvChange := t.Values[lastIndex] - t.Values[previousIndex]
	priceChange := prices[lastIndex] - prices[previousIndex]
	return obvChange > threshold*priceChange
}
