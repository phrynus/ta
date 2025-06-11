package ta

import (
	"fmt"
)

type TaSuperTrendPivotHl2 struct {
	Values     []float64 `json:"values"`
	Direction  []int     `json:"direction"`
	UpperBand  []float64 `json:"upper_band"`
	LowerBand  []float64 `json:"lower_band"`
	Period     int       `json:"period"`
	Multiplier float64   `json:"multiplier"`
}

func CalculateSuperTrendPivotHl2(klineData KlineDatas, period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	length := len(klineData)
	if length < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	atr, err := CalculateATR(klineData, period)
	if err != nil {
		return nil, err
	}

	slices := preallocateSlices(length, 4)
	values, direction, upperBand, lowerBand := slices[0], make([]int, length), slices[2], slices[3]

	for i := 0; i < length; i++ {

		hl2 := (klineData[i].High + klineData[i].Low) / 2

		if i < period {

			upperBand[i] = hl2 + multiplier*atr.Values[i]
			lowerBand[i] = hl2 - multiplier*atr.Values[i]
			direction[i] = 0
			values[i] = hl2
			continue
		}

		basicUpperBand := hl2 + multiplier*atr.Values[i]
		basicLowerBand := hl2 - multiplier*atr.Values[i]

		if basicLowerBand > lowerBand[i-1] || klineData[i-1].Close < lowerBand[i-1] {
			lowerBand[i] = basicLowerBand
		} else {
			lowerBand[i] = lowerBand[i-1]
		}

		if basicUpperBand < upperBand[i-1] || klineData[i-1].Close > upperBand[i-1] {
			upperBand[i] = basicUpperBand
		} else {
			upperBand[i] = upperBand[i-1]
		}

		if direction[i-1] <= 0 {
			if klineData[i].Close > upperBand[i] {
				direction[i] = 1
			} else {
				direction[i] = -1
			}
		} else {
			if klineData[i].Close < lowerBand[i] {
				direction[i] = -1
			} else {
				direction[i] = 1
			}
		}

		if direction[i] == 1 {
			values[i] = lowerBand[i]
		} else {
			values[i] = upperBand[i]
		}
	}

	return &TaSuperTrendPivotHl2{
		Values:     values,
		Direction:  direction,
		UpperBand:  upperBand,
		LowerBand:  lowerBand,
		Period:     period,
		Multiplier: multiplier,
	}, nil
}

func (k *KlineDatas) SuperTrendPivotHl2(period int, multiplier float64) (*TaSuperTrendPivotHl2, error) {
	return CalculateSuperTrendPivotHl2(*k, period, multiplier)
}

func (k *KlineDatas) SuperTrendPivotHl2_(period int, multiplier float64) float64 {
	_k, err := k.Keep(period * 2)
	if err != nil {
		_k = *k
	}
	st, err := CalculateSuperTrendPivotHl2(_k, period, multiplier)
	if err != nil {
		return 0
	}
	return st.Value()
}

func (t *TaSuperTrendPivotHl2) Value() float64 {
	return t.Values[len(t.Values)-1]
}

func (t *TaSuperTrendPivotHl2) GetDirection() int {
	return t.Direction[len(t.Direction)-1]
}

func (t *TaSuperTrendPivotHl2) GetBands() (upper, lower float64) {
	lastIndex := len(t.Values) - 1
	return t.UpperBand[lastIndex], t.LowerBand[lastIndex]
}

/* Pine

indicator('SuperTrend Pivot HL2', overlay = true)


atrPeriod = input.int(14, 'ATR周期', minval = 1)
atrMultiplier = input.float(3.0, 'ATR乘数', minval = 0.1, step = 0.1)


atr = ta.atr(atrPeriod)


upperBand = hl2 + atrMultiplier * atr
lowerBand = hl2 - atrMultiplier * atr


var float finalUpperBand = na
var float finalLowerBand = na
var int trend = 0


finalUpperBand := if upperBand < nz(finalUpperBand[1], upperBand) or close[1] > nz(finalUpperBand[1], upperBand)
    upperBand
else
    nz(finalUpperBand[1], upperBand)

finalLowerBand := if lowerBand > nz(finalLowerBand[1], lowerBand) or close[1] < nz(finalLowerBand[1], lowerBand)
    lowerBand
else
    nz(finalLowerBand[1], lowerBand)


trend := if trend[1] <= 0
    close > finalUpperBand ? 1 : -1
else
    close < finalLowerBand ? -1 : 1


superTrend = trend == 1 ? finalLowerBand : finalUpperBand


upTrend = trend == 1
downTrend = trend == -1

plot(superTrend, 'SuperTrend', color = upTrend ? color.green : color.red, linewidth = 2)
plot(finalUpperBand, 'Upper Band', color = color.new(color.gray, 50))
plot(finalLowerBand, 'Lower Band', color = color.new(color.gray, 50))





plotshape(trend != trend[1] and upTrend, 'Up Trend', style = shape.triangleup, location = location.belowbar, color = color.green, size = size.small)
plotshape(trend != trend[1] and downTrend, 'Down Trend', style = shape.triangledown, location = location.abovebar, color = color.red, size = size.small)


alertcondition(trend != trend[1] and upTrend, '买入信号', '价格突破上轨，趋势转多')
alertcondition(trend != trend[1] and downTrend, '卖出信号', '价格突破下轨，趋势转空')

*/
