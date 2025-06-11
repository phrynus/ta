package ta

import (
	"fmt"
	"math"
)

type TaWilliamsR struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateWilliamsR(high, low, close []float64, period int) (*TaWilliamsR, error) {
	if len(high) < period || len(low) < period || len(close) < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)

	slices := preallocateSlices(length, 1)
	wr := slices[0]

	for i := period - 1; i < length; i++ {

		var highestHigh, lowestLow = high[i], low[i]
		for j := 0; j < period; j++ {
			idx := i - j
			if high[idx] > highestHigh {
				highestHigh = high[idx]
			}
			if low[idx] < lowestLow {
				lowestLow = low[idx]
			}
		}

		if highestHigh != lowestLow {
			wr[i] = ((highestHigh - close[i]) / (highestHigh - lowestLow)) * -100
		} else {
			wr[i] = -50
		}
	}

	return &TaWilliamsR{
		Values: wr,
		Period: period,
	}, nil
}

func (k *KlineDatas) WilliamsR(period int) (*TaWilliamsR, error) {
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
	return CalculateWilliamsR(high, low, close, period)
}

func (k *KlineDatas) WilliamsR_(period int) float64 {

	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	wr, err := _k.WilliamsR(period)
	if err != nil {
		return 0
	}
	return wr.Value()
}

func IsWilliamsROverbought(wr float64, threshold ...float64) bool {
	th := -20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return wr > th
}

func IsWilliamsROversold(wr float64, threshold ...float64) bool {
	th := -80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return wr < th
}

func IsWilliamsRCrossOver(current, previous, level float64) bool {
	return previous < level && current >= level
}

func IsWilliamsRCrossUnder(current, previous, level float64) bool {
	return previous > level && current <= level
}

func (t *TaWilliamsR) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaWilliamsR) IsOverbought(threshold ...float64) bool {
	th := -20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() > th
}

func (t *TaWilliamsR) IsOversold(threshold ...float64) bool {
	th := -80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.Value() < th
}

func (t *TaWilliamsR) IsBuySignal(threshold ...float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	th := -80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] < th && t.Values[lastIndex] > th
}

func (t *TaWilliamsR) IsSellSignal(threshold ...float64) bool {
	if len(t.Values) < 2 {
		return false
	}
	th := -20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex-1] > th && t.Values[lastIndex] < th
}

func (t *TaWilliamsR) IsBullishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	wrLow := t.Values[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] < wrLow {
			wrLow = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] > wrLow && prices[lastIndex] < priceLow
}

func (t *TaWilliamsR) IsBearishDivergence(prices []float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	wrHigh := t.Values[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Values[lastIndex-i] > wrHigh {
			wrHigh = t.Values[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.Values[lastIndex] < wrHigh && prices[lastIndex] > priceHigh
}

func (t *TaWilliamsR) IsCenterCross() (up, down bool) {
	if len(t.Values) < 2 {
		return false, false
	}
	lastIndex := len(t.Values) - 1
	up = t.Values[lastIndex-1] < -50 && t.Values[lastIndex] > -50
	down = t.Values[lastIndex-1] > -50 && t.Values[lastIndex] < -50
	return
}

func (t *TaWilliamsR) GetTrend() int {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	if t.Values[lastIndex] > -20 {
		return 1
	} else if t.Values[lastIndex] > -50 {
		return 2
	} else if t.Values[lastIndex] < -80 {
		return -1
	} else if t.Values[lastIndex] < -50 {
		return -2
	}
	return 0
}

func (t *TaWilliamsR) GetStrength() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex] + 50)
}

func (t *TaWilliamsR) IsStrengthening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]+50) > math.Abs(t.Values[lastIndex-1]+50)
}

func (t *TaWilliamsR) IsWeakening() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return math.Abs(t.Values[lastIndex]+50) < math.Abs(t.Values[lastIndex-1]+50)
}

func (t *TaWilliamsR) GetMomentum() float64 {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] - t.Values[lastIndex-1]
}

func (t *TaWilliamsR) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.Values) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Values) - 1
	wrChange := (t.Values[lastIndex] - t.Values[lastIndex-1]) / math.Abs(t.Values[lastIndex-1]) * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(wrChange-priceChange) > threshold
}

func (t *TaWilliamsR) GetZonePosition() int {
	value := t.Value()
	if value > -20 {
		return 1
	} else if value > -50 {
		return 2
	} else if value < -80 {
		return -1
	} else if value < -50 {
		return -2
	}
	return 0
}
