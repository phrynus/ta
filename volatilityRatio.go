package ta

import (
	"fmt"
	"math"
)

type TaVolatilityRatio struct {
	Values []float64 `json:"values"`
	Period int       `json:"period"`
}

func CalculateVolatilityRatio(klineData KlineDatas, shortPeriod, longPeriod int) (*TaVolatilityRatio, error) {
	if len(klineData) < longPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)
	slices := preallocateSlices(length, 2)
	trueRange, ratio := slices[0], slices[1]

	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevClose := klineData[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	for i := longPeriod; i < length; i++ {
		var shortTR float64
		for j := i - shortPeriod + 1; j <= i; j++ {
			shortTR += trueRange[j]
		}
		shortTR /= float64(shortPeriod)

		var longTR float64
		for j := i - longPeriod + 1; j <= i; j++ {
			longTR += trueRange[j]
		}
		longTR /= float64(longPeriod)

		if longTR != 0 {
			ratio[i] = shortTR / longTR
		} else {
			ratio[i] = 1.0
		}
	}

	return &TaVolatilityRatio{
		Values: ratio,
		Period: longPeriod,
	}, nil
}

func (k KlineDatas) VolatilityRatio(shortPeriod, longPeriod int) (*TaVolatilityRatio, error) {
	return CalculateVolatilityRatio(k, shortPeriod, longPeriod)
}

func (k KlineDatas) VolatilityRatio_(shortPeriod, longPeriod int) float64 {
	vr, err := k.VolatilityRatio(shortPeriod, longPeriod)
	if err != nil || len(vr.Values) == 0 {
		return 0
	}
	return vr.Values[len(vr.Values)-1]
}

func (vr *TaVolatilityRatio) Value() float64 {
	if len(vr.Values) == 0 {
		return 0
	}
	return vr.Values[len(vr.Values)-1]
}

func (vr *TaVolatilityRatio) IsHighVolatility(threshold float64) bool {
	if len(vr.Values) == 0 {
		return false
	}
	return vr.Value() > threshold
}

func (vr *TaVolatilityRatio) IsLowVolatility(threshold float64) bool {
	if len(vr.Values) == 0 {
		return false
	}
	return vr.Value() < threshold
}

func (vr *TaVolatilityRatio) IsIncreasing() bool {
	if len(vr.Values) < 2 {
		return false
	}
	return vr.Values[len(vr.Values)-1] > vr.Values[len(vr.Values)-2]
}

func (vr *TaVolatilityRatio) IsDecreasing() bool {
	if len(vr.Values) < 2 {
		return false
	}
	return vr.Values[len(vr.Values)-1] < vr.Values[len(vr.Values)-2]
}

func (vr *TaVolatilityRatio) GetTrend() int {
	if len(vr.Values) < 2 {
		return 0
	}
	if vr.IsIncreasing() {
		return 1
	}
	if vr.IsDecreasing() {
		return -1
	}
	return 0
}

func (vr *TaVolatilityRatio) GetStrength() float64 {
	if len(vr.Values) == 0 {
		return 0
	}
	return math.Abs(vr.Value() - 1.0)
}

func (vr *TaVolatilityRatio) IsDiverging(prices []float64) bool {
	if len(vr.Values) < 2 || len(prices) < 2 {
		return false
	}

	pricesTrend := prices[len(prices)-1] > prices[len(prices)-2]
	vrTrend := vr.Values[len(vr.Values)-1] > vr.Values[len(vr.Values)-2]

	return pricesTrend != vrTrend
}

func (vr *TaVolatilityRatio) GetOptimalPeriod(prices []float64) int {

	if len(prices) < 20 {
		return vr.Period
	}

	mean := 0.0
	for _, price := range prices {
		mean += price
	}
	mean /= float64(len(prices))

	variance := 0.0
	for _, price := range prices {
		variance += math.Pow(price-mean, 2)
	}
	variance /= float64(len(prices))
	stdDev := math.Sqrt(variance)

	if stdDev > mean*0.1 {
		return int(math.Max(5, float64(vr.Period)*0.7))
	} else if stdDev < mean*0.01 {
		return int(math.Min(30, float64(vr.Period)*1.3))
	}
	return vr.Period
}

func (vr *TaVolatilityRatio) IsBreakoutLikely() bool {
	if len(vr.Values) < vr.Period {
		return false
	}

	recentLow := true
	for i := len(vr.Values) - vr.Period; i < len(vr.Values)-1; i++ {
		if vr.Values[i] > 0.8 {
			recentLow = false
			break
		}
	}

	return recentLow && vr.IsIncreasing()
}

func (vr *TaVolatilityRatio) GetVolatilityZone() string {
	if len(vr.Values) == 0 {
		return "normal"
	}

	currentValue := vr.Value()
	if currentValue > 1.5 {
		return "high"
	} else if currentValue < 0.5 {
		return "low"
	}
	return "normal"
}

func (vr *TaVolatilityRatio) IsConsolidating() bool {
	if len(vr.Values) < vr.Period {
		return false
	}

	var sum, sumSquares float64
	for i := len(vr.Values) - vr.Period; i < len(vr.Values); i++ {
		sum += vr.Values[i]
		sumSquares += vr.Values[i] * vr.Values[i]
	}

	mean := sum / float64(vr.Period)
	variance := sumSquares/float64(vr.Period) - mean*mean
	stdDev := math.Sqrt(variance)

	return stdDev < 0.1
}

func (vr *TaVolatilityRatio) GetVolatilityScore() float64 {
	if len(vr.Values) == 0 {
		return 50
	}

	value := vr.Value()
	if value <= 0.5 {
		return value * 100 / 0.5
	} else if value >= 1.5 {
		return math.Min(100, 50+50*(value-1)/0.5)
	}
	return 50 + 50*(value-1)
}
