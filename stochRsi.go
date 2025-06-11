package ta

import (
	"fmt"
	"math"
)

type TaStochRSI struct {
	K           []float64 `json:"k"`
	D           []float64 `json:"d"`
	RsiPeriod   int       `json:"rsi_period"`
	StochPeriod int       `json:"stoch_period"`
	KPeriod     int       `json:"k_period"`
	DPeriod     int       `json:"d_period"`
}

func CalculateStochRSI(prices []float64, rsiPeriod, stochPeriod, kPeriod, dPeriod int) (*TaStochRSI, error) {
	if len(prices) < rsiPeriod+stochPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	rsi, err := CalculateRSI(prices, rsiPeriod)
	if err != nil {
		return nil, err
	}

	length := len(prices)

	slices := preallocateSlices(length, 3)
	stochRsi, k, d := slices[0], slices[1], slices[2]

	for i := stochPeriod - 1; i < length; i++ {

		var highestRsi, lowestRsi = rsi.Values[i], rsi.Values[i]
		for j := 0; j < stochPeriod; j++ {
			idx := i - j
			if rsi.Values[idx] > highestRsi {
				highestRsi = rsi.Values[idx]
			}
			if rsi.Values[idx] < lowestRsi {
				lowestRsi = rsi.Values[idx]
			}
		}

		if highestRsi != lowestRsi {
			stochRsi[i] = (rsi.Values[i] - lowestRsi) / (highestRsi - lowestRsi) * 100
		} else {
			stochRsi[i] = 50
		}
	}

	var sumK float64
	for i := 0; i < kPeriod && i < length; i++ {
		sumK += stochRsi[i]
	}
	k[kPeriod-1] = sumK / float64(kPeriod)

	for i := kPeriod; i < length; i++ {
		sumK = sumK - stochRsi[i-kPeriod] + stochRsi[i]
		k[i] = sumK / float64(kPeriod)
	}

	var sumD float64
	for i := 0; i < dPeriod && i < length; i++ {
		sumD += k[i]
	}
	d[dPeriod-1] = sumD / float64(dPeriod)

	for i := dPeriod; i < length; i++ {
		sumD = sumD - k[i-dPeriod] + k[i]
		d[i] = sumD / float64(dPeriod)
	}

	return &TaStochRSI{
		K:           k,
		D:           d,
		RsiPeriod:   rsiPeriod,
		StochPeriod: stochPeriod,
		KPeriod:     kPeriod,
		DPeriod:     dPeriod,
	}, nil
}

func (k *KlineDatas) StochRSI(rsiPeriod, stochPeriod, kPeriod, dPeriod int, source string) (*TaStochRSI, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateStochRSI(prices, rsiPeriod, stochPeriod, kPeriod, dPeriod)
}

func (k *KlineDatas) StochRSI_(rsiPeriod, stochPeriod, kPeriod, dPeriod int, source string) (kValue, dValue float64) {
	_k, err := k.Keep((rsiPeriod + stochPeriod) * 2)
	if err != nil {
		_k = *k
	}
	stochRsi, err := _k.StochRSI(rsiPeriod, stochPeriod, kPeriod, dPeriod, source)
	if err != nil {
		return 0, 0
	}
	return stochRsi.Value()
}

func (t *TaStochRSI) Value() (kValue, dValue float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex]
}

func (t *TaStochRSI) IsOverbought(threshold ...float64) bool {
	th := 80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] > th && t.D[len(t.D)-1] > th
}

func (t *TaStochRSI) IsOversold(threshold ...float64) bool {
	th := 20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] < th && t.D[len(t.D)-1] < th
}

func (t *TaStochRSI) IsGoldenCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] <= t.D[lastIndex-1] && t.K[lastIndex] > t.D[lastIndex]
}

func (t *TaStochRSI) IsDeathCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] >= t.D[lastIndex-1] && t.K[lastIndex] < t.D[lastIndex]
}

func (t *TaStochRSI) IsBullishDivergence(prices []float64) bool {
	if len(t.K) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	kLow := t.K[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.K[lastIndex-i] < kLow {
			kLow = t.K[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.K[lastIndex] > kLow && prices[lastIndex] < priceLow
}

func (t *TaStochRSI) IsBearishDivergence(prices []float64) bool {
	if len(t.K) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	kHigh := t.K[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.K[lastIndex-i] > kHigh {
			kHigh = t.K[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.K[lastIndex] < kHigh && prices[lastIndex] > priceHigh
}

func (t *TaStochRSI) IsCenterCross() (up, down bool) {
	if len(t.K) < 2 {
		return false, false
	}
	lastIndex := len(t.K) - 1
	up = t.K[lastIndex-1] <= 50 && t.K[lastIndex] > 50
	down = t.K[lastIndex-1] >= 50 && t.K[lastIndex] < 50
	return
}

func (t *TaStochRSI) GetTrend() int {
	lastIndex := len(t.K) - 1
	if t.K[lastIndex] > 80 {
		return 1
	} else if t.K[lastIndex] > 50 {
		return 2
	} else if t.K[lastIndex] < 20 {
		return -1
	} else if t.K[lastIndex] < 50 {
		return -2
	}
	return 0
}

func (t *TaStochRSI) GetStrength() float64 {
	lastIndex := len(t.K) - 1
	return math.Abs(t.K[lastIndex] - 50)
}

func (t *TaStochRSI) IsStrengthening() bool {
	if len(t.K) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return math.Abs(t.K[lastIndex]-50) > math.Abs(t.K[lastIndex-1]-50)
}

func (t *TaStochRSI) IsWeakening() bool {
	if len(t.K) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return math.Abs(t.K[lastIndex]-50) < math.Abs(t.K[lastIndex-1]-50)
}

func (t *TaStochRSI) GetKDSpread() float64 {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex] - t.D[lastIndex]
}

func (t *TaStochRSI) IsKDConverging() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	currentSpread := math.Abs(t.K[lastIndex] - t.D[lastIndex])
	previousSpread := math.Abs(t.K[lastIndex-1] - t.D[lastIndex-1])
	return currentSpread < previousSpread
}

func (t *TaStochRSI) IsKDDiverging() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	currentSpread := math.Abs(t.K[lastIndex] - t.D[lastIndex])
	previousSpread := math.Abs(t.K[lastIndex-1] - t.D[lastIndex-1])
	return currentSpread > previousSpread
}

func (t *TaStochRSI) GetMomentum() float64 {
	if len(t.K) < 2 {
		return 0
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex] - t.K[lastIndex-1]
}

func (t *TaStochRSI) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.K) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	kChange := (t.K[lastIndex] - t.K[lastIndex-1]) / math.Abs(t.K[lastIndex-1]) * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(kChange-priceChange) > threshold
}

func (t *TaStochRSI) GetZonePosition() int {
	value := t.K[len(t.K)-1]
	if value > 80 {
		return 1
	} else if value > 50 {
		return 2
	} else if value < 20 {
		return -1
	} else if value < 50 {
		return -2
	}
	return 0
}
