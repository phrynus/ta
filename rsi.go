package ta

import (
	"fmt"
	"math"
)

type TaRSI struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
	Gains  []float64 `json:"gains"`
	Losses []float64 `json:"losses"`
}

func CalculateRSI(prices []float64, period int) (*TaRSI, error) {
	if len(prices) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(prices)

	slices := preallocateSlices(length, 3)
	rsi, gains, losses := slices[0], slices[1], slices[2]

	for i := 1; i < length; i++ {
		change := prices[i] - prices[i-1]
		gains[i] = math.Max(0, change)
		losses[i] = math.Max(0, -change)
	}

	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	for i := period; i < length; i++ {
		if i > period {
			avgGain = (avgGain*(float64(period)-1) + gains[i]) / float64(period)
			avgLoss = (avgLoss*(float64(period)-1) + losses[i]) / float64(period)
		}

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return &TaRSI{
		Values: rsi,
		Period: period,
		Gains:  gains,
		Losses: losses,
	}, nil
}

func (k *KlineDatas) RSI(period int, source string) (*TaRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateRSI(prices, period)
}

func (k *KlineDatas) RSI_(period int, source string) float64 {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0
	}
	rsi, err := CalculateRSI(prices, period)
	if err != nil {
		return 0
	}
	return rsi.Value()
}

func (t *TaRSI) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaRSI) IsOverbought(threshold ...float64) bool {
	th := 70.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() > th
}

func (t *TaRSI) IsOversold(threshold ...float64) bool {
	th := 30.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() < th
}

func (t *TaRSI) IsBullishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	rsiLow := t.Values[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] < rsiLow {
			rsiLow = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] > rsiLow && prices[lastIndex] < priceLow
}

func (t *TaRSI) IsBearishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	rsiHigh := t.Values[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] > rsiHigh {
			rsiHigh = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] < rsiHigh && prices[lastIndex] > priceHigh
}

func (t *TaRSI) IsCenterCross() (up, down bool) {
	if len(t.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	up = t.Values[lastIndex-1] <= 50 && t.Values[lastIndex] > 50
	down = t.Values[lastIndex-1] >= 50 && t.Values[lastIndex] < 50
	return
}

func (t *TaRSI) GetTrend() int {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	if t.Values[lastIndex] > 70 {
		return 1
	} else if t.Values[lastIndex] > 50 {
		return 2
	} else if t.Values[lastIndex] < 30 {
		return -1
	} else if t.Values[lastIndex] < 50 {
		return -2
	}
	return 0
}

func (t *TaRSI) GetStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex] - 50)
}

func (t *TaRSI) IsStrengthening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]-50) > math.Abs(t.Values[lastIndex-1]-50)
}

func (t *TaRSI) IsWeakening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]-50) < math.Abs(t.Values[lastIndex-1]-50)
}

func (t *TaRSI) GetGainLossRatio() float64 {
	lastIndex := len(t.Values) - 1
	if t.Losses[lastIndex] == 0 {
		return math.Inf(1)
	}
	return t.Gains[lastIndex] / t.Losses[lastIndex]
}

func (t *TaRSI) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	rsiChange := (t.Values[lastIndex] - t.Values[lastIndex-1]) / t.Values[lastIndex-1] * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(rsiChange-priceChange) > threshold
}
