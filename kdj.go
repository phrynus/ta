package ta

import (
	"fmt"
)

type TaKDJ struct {
	K []float64 `json:"k"`
	D []float64 `json:"d"`
	J []float64 `json:"j"`
}

func CalculateKDJ(high, low, close []float64, rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
	if len(high) < rsvPeriod || len(low) < rsvPeriod || len(close) < rsvPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(close)

	slices := preallocateSlices(length, 4)
	rsv, k, d, j := slices[0], slices[1], slices[2], slices[3]

	for i := rsvPeriod - 1; i < length; i++ {

		var highestHigh, lowestLow = high[i], low[i]
		for j := 0; j < rsvPeriod; j++ {
			idx := i - j
			if high[idx] > highestHigh {
				highestHigh = high[idx]
			}
			if low[idx] < lowestLow {
				lowestLow = low[idx]
			}
		}

		if highestHigh != lowestLow {
			rsv[i] = (close[i] - lowestLow) / (highestHigh - lowestLow) * 100
		} else {
			rsv[i] = 50
		}
	}

	k[rsvPeriod-1] = rsv[rsvPeriod-1]
	d[rsvPeriod-1] = rsv[rsvPeriod-1]
	j[rsvPeriod-1] = rsv[rsvPeriod-1]

	for i := rsvPeriod; i < length; i++ {

		k[i] = (2.0*k[i-1] + rsv[i]) / 3.0

		d[i] = (2.0*d[i-1] + k[i]) / 3.0

		j[i] = 3.0*k[i] - 2.0*d[i]
	}

	return &TaKDJ{
		K: k,
		D: d,
		J: j,
	}, nil
}

func (k *KlineDatas) KDJ(rsvPeriod, kPeriod, dPeriod int) (*TaKDJ, error) {
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
	return CalculateKDJ(high, low, close, rsvPeriod, kPeriod, dPeriod)
}

func (k *KlineDatas) KDJ_(rsvPeriod, kPeriod, dPeriod int) (kValue, dValue, jValue float64) {

	_k, err := k.Keep(rsvPeriod * 2)
	if err != nil {
		_k = *k
	}
	kdj, err := _k.KDJ(rsvPeriod, kPeriod, dPeriod)
	if err != nil {
		return 0, 0, 0
	}
	lastIndex := len(kdj.K) - 1
	return kdj.K[lastIndex], kdj.D[lastIndex], kdj.J[lastIndex]
}

func (t *TaKDJ) Value() (k, d, j float64) {
	lastIndex := len(t.K) - 1
	return t.K[lastIndex], t.D[lastIndex], t.J[lastIndex]
}

func (t *TaKDJ) IsGoldenCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] <= t.D[lastIndex-1] && t.K[lastIndex] > t.D[lastIndex]
}

func (t *TaKDJ) IsDeathCross() bool {
	if len(t.K) < 2 || len(t.D) < 2 {
		return false
	}
	lastIndex := len(t.K) - 1
	return t.K[lastIndex-1] >= t.D[lastIndex-1] && t.K[lastIndex] < t.D[lastIndex]
}

func (t *TaKDJ) IsOverbought(threshold ...float64) bool {
	th := 80.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] > th && t.D[len(t.D)-1] > th
}

func (t *TaKDJ) IsOversold(threshold ...float64) bool {
	th := 20.0
	if len(threshold) > 0 {
		th = threshold[0]
	}
	return t.K[len(t.K)-1] < th && t.D[len(t.D)-1] < th
}

func (t *TaKDJ) IsBullishDivergence() bool {
	if len(t.K) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	previousLow := t.K[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.K[lastIndex-i] < previousLow {
			previousLow = t.K[lastIndex-i]
		}
	}
	return t.K[lastIndex] > previousLow && t.K[lastIndex] < 20
}

func (t *TaKDJ) IsBearishDivergence() bool {
	if len(t.K) < 20 {
		return false
	}
	lastIndex := len(t.K) - 1
	previousHigh := t.K[lastIndex-1]
	for i := 2; i < 20; i++ {
		if t.K[lastIndex-i] > previousHigh {
			previousHigh = t.K[lastIndex-i]
		}
	}
	return t.K[lastIndex] < previousHigh && t.K[lastIndex] > 80
}

func (t *TaKDJ) IsExtremeBought() bool {
	return t.K[len(t.K)-1] > 90 && t.D[len(t.D)-1] > 90
}

func (t *TaKDJ) IsExtremeSold() bool {
	return t.K[len(t.K)-1] < 10 && t.D[len(t.D)-1] < 10
}

func (t *TaKDJ) IsTrendStrengthening() bool {
	if len(t.K) < 4 {
		return false
	}
	lastIndex := len(t.K) - 1
	diff1 := t.K[lastIndex] - t.K[lastIndex-1]
	diff2 := t.K[lastIndex-1] - t.K[lastIndex-2]
	diff3 := t.K[lastIndex-2] - t.K[lastIndex-3]
	return (diff1 > diff2 && diff2 > diff3) || (diff1 < diff2 && diff2 < diff3)
}

func (t *TaKDJ) IsJCrossK() (bullish, bearish bool) {
	if len(t.K) < 2 || len(t.J) < 2 {
		return false, false
	}
	lastIndex := len(t.K) - 1
	bullish = t.J[lastIndex-1] <= t.K[lastIndex-1] && t.J[lastIndex] > t.K[lastIndex]
	bearish = t.J[lastIndex-1] >= t.K[lastIndex-1] && t.J[lastIndex] < t.K[lastIndex]
	return
}
