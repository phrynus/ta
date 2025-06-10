package ta

import (
	"fmt"
	"math"
)

// TaKAMA 考夫曼自适应移动平均线指标结构体(Kaufman's Adaptive Moving Average)
type TaKAMA struct {
	Values     []float64 `json:"values"`      // KAMA值序列
	Period     int       `json:"period"`      // 计算周期
	FastPeriod int       `json:"fast_period"` // 快速EMA周期
	SlowPeriod int       `json:"slow_period"` // 慢速EMA周期
	Efficiency []float64 `json:"efficiency"`  // 效率因子
}

// CalculateKAMA 计算考夫曼自适应移动平均线
// 参数：
//   - prices: 价格序列
//   - period: 计算周期，通常为10
//   - fastPeriod: 快速EMA周期，通常为2
//   - slowPeriod: 慢速EMA周期，通常为30
//
// 返回值：
//   - *TaKAMA: KAMA指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	KAMA是一种自适应移动平均线，能够根据市场波动自动调整平滑系数
//	计算步骤：
//	1. 计算方向变化(Direction)：当前价格与n周期前价格的差
//	2. 计算波动性(Volatility)：n周期内价格变化的绝对值之和
//	3. 计算效率因子(ER)：Direction/Volatility
//	4. 计算平滑系数(SC)：[ER * (快速常数 - 慢速常数) + 慢速常数]²
//	5. 计算KAMA：前一期KAMA + SC * (当前价格 - 前一期KAMA)
//
// 示例：
//
//	kama, err := CalculateKAMA(prices, 10, 2, 30)
func CalculateKAMA(prices []float64, period, fastPeriod, slowPeriod int) (*TaKAMA, error) {
	length := len(prices)
	if length < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 预分配切片
	slices := preallocateSlices(length, 2) // [values, efficiency]
	values, efficiency := slices[0], slices[1]

	// 计算快速和慢速常数
	fastConst := 2.0 / float64(fastPeriod+1)
	slowConst := 2.0 / float64(slowPeriod+1)

	// 初始化第一个KAMA值
	values[period-1] = prices[period-1]

	for i := period; i < length; i++ {
		// 计算方向变化
		change := math.Abs(prices[i] - prices[i-period])

		// 计算波动性
		volatility := 0.0
		for j := 0; j < period; j++ {
			volatility += math.Abs(prices[i-j] - prices[i-j-1])
		}

		// 计算效率因子
		var er float64
		if volatility > 0 {
			er = change / volatility
		}
		efficiency[i] = er

		// 计算平滑系数
		sc := math.Pow(er*(fastConst-slowConst)+slowConst, 2)

		// 计算KAMA
		values[i] = values[i-1] + sc*(prices[i]-values[i-1])
	}

	return &TaKAMA{
		Values:     values,
		Period:     period,
		FastPeriod: fastPeriod,
		SlowPeriod: slowPeriod,
		Efficiency: efficiency,
	}, nil
}

// KAMA 计算K线数据的KAMA指标
// 参数：
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//   - period: 计算周期
//   - fastPeriod: 快速EMA周期
//   - slowPeriod: 慢速EMA周期
//
// 返回值：
//   - *TaKAMA: KAMA指标结构体
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	kama, err := kline.KAMA("close", 10, 2, 30)
func (k *KlineDatas) KAMA(source string, period, fastPeriod, slowPeriod int) (*TaKAMA, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateKAMA(prices, period, fastPeriod, slowPeriod)
}

// KAMA_ 获取最新的KAMA值
// 参数：
//   - source: 数据来源
//   - period: 计算周期
//   - fastPeriod: 快速EMA周期
//   - slowPeriod: 慢速EMA周期
//
// 返回值：
//   - float64: 最新的KAMA值
//   - float64: 最新的效率因子值
//
// 示例：
//
//	value, efficiency := kline.KAMA_("close", 10, 2, 30)
func (k *KlineDatas) KAMA_(source string, period, fastPeriod, slowPeriod int) (value, efficiency float64) {
	_k, err := k.Keep(period * 2)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0
	}
	kama, err := CalculateKAMA(prices, period, fastPeriod, slowPeriod)
	if err != nil {
		return 0, 0
	}
	return kama.Value()
}

// Value 返回最新的KAMA值和效率因子
// 返回值：
//   - float64: 最新的KAMA值
//   - float64: 最新的效率因子值
//
// 示例：
//
//	value, efficiency := kama.Value()
func (t *TaKAMA) Value() (value, efficiency float64) {
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex], t.Efficiency[lastIndex]
}

// GetTrend 获取KAMA趋势
// 返回值：
//   - int: 趋势值，1=上涨，-1=下跌，0=盘整
//
// 示例：
//
//	trend := kama.GetTrend()
func (t *TaKAMA) GetTrend() int {
	if len(t.Values) < 2 {
		return 0
	}
	lastIndex := len(t.Values) - 1
	diff := t.Values[lastIndex] - t.Values[lastIndex-1]
	if diff > 0 {
		return 1
	} else if diff < 0 {
		return -1
	}
	return 0
}

// IsEfficient 判断当前市场是否有效
// 参数：
//   - threshold: 效率因子阈值，通常为0.5
//
// 返回值：
//   - bool: 如果效率因子大于阈值返回true，否则返回false
//
// 示例：
//
//	isEfficient := kama.IsEfficient(0.5)
func (t *TaKAMA) IsEfficient(threshold float64) bool {
	return t.Efficiency[len(t.Efficiency)-1] > threshold
}

// GetVolatility 获取价格波动率
// 返回值：
//   - float64: 最近period个周期的价格波动率
//
// 示例：
//
//	volatility := kama.GetVolatility()
func (t *TaKAMA) GetVolatility() float64 {
	if len(t.Values) < t.Period {
		return 0
	}
	lastIndex := len(t.Values) - 1
	var sum float64
	for i := 0; i < t.Period; i++ {
		diff := t.Values[lastIndex-i] - t.Values[lastIndex-i-1]
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(t.Period))
}
