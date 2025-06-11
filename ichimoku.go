package ta

import (
	"fmt"
	"sync"
)

type TaIchimoku struct {
	Tenkan    []float64 `json:"tenkan"`
	Kijun     []float64 `json:"kijun"`
	SenkouA   []float64 `json:"senkou_a"`
	SenkouB   []float64 `json:"senkou_b"`
	Chikou    []float64 `json:"chikou"`
	Future    int       `json:"future"`
	ShiftBack int       `json:"shift_back"`
}

func calculateMidpoint(high, low []float64, period int, start, end int, result []float64) {
	for i := start; i < end; i++ {
		if i < period-1 {
			continue
		}

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

		result[i] = (highestHigh + lowestLow) / 2
	}
}

func CalculateIchimoku(high, low []float64, tenkanPeriod, kijunPeriod, senkouBPeriod int) (*TaIchimoku, error) {
	if len(high) < senkouBPeriod || len(low) < senkouBPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(high)
	future := kijunPeriod
	shiftBack := kijunPeriod

	slices := preallocateSlices(length+future, 5)
	tenkan, kijun, senkouA, senkouB, chikou := slices[0], slices[1], slices[2], slices[3], slices[4]

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		calculateMidpoint(high, low, tenkanPeriod, 0, length, tenkan)
	}()

	go func() {
		defer wg.Done()
		calculateMidpoint(high, low, kijunPeriod, 0, length, kijun)
	}()

	wg.Wait()

	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := 0; i < length; i++ {
			if i < kijunPeriod-1 {
				continue
			}
			senkouA[i+future] = (tenkan[i] + kijun[i]) / 2
		}
	}()

	go func() {
		defer wg.Done()

		calculateMidpoint(high, low, senkouBPeriod, 0, length, senkouB)

		for i := length - 1; i >= 0; i-- {
			if i+future < len(senkouB) {
				senkouB[i+future] = senkouB[i]
			}
		}
	}()

	wg.Wait()

	for i := shiftBack; i < length; i++ {
		chikou[i-shiftBack] = high[i]
	}

	return &TaIchimoku{
		Tenkan:    tenkan,
		Kijun:     kijun,
		SenkouA:   senkouA,
		SenkouB:   senkouB,
		Chikou:    chikou,
		Future:    future,
		ShiftBack: shiftBack,
	}, nil
}

func (k *KlineDatas) Ichimoku(tenkanPeriod, kijunPeriod, senkouBPeriod int) (*TaIchimoku, error) {
	high, err := k.ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k.ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	return CalculateIchimoku(high, low, tenkanPeriod, kijunPeriod, senkouBPeriod)
}

func (k *KlineDatas) Ichimoku_(tenkanPeriod, kijunPeriod, senkouBPeriod int) (tenkan, kijun, senkouA, senkouB, chikou float64) {

	_k, err := k.Keep(senkouBPeriod * 2)
	if err != nil {
		_k = *k
	}
	ichimoku, err := _k.Ichimoku(tenkanPeriod, kijunPeriod, senkouBPeriod)
	if err != nil {
		return 0, 0, 0, 0, 0
	}
	lastIndex := len(ichimoku.Tenkan) - 1
	return ichimoku.Tenkan[lastIndex],
		ichimoku.Kijun[lastIndex],
		ichimoku.SenkouA[lastIndex],
		ichimoku.SenkouB[lastIndex],
		ichimoku.Chikou[lastIndex]
}

func (t *TaIchimoku) IsTenkanKijunCross() (golden, death bool) {
	if len(t.Tenkan) < 2 || len(t.Kijun) < 2 {
		return false, false
	}
	lastIndex := len(t.Tenkan) - 1
	golden = t.Tenkan[lastIndex-1] <= t.Kijun[lastIndex-1] && t.Tenkan[lastIndex] > t.Kijun[lastIndex]
	death = t.Tenkan[lastIndex-1] >= t.Kijun[lastIndex-1] && t.Tenkan[lastIndex] < t.Kijun[lastIndex]
	return
}

func (t *TaIchimoku) IsInKumo(price float64) bool {
	lastIndex := len(t.SenkouA) - 1
	high := max(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	low := min(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	return price >= low && price <= high
}

func (t *TaIchimoku) IsAboveKumo(price float64) bool {
	lastIndex := len(t.SenkouA) - 1
	high := max(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	return price > high
}

func (t *TaIchimoku) IsBelowKumo(price float64) bool {
	lastIndex := len(t.SenkouA) - 1
	low := min(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	return price < low
}

func (t *TaIchimoku) IsKumoTwist() (bullish, bearish bool) {
	if len(t.SenkouA) < 2 || len(t.SenkouB) < 2 {
		return false, false
	}
	lastIndex := len(t.SenkouA) - 1
	bullish = t.SenkouA[lastIndex-1] <= t.SenkouB[lastIndex-1] && t.SenkouA[lastIndex] > t.SenkouB[lastIndex]
	bearish = t.SenkouA[lastIndex-1] >= t.SenkouB[lastIndex-1] && t.SenkouA[lastIndex] < t.SenkouB[lastIndex]
	return
}

func (t *TaIchimoku) IsChikouCrossPrice(price float64) (bullish, bearish bool) {
	if len(t.Chikou) < 2 {
		return false, false
	}
	lastIndex := len(t.Chikou) - 1
	bullish = t.Chikou[lastIndex-1] <= price && t.Chikou[lastIndex] > price
	bearish = t.Chikou[lastIndex-1] >= price && t.Chikou[lastIndex] < price
	return
}

func (t *TaIchimoku) IsStrongTrend(price float64) (bullish, bearish bool) {
	lastIndex := len(t.Tenkan) - 1
	bullish = price > t.Tenkan[lastIndex] && t.Tenkan[lastIndex] > t.Kijun[lastIndex] &&
		t.Kijun[lastIndex] > t.SenkouA[lastIndex] && t.SenkouA[lastIndex] > t.SenkouB[lastIndex]
	bearish = price < t.Tenkan[lastIndex] && t.Tenkan[lastIndex] < t.Kijun[lastIndex] &&
		t.Kijun[lastIndex] < t.SenkouA[lastIndex] && t.SenkouA[lastIndex] < t.SenkouB[lastIndex]
	return
}
