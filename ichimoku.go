package ta

import (
	"fmt"
	"sync"
)

// TaIchimoku 一目均衡图指标结构体(Ichimoku Cloud)
// 说明：
//
//	一目均衡图是一个多功能指标，由五条线组成，用于识别支撑位、阻力位、趋势方向、动量和潜在的交易信号
type TaIchimoku struct {
	Tenkan    []float64 `json:"tenkan"`     // 转换线(Conversion Line)
	Kijun     []float64 `json:"kijun"`      // 基准线(Base Line)
	SenkouA   []float64 `json:"senkou_a"`   // 先行带A(Leading Span A)
	SenkouB   []float64 `json:"senkou_b"`   // 先行带B(Leading Span B)
	Chikou    []float64 `json:"chikou"`     // 延迟线(Lagging Span)
	Future    int       `json:"future"`     // 未来周期数
	ShiftBack int       `json:"shift_back"` // 延迟周期数
}

// calculateMidpoint 计算最高价和最低价的中点(Calculate Midpoint)
// 参数：
//   - high: 最高价数组
//   - low: 最低价数组
//   - period: 计算周期
//   - start: 起始位置
//   - end: 结束位置
//   - result: 结果数组
//
// 说明：
//
//	计算指定周期内最高价和最低价的中点值
func calculateMidpoint(high, low []float64, period int, start, end int, result []float64) {
	for i := start; i < end; i++ {
		if i < period-1 {
			continue
		}

		// 计算周期内的最高价和最低价
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

		// 计算中点
		result[i] = (highestHigh + lowestLow) / 2
	}
}

// CalculateIchimoku 计算一目均衡图指标(Calculate Ichimoku Cloud)
// 参数：
//   - high: 最高价数组
//   - low: 最低价数组
//   - tenkanPeriod: 转换线周期(默认9)
//   - kijunPeriod: 基准线周期(默认26)
//   - senkouBPeriod: 先行带B周期(默认52)
//
// 返回值：
//   - *TaIchimoku: 一目均衡图指标结构体指针
//   - error: 错误信息
//
// 说明：
//
//	计算一目均衡图的五条线：转换线、基准线、先行带A、先行带B和延迟线
//
// 示例：
//
//	ichimoku, err := CalculateIchimoku(high, low, 9, 26, 52)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前转换线值：%v\n", ichimoku.Tenkan[len(ichimoku.Tenkan)-1])
func CalculateIchimoku(high, low []float64, tenkanPeriod, kijunPeriod, senkouBPeriod int) (*TaIchimoku, error) {
	if len(high) < senkouBPeriod || len(low) < senkouBPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(high)
	future := kijunPeriod
	shiftBack := kijunPeriod

	// 预分配所有需要的切片
	slices := preallocateSlices(length+future, 5) // [tenkan, kijun, senkouA, senkouB, chikou]
	tenkan, kijun, senkouA, senkouB, chikou := slices[0], slices[1], slices[2], slices[3], slices[4]

	// 并行计算转换线和基准线
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

	// 计算先行带A和B
	wg.Add(2)

	go func() {
		defer wg.Done()
		// 计算先行带A
		for i := 0; i < length; i++ {
			if i < kijunPeriod-1 {
				continue
			}
			senkouA[i+future] = (tenkan[i] + kijun[i]) / 2
		}
	}()

	go func() {
		defer wg.Done()
		// 计算先行带B
		calculateMidpoint(high, low, senkouBPeriod, 0, length, senkouB)
		// 将先行带B向前移动future个周期
		for i := length - 1; i >= 0; i-- {
			if i+future < len(senkouB) {
				senkouB[i+future] = senkouB[i]
			}
		}
	}()

	wg.Wait()

	// 计算延迟线（当前收盘价向后移动shiftBack个周期）
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

// Ichimoku 计算K线数据的一目均衡图指标(Calculate Ichimoku from Kline)
// 参数：
//   - tenkanPeriod: 转换线周期(默认9)
//   - kijunPeriod: 基准线周期(默认26)
//   - senkouBPeriod: 先行带B周期(默认52)
//
// 返回值：
//   - *TaIchimoku: 一目均衡图指标结构体指针
//   - error: 错误信息
//
// 示例：
//
//	ichimoku, err := kline.Ichimoku(9, 26, 52)
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) Ichimoku(tenkanPeriod, kijunPeriod, senkouBPeriod int) (*TaIchimoku, error) {
	high, err := k._ExtractSlice("high")
	if err != nil {
		return nil, err
	}
	low, err := k._ExtractSlice("low")
	if err != nil {
		return nil, err
	}
	return CalculateIchimoku(high, low, tenkanPeriod, kijunPeriod, senkouBPeriod)
}

// Ichimoku_ 获取最新的一目均衡图值(Get Latest Ichimoku Values)
// 参数：
//   - tenkanPeriod: 转换线周期(默认9)
//   - kijunPeriod: 基准线周期(默认26)
//   - senkouBPeriod: 先行带B周期(默认52)
//
// 返回值：
//   - tenkan: 最新转换线值
//   - kijun: 最新基准线值
//   - senkouA: 最新先行带A值
//   - senkouB: 最新先行带B值
//   - chikou: 最新延迟线值
//
// 示例：
//
//	tenkan, kijun, senkouA, senkouB, chikou := kline.Ichimoku_(9, 26, 52)
func (k *KlineDatas) Ichimoku_(tenkanPeriod, kijunPeriod, senkouBPeriod int) (tenkan, kijun, senkouA, senkouB, chikou float64) {
	// 只保留必要的计算数据
	_k, err := k._Keep(senkouBPeriod * 2)
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

// IsTenkanKijunCross 判断转换线和基准线是否交叉(Check Tenkan-Kijun Cross)
// 返回值：
//   - golden: 金叉信号(买入)
//   - death: 死叉信号(卖出)
//
// 说明：
//
//	检测转换线和基准线是否发生交叉，用于生成交易信号
//
// 示例：
//
//	golden, death := ichimoku.IsTenkanKijunCross()
//	if golden {
//	    fmt.Println("出现金叉买入信号")
//	}
func (t *TaIchimoku) IsTenkanKijunCross() (golden, death bool) {
	if len(t.Tenkan) < 2 || len(t.Kijun) < 2 {
		return false, false
	}
	lastIndex := len(t.Tenkan) - 1
	golden = t.Tenkan[lastIndex-1] <= t.Kijun[lastIndex-1] && t.Tenkan[lastIndex] > t.Kijun[lastIndex]
	death = t.Tenkan[lastIndex-1] >= t.Kijun[lastIndex-1] && t.Tenkan[lastIndex] < t.Kijun[lastIndex]
	return
}

// IsInKumo 判断价格是否在云层中(Check if Price is Inside Kumo)
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - bool: 是否在云层中
//
// 说明：
//
//	判断当前价格是否位于云层(先行带A和B之间)
//
// 示例：
//
//	if ichimoku.IsInKumo(currentPrice) {
//	    fmt.Println("价格在云层中，趋势不明确")
//	}
func (t *TaIchimoku) IsInKumo(price float64) bool {
	lastIndex := len(t.SenkouA) - 1
	high := max(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	low := min(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	return price >= low && price <= high
}

// IsAboveKumo 判断价格是否在云层之上(Check if Price is Above Kumo)
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - bool: 是否在云层之上
//
// 说明：
//
//	判断当前价格是否位于云层上方，表示可能处于上升趋势
//
// 示例：
//
//	if ichimoku.IsAboveKumo(currentPrice) {
//	    fmt.Println("价格在云层上方，可能是上升趋势")
//	}
func (t *TaIchimoku) IsAboveKumo(price float64) bool {
	lastIndex := len(t.SenkouA) - 1
	high := max(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	return price > high
}

// IsBelowKumo 判断价格是否在云层之下(Check if Price is Below Kumo)
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - bool: 是否在云层之下
//
// 说明：
//
//	判断当前价格是否位于云层下方，表示可能处于下降趋势
//
// 示例：
//
//	if ichimoku.IsBelowKumo(currentPrice) {
//	    fmt.Println("价格在云层下方，可能是下降趋势")
//	}
func (t *TaIchimoku) IsBelowKumo(price float64) bool {
	lastIndex := len(t.SenkouA) - 1
	low := min(t.SenkouA[lastIndex], t.SenkouB[lastIndex])
	return price < low
}

// IsKumoTwist 判断云层是否发生扭转(Check Kumo Twist)
// 返回值：
//   - bullish: 看涨扭转
//   - bearish: 看跌扭转
//
// 说明：
//
//	检测云层(先行带A和B)是否发生扭转，这通常是重要的趋势转换信号
//
// 示例：
//
//	bullish, bearish := ichimoku.IsKumoTwist()
//	if bullish {
//	    fmt.Println("云层发生看涨扭转")
//	}
func (t *TaIchimoku) IsKumoTwist() (bullish, bearish bool) {
	if len(t.SenkouA) < 2 || len(t.SenkouB) < 2 {
		return false, false
	}
	lastIndex := len(t.SenkouA) - 1
	bullish = t.SenkouA[lastIndex-1] <= t.SenkouB[lastIndex-1] && t.SenkouA[lastIndex] > t.SenkouB[lastIndex]
	bearish = t.SenkouA[lastIndex-1] >= t.SenkouB[lastIndex-1] && t.SenkouA[lastIndex] < t.SenkouB[lastIndex]
	return
}

// IsChikouCrossPrice 判断延迟线是否穿越价格(Check Chikou Span Cross Price)
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - bullish: 看涨穿越
//   - bearish: 看跌穿越
//
// 说明：
//
//	检测延迟线是否穿越价格，这是一个重要的确认信号
//
// 示例：
//
//	bullish, bearish := ichimoku.IsChikouCrossPrice(currentPrice)
//	if bullish {
//	    fmt.Println("延迟线向上穿越价格，确认买入信号")
//	}
func (t *TaIchimoku) IsChikouCrossPrice(price float64) (bullish, bearish bool) {
	if len(t.Chikou) < 2 {
		return false, false
	}
	lastIndex := len(t.Chikou) - 1
	bullish = t.Chikou[lastIndex-1] <= price && t.Chikou[lastIndex] > price
	bearish = t.Chikou[lastIndex-1] >= price && t.Chikou[lastIndex] < price
	return
}

// IsStrongTrend 判断是否处于强势趋势
func (t *TaIchimoku) IsStrongTrend(price float64) (bullish, bearish bool) {
	lastIndex := len(t.Tenkan) - 1
	bullish = price > t.Tenkan[lastIndex] && t.Tenkan[lastIndex] > t.Kijun[lastIndex] &&
		t.Kijun[lastIndex] > t.SenkouA[lastIndex] && t.SenkouA[lastIndex] > t.SenkouB[lastIndex]
	bearish = price < t.Tenkan[lastIndex] && t.Tenkan[lastIndex] < t.Kijun[lastIndex] &&
		t.Kijun[lastIndex] < t.SenkouA[lastIndex] && t.SenkouA[lastIndex] < t.SenkouB[lastIndex]
	return
}
