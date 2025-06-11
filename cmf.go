package ta

import (
	"fmt"
	"math"
)

type TaCMF struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateCMF(high, low, close, volume []float64, period int) (*TaCMF, error) {
	if len(high) != len(low) || len(high) != len(close) || len(high) != len(volume) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(high) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(high)
	mfv := make([]float64, length)
	cmf := make([]float64, length)

	for i := 0; i < length; i++ {
		if high[i] == low[i] {
			mfv[i] = 0
		} else {

			mfm := ((close[i] - low[i]) - (high[i] - close[i])) / (high[i] - low[i])
			mfv[i] = mfm * volume[i]
		}
	}

	for i := period - 1; i < length; i++ {
		sumMFV := 0.0
		sumVolume := 0.0
		for j := 0; j < period; j++ {
			sumMFV += mfv[i-j]
			sumVolume += volume[i-j]
		}
		if sumVolume != 0 {
			cmf[i] = sumMFV / sumVolume
		}
	}

	return &TaCMF{
		Values: cmf,
		Period: period,
	}, nil
}

func (k *KlineDatas) CMF(period int, source string) (*TaCMF, error) {
	high, err := k.ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k.ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateCMF(high, low, close, volume, period)
}

func (t *TaCMF) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaCMF) IsPositive() bool {
	return t.Values[len(t.Values)-1] > 0
}

func (t *TaCMF) IsNegative() bool {
	return t.Values[len(t.Values)-1] < 0
}

func (t *TaCMF) IsBullishDivergence(prices []float64) bool {
	if len(prices) < 20 {
		return false
	}
	lastIndex := len(prices) - 1
	previousLow := prices[lastIndex-1]
	for i := 2; i < 20; i++ {
		if prices[lastIndex-i] < previousLow {
			previousLow = prices[lastIndex-i]
		}
	}
	return prices[lastIndex] > previousLow && t.Values[lastIndex] < 0
}

func (t *TaCMF) IsBearishDivergence(prices []float64) bool {
	if len(prices) < 20 {
		return false
	}
	lastIndex := len(prices) - 1
	previousHigh := prices[lastIndex-1]
	for i := 2; i < 20; i++ {
		if prices[lastIndex-i] > previousHigh {
			previousHigh = prices[lastIndex-i]
		}
	}
	return prices[lastIndex] < previousHigh && t.Values[lastIndex] > 0
}

func (k *KlineDatas) CMF_(period int, source string) float64 {

	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	cmf, err := _k.CMF(period, source)
	if err != nil {
		return 0
	}
	return cmf.Value()
}

func (t *TaCMF) GetStrength() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaCMF) IsCrossZero() (bool, bool) {
	lastValue := t.Values[len(t.Values)-1]
	return lastValue > 0, lastValue < 0
}

func (t *TaCMF) GetAccumulation() float64 {
	sum := 0.0
	for _, value := range t.Values {
		sum += value
	}
	return sum
}

func (t *TaCMF) IsStrengthening() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > t.Values[len(t.Values)-2]
}

func (t *TaCMF) IsWeakening() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] < t.Values[len(t.Values)-2]
}

func (t *TaCMF) GetMoneyFlowZone() string {
	lastValue := t.Values[len(t.Values)-1]
	if lastValue > 0 {
		return "strong_inflow"
	} else if lastValue > -0.1 && lastValue < 0.1 {
		return "neutral"
	} else {
		return "strong_outflow"
	}
}

func (t *TaCMF) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(prices) < 2 {
		return false
	}
	lastIndex := len(prices) - 1
	lastCMF := t.Values[lastIndex]
	previousCMF := t.Values[lastIndex-1]
	priceChange := prices[lastIndex] - prices[lastIndex-1]
	cmfChange := lastCMF - previousCMF
	return math.Abs(cmfChange/priceChange) > threshold
}

func (t *TaCMF) GetOptimalPeriod(prices []float64) int {
	if len(prices) < 2 {
		return 1
	}
	period := 1
	for i := 1; i < len(prices); i++ {
		if prices[i] != prices[i-1] {
			period = i + 1
		}
	}
	return period
}
