package ta

import (
	"math"
)

type TaMacd struct {
	Macd         []float64 `json:"macd"`
	Dif          []float64 `json:"dif"`
	Dea          []float64 `json:"dea"`
	ShortPeriod  int       `json:"short_period"`
	LongPeriod   int       `json:"long_period"`
	SignalPeriod int       `json:"signal_period"`
}

func CalculateMACD(prices []float64, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {

	shortEMA, err := CalculateEMA(prices, shortPeriod)
	if err != nil {
		return nil, err
	}
	longEMA, err := CalculateEMA(prices, longPeriod)
	if err != nil {
		return nil, err
	}

	dif := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		if i < longPeriod-1 {
			dif[i] = 0
		} else {
			dif[i] = shortEMA.Values[i] - longEMA.Values[i]
		}
	}

	dea, err := CalculateEMA(dif, signalPeriod)
	if err != nil {
		return nil, err
	}

	macd := make([]float64, len(prices))
	for i := 0; i < len(prices); i++ {
		macd[i] = 2 * (dif[i] - dea.Values[i]) / 2
	}
	return &TaMacd{
		Macd:         macd,
		Dif:          dif,
		Dea:          dea.Values,
		ShortPeriod:  shortPeriod,
		LongPeriod:   longPeriod,
		SignalPeriod: signalPeriod,
	}, nil
}

func (k *KlineDatas) MACD(source string, shortPeriod, longPeriod, signalPeriod int) (*TaMacd, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
}

func (k *KlineDatas) MACD_(source string, shortPeriod, longPeriod, signalPeriod int) (macd, dif, dea float64) {
	_k, err := k.Keep((longPeriod + signalPeriod) * 2)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0, 0
	}
	m, err := CalculateMACD(prices, shortPeriod, longPeriod, signalPeriod)
	if err != nil {
		return 0, 0, 0
	}
	return m.Value()
}

func (t *TaMacd) Value() (macd, dif, dea float64) {
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex], t.Dif[lastIndex], t.Dea[lastIndex]
}

func (t *TaMacd) IsGoldenCross() bool {
	if len(t.Dif) < 2 || len(t.Dea) < 2 {
		return false
	}
	lastIndex := len(t.Dif) - 1
	return t.Dif[lastIndex-1] <= t.Dea[lastIndex-1] && t.Dif[lastIndex] > t.Dea[lastIndex]
}

func (t *TaMacd) IsDeathCross() bool {
	if len(t.Dif) < 2 || len(t.Dea) < 2 {
		return false
	}
	lastIndex := len(t.Dif) - 1
	return t.Dif[lastIndex-1] >= t.Dea[lastIndex-1] && t.Dif[lastIndex] < t.Dea[lastIndex]
}

func (t *TaMacd) IsBullishDivergence(prices []float64) bool {
	if len(t.Macd) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	macdLow := t.Macd[lastIndex]
	priceLow := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Macd[lastIndex-i] < macdLow {
			macdLow = t.Macd[lastIndex-i]
		}
		if prices[lastIndex-i] < priceLow {
			priceLow = prices[lastIndex-i]
		}
	}
	return t.Macd[lastIndex] > macdLow && prices[lastIndex] < priceLow
}

func (t *TaMacd) IsBearishDivergence(prices []float64) bool {
	if len(t.Macd) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	macdHigh := t.Macd[lastIndex]
	priceHigh := prices[lastIndex]
	for i := 1; i < 20; i++ {
		if t.Macd[lastIndex-i] > macdHigh {
			macdHigh = t.Macd[lastIndex-i]
		}
		if prices[lastIndex-i] > priceHigh {
			priceHigh = prices[lastIndex-i]
		}
	}
	return t.Macd[lastIndex] < macdHigh && prices[lastIndex] > priceHigh
}

func (t *TaMacd) IsZeroCross() (up, down bool) {
	if len(t.Macd) < 2 {
		return false, false
	}
	lastIndex := len(t.Macd) - 1
	up = t.Macd[lastIndex-1] <= 0 && t.Macd[lastIndex] > 0
	down = t.Macd[lastIndex-1] >= 0 && t.Macd[lastIndex] < 0
	return
}

func (t *TaMacd) GetTrend() int {
	if len(t.Macd) < t.SignalPeriod {
		return 0
	}
	lastIndex := len(t.Macd) - 1
	if t.Macd[lastIndex] > 0 && t.Dif[lastIndex] > t.Dea[lastIndex] {
		return 1
	} else if t.Macd[lastIndex] > 0 && t.Dif[lastIndex] < t.Dea[lastIndex] {
		return 2
	} else if t.Macd[lastIndex] < 0 && t.Dif[lastIndex] < t.Dea[lastIndex] {
		return -1
	} else if t.Macd[lastIndex] < 0 && t.Dif[lastIndex] > t.Dea[lastIndex] {
		return -2
	}
	return 0
}

func (t *TaMacd) GetHistogramWidth() float64 {
	return t.Macd[len(t.Macd)-1]
}

func (t *TaMacd) IsHistogramIncreasing() bool {
	if len(t.Macd) < 2 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex] > t.Macd[lastIndex-1]
}

func (t *TaMacd) IsHistogramDecreasing() bool {
	if len(t.Macd) < 2 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	return t.Macd[lastIndex] < t.Macd[lastIndex-1]
}

func (t *TaMacd) GetConvergence() float64 {
	if len(t.Dif) < 1 || len(t.Dea) < 1 {
		return 0
	}
	lastIndex := len(t.Dif) - 1
	return math.Abs(t.Dif[lastIndex] - t.Dea[lastIndex])
}

func (t *TaMacd) IsDivergenceConfirmed(prices []float64, threshold float64) bool {
	if len(t.Macd) < 20 || len(prices) < 20 {
		return false
	}
	lastIndex := len(t.Macd) - 1
	macdChange := (t.Macd[lastIndex] - t.Macd[lastIndex-1]) / math.Abs(t.Macd[lastIndex-1]) * 100
	priceChange := (prices[lastIndex] - prices[lastIndex-1]) / prices[lastIndex-1] * 100
	return math.Abs(macdChange-priceChange) > threshold
}
