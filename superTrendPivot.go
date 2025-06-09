package ta

import (
	"fmt"
	"math"
)

// TaSuperTrendPivot 基于轴点的超级趋势指标结构体(SuperTrend Pivot Indicator)
// 说明：
//
//	SuperTrendPivot是SuperTrend的增强版本，结合轴点(Pivot Points)来提高趋势判断的准确性
//	主要用于：
//	1. 识别关键支撑和阻力位
//	2. 动态调整趋势带宽度
//	3. 提供更准确的趋势转换信号
//
// 计算公式：
//
//	轴点 = 局部最高点或最低点
//	趋势带宽度 = ATR * Factor
//	上轨 = 轴点 + 趋势带宽度
//	下轨 = 轴点 - 趋势带宽度
type TaSuperTrendPivot struct {
	Upper       []float64 `json:"upper"`        // 上轨线(Upper Band)，阻力位
	Lower       []float64 `json:"lower"`        // 下轨线(Lower Band)，支撑位
	Trend       []int     `json:"trend"`        // 趋势方向(Trend Direction)：1表示上涨，-1表示下跌，0表示横盘
	PivotPeriod int       `json:"pivot_period"` // 轴点计算周期(Pivot Period)
	Factor      float64   `json:"factor"`       // ATR乘数(ATR Factor)
	AtrPeriod   int       `json:"atr_period"`   // ATR计算周期(ATR Period)
}

// FindPivotHighPoint 查找枢轴高点(Find Pivot High Point)
// 参数：
//   - klineData: K线数据集合
//   - index: 当前K线索引
//   - period: 查找周期
//
// 返回值：
//   - float64: 枢轴高点价格
//
// 说明：
//
//	在指定周期内查找局部最高点，通过比较前后period个周期的高点来确定
//
// 示例：
//
//	high := FindPivotHighPoint(klineData, index, 10)
//	if !math.IsNaN(high) {
//	    fmt.Printf("找到枢轴高点：%.2f\n", high)
//	}
func FindPivotHighPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].High > klineData[index].High {
			return math.NaN()
		}
	}
	return klineData[index].High
}

// FindPivotLowPoint 查找枢轴低点(Find Pivot Low Point)
// 参数：
//   - klineData: K线数据集合
//   - index: 当前K线索引
//   - period: 查找周期
//
// 返回值：
//   - float64: 枢轴低点价格
//
// 说明：
//
//	在指定周期内查找局部最低点，通过比较前后period个周期的低点来确定
//
// 示例：
//
//	low := FindPivotLowPoint(klineData, index, 10)
//	if !math.IsNaN(low) {
//	    fmt.Printf("找到枢轴低点：%.2f\n", low)
//	}
func FindPivotLowPoint(klineData KlineDatas, index, period int) float64 {
	if index < period || index+period >= len(klineData) {
		return math.NaN()
	}
	for i := index - period; i <= index+period; i++ {
		if klineData[i].Low < klineData[index].Low {
			return math.NaN()
		}
	}
	return klineData[index].Low
}

// CalculateSuperTrendPivot 计算基于枢轴点的超级趋势指标(Calculate SuperTrend Pivot)
// 参数：
//   - klineData: K线数据集合
//   - pivotPeriod: 枢轴点计算周期
//   - factor: ATR乘数
//   - atrPeriod: ATR计算周期
//
// 返回值：
//   - *TaSuperTrendPivot: 超级趋势指标结构体指针
//   - error: 错误信息
//
// 说明：
//
//	结合枢轴点和ATR来计算超级趋势指标，提供更准确的趋势判断
//
// 示例：
//
//	stp, err := CalculateSuperTrendPivot(klineData, 10, 3, 14)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前趋势：%v\n", stp.IsUp())
func CalculateSuperTrendPivot(klineData KlineDatas, pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {
	// 数据长度检查
	dataLen := len(klineData)
	if dataLen < pivotPeriod || dataLen < atrPeriod {
		return nil, fmt.Errorf("计算数据不足: 数据长度%d, 需要轴点周期%d和ATR周期%d", dataLen, pivotPeriod, atrPeriod)
	}

	// 初始化结果数组
	trendUp := make([]float64, dataLen)
	trendDown := make([]float64, dataLen)
	trend := make([]int, dataLen)

	// 计算ATR
	atr, err := klineData.ATR(atrPeriod)
	if err != nil {
		return nil, fmt.Errorf("计算ATR失败: %v", err)
	}

	// 初始化中心价格
	var center float64
	var centerCount int

	// 计算趋势
	for i := pivotPeriod; i < dataLen; i++ {
		// 寻找轴点
		pivotHigh := FindPivotHighPoint(klineData, i, pivotPeriod)
		pivotLow := FindPivotLowPoint(klineData, i, pivotPeriod)

		// 更新中心价格
		if !math.IsNaN(pivotHigh) || !math.IsNaN(pivotLow) {
			newCenter := 0.0
			if !math.IsNaN(pivotHigh) && !math.IsNaN(pivotLow) {
				// 同时存在高点和低点，取平均值
				newCenter = (pivotHigh + pivotLow) / 2
			} else if !math.IsNaN(pivotHigh) {
				newCenter = pivotHigh
			} else {
				newCenter = pivotLow
			}

			// 使用指数平滑方式更新中心价格
			if centerCount == 0 {
				center = newCenter
			} else {
				center = (center*2 + newCenter) / 3
			}
			centerCount++
		}

		// 如果还没有找到有效的中心价格，使用当前价格的中点
		if centerCount == 0 {
			center = (klineData[i].High + klineData[i].Low) / 2
		}

		// 计算轨道带
		band := factor * atr.Values[i]
		upperBand := center + band
		lowerBand := center - band

		// 更新趋势线
		if i > 0 {
			// 上轨趋势线
			if klineData[i-1].Close > trendUp[i-1] {
				trendUp[i] = math.Max(lowerBand, trendUp[i-1])
			} else {
				trendUp[i] = lowerBand
			}

			// 下轨趋势线
			if klineData[i-1].Close < trendDown[i-1] {
				trendDown[i] = math.Min(upperBand, trendDown[i-1])
			} else {
				trendDown[i] = upperBand
			}

			// 确定趋势方向
			if klineData[i].Close > trendDown[i-1] {
				trend[i] = 1 // 上涨趋势
			} else if klineData[i].Close < trendUp[i-1] {
				trend[i] = -1 // 下跌趋势
			} else {
				trend[i] = trend[i-1] // 保持前一趋势
			}
		} else {
			// 初始值设置
			trendUp[i] = lowerBand
			trendDown[i] = upperBand
			trend[i] = 0 // 初始状态设为中性
		}
	}

	return &TaSuperTrendPivot{
		Upper:       trendDown, // 注意：这里使用trendDown作为上轨，因为它代表阻力位
		Lower:       trendUp,   // 使用trendUp作为下轨，因为它代表支撑位
		Trend:       trend,
		PivotPeriod: pivotPeriod,
		Factor:      factor,
		AtrPeriod:   atrPeriod,
	}, nil
}

// SuperTrendPivot 计算K线数据的SuperTrendPivot指标(Calculate SuperTrend Pivot from Kline)
// 参数：
//   - pivotPeriod: 轴点周期，通常为5-10
//   - factor: ATR乘数，通常为2-3
//   - atrPeriod: ATR计算周期，通常为14-21
//
// 返回值：
//   - *TaSuperTrendPivot: SuperTrendPivot指标结构体指针
//   - error: 错误信息
//
// 示例：
//
//	stp, err := kline.SuperTrendPivot(10, 3, 14)
//	if err != nil {
//	    return err
//	}
func (k *KlineDatas) SuperTrendPivot(pivotPeriod int, factor float64, atrPeriod int) (*TaSuperTrendPivot, error) {
	return CalculateSuperTrendPivot(*k, pivotPeriod, factor, atrPeriod)
}

// SuperTrendPivot_IsUp 判断当前是否为上涨趋势(Check Uptrend)
// 参数：
//   - pivotPeriod: 轴点周期
//   - factor: ATR乘数
//   - atrPeriod: ATR计算周期
//
// 返回值：
//   - bool: 是否处于上涨趋势
//
// 示例：
//
//	if kline.SuperTrendPivot_IsUp(10, 3, 14) {
//	    fmt.Println("当前处于上涨趋势")
//	}
func (k *KlineDatas) SuperTrendPivot_IsUp(pivotPeriod int, factor float64, atrPeriod int) bool {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return false
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return false
	}
	return superTrend.Trend[len(superTrend.Trend)-1] == 1
}

// SuperTrendPivot_IsDown 判断当前是否为下跌趋势(Check Downtrend)
// 参数：
//   - pivotPeriod: 轴点周期
//   - factor: ATR乘数
//   - atrPeriod: ATR计算周期
//
// 返回值：
//   - bool: 是否处于下跌趋势
//
// 示例：
//
//	if kline.SuperTrendPivot_IsDown(10, 3, 14) {
//	    fmt.Println("当前处于下跌趋势")
//	}
func (k *KlineDatas) SuperTrendPivot_IsDown(pivotPeriod int, factor float64, atrPeriod int) bool {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return false
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return false
	}
	return superTrend.Trend[len(superTrend.Trend)-1] == -1
}

// SuperTrendPivot_GetUpper 获取最新的上轨线值(Get Latest Upper Band)
// 参数：
//   - pivotPeriod: 轴点周期
//   - factor: ATR乘数
//   - atrPeriod: ATR计算周期
//
// 返回值：
//   - float64: 最新的上轨线值
//
// 示例：
//
//	upper := kline.SuperTrendPivot_GetUpper(10, 3, 14)
func (k *KlineDatas) SuperTrendPivot_GetUpper(pivotPeriod int, factor float64, atrPeriod int) float64 {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return -1
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return -1
	}
	return superTrend.Upper[len(superTrend.Upper)-1]
}

// SuperTrendPivot_GetLower 获取最新的下轨线值(Get Latest Lower Band)
// 参数：
//   - pivotPeriod: 轴点周期
//   - factor: ATR乘数
//   - atrPeriod: ATR计算周期
//
// 返回值：
//   - float64: 最新的下轨线值
//
// 示例：
//
//	lower := kline.SuperTrendPivot_GetLower(10, 3, 14)
func (k *KlineDatas) SuperTrendPivot_GetLower(pivotPeriod int, factor float64, atrPeriod int) float64 {
	_k, err := k.Keep(atrPeriod * 14)
	if err != nil {
		return -1
	}
	superTrend, err := _k.SuperTrendPivot(pivotPeriod, factor, atrPeriod)
	if err != nil {
		return -1
	}
	return superTrend.Lower[len(superTrend.Lower)-1]
}

// Value 返回最新的SuperTrendPivot值(Get Latest Values)
// 返回值：
//   - upper: 最新上轨线值
//   - lower: 最新下轨线值
//   - trend: 趋势方向(1上涨，-1下跌，0横盘)
//
// 说明：
//
//	获取计算结果中的最新值
//
// 示例：
//
//	upper, lower, trend := stp.Value()
func (t *TaSuperTrendPivot) Value() (upper, lower float64, trend int) {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex], t.Lower[lastIndex], t.Trend[lastIndex]
}

// IsUp 判断最新趋势是否为上涨(Check Uptrend)
// 返回值：
//   - bool: 是否处于上涨趋势
//
// 说明：
//
//	判断当前是否处于上涨趋势
//
// 示例：
//
//	if stp.IsUp() {
//	    fmt.Println("当前处于上涨趋势")
//	}
func (t *TaSuperTrendPivot) IsUp() bool {
	return t.Trend[len(t.Trend)-1] == 1
}

// IsDown 判断最新趋势是否为下跌(Check Downtrend)
// 返回值：
//   - bool: 是否处于下跌趋势
//
// 说明：
//
//	判断当前是否处于下跌趋势
//
// 示例：
//
//	if stp.IsDown() {
//	    fmt.Println("当前处于下跌趋势")
//	}
func (t *TaSuperTrendPivot) IsDown() bool {
	return t.Trend[len(t.Trend)-1] == -1
}

// IsSideways 判断最新趋势是否为横盘(Check Sideways)
// 返回值：
//   - bool: 是否处于横盘趋势
//
// 说明：
//
//	判断当前是否处于横盘整理阶段
//
// 示例：
//
//	if stp.IsSideways() {
//	    fmt.Println("当前处于横盘整理")
//	}
func (t *TaSuperTrendPivot) IsSideways() bool {
	return t.Trend[len(t.Trend)-1] == 0
}

// GetUpper 获取最新的上轨线值(Get Latest Upper Band)
// 返回值：
//   - float64: 最新的上轨线值
//
// 说明：
//
//	获取最新计算的上轨线值
//
// 示例：
//
//	upperBand := stp.GetUpper()
func (t *TaSuperTrendPivot) GetUpper() float64 {
	return t.Upper[len(t.Upper)-1]
}

// GetLower 获取最新的下轨线值(Get Latest Lower Band)
// 返回值：
//   - float64: 最新的下轨线值
//
// 说明：
//
//	获取最新计算的下轨线值
//
// 示例：
//
//	lowerBand := stp.GetLower()
func (t *TaSuperTrendPivot) GetLower() float64 {
	return t.Lower[len(t.Lower)-1]
}

// IsTrendChange 判断是否发生趋势转换(Check Trend Change)
// 返回值：
//   - bool: 是否发生趋势转换
//
// 说明：
//
//	判断最近两个周期的趋势是否发生改变
//
// 示例：
//
//	if stp.IsTrendChange() {
//	    fmt.Println("趋势发生转换")
//	}
func (t *TaSuperTrendPivot) IsTrendChange() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex] != t.Trend[lastIndex-1]
}

// IsBullishCross 判断是否发生多头趋势转换(Check Bullish Cross)
// 返回值：
//   - bool: 是否发生多头趋势转换
//
// 说明：
//
//	判断是否从空头趋势转换为多头趋势
//
// 示例：
//
//	if stp.IsBullishCross() {
//	    fmt.Println("转换为多头趋势")
//	}
func (t *TaSuperTrendPivot) IsBullishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex-1] <= 0 && t.Trend[lastIndex] == 1
}

// IsBearishCross 判断是否发生空头趋势转换(Check Bearish Cross)
// 返回值：
//   - bool: 是否发生空头趋势转换
//
// 说明：
//
//	判断是否从多头趋势转换为空头趋势
//
// 示例：
//
//	if stp.IsBearishCross() {
//	    fmt.Println("转换为空头趋势")
//	}
func (t *TaSuperTrendPivot) IsBearishCross() bool {
	if len(t.Trend) < 2 {
		return false
	}
	lastIndex := len(t.Trend) - 1
	return t.Trend[lastIndex-1] >= 0 && t.Trend[lastIndex] == -1
}

// GetTrendStrength 获取趋势强度(Get Trend Strength)
// 返回值：
//   - float64: 趋势强度值
//
// 说明：
//
//	计算当前趋势的强度，值越大表示趋势越强
//
// 示例：
//
//	strength := stp.GetTrendStrength()
//	fmt.Printf("当前趋势强度：%.2f\n", strength)
func (t *TaSuperTrendPivot) GetTrendStrength() float64 {
	lastIndex := len(t.Upper) - 1
	if t.Trend[lastIndex] == 1 {
		return t.Upper[lastIndex] - t.Lower[lastIndex]
	} else if t.Trend[lastIndex] == -1 {
		return t.Lower[lastIndex] - t.Upper[lastIndex]
	}
	return 0
}

// IsTrendStrengthening 判断趋势是否增强(Check if Trend is Strengthening)
// 返回值：
//   - bool: 趋势是否正在增强
//
// 说明：
//
//	通过比较最近几个周期的趋势强度变化判断
//
// 示例：
//
//	if stp.IsTrendStrengthening() {
//	    fmt.Println("趋势正在增强")
//	}
func (t *TaSuperTrendPivot) IsTrendStrengthening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength > previousStrength
}

// IsTrendWeakening 判断趋势是否减弱(Check if Trend is Weakening)
// 返回值：
//   - bool: 趋势是否正在减弱
//
// 说明：
//
//	通过比较最近几个周期的趋势强度变化判断
//
// 示例：
//
//	if stp.IsTrendWeakening() {
//	    fmt.Println("趋势正在减弱")
//	}
func (t *TaSuperTrendPivot) IsTrendWeakening() bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	lastIndex := len(t.Upper) - 1
	currentStrength := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousStrength := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return currentStrength < previousStrength
}

// GetTrendDuration 获取当前趋势持续时间(Get Trend Duration)
// 返回值：
//   - int: 趋势持续的周期数
//
// 说明：
//
//	计算当前趋势已经持续的周期数
//
// 示例：
//
//	duration := stp.GetTrendDuration()
//	fmt.Printf("当前趋势已持续%d个周期\n", duration)
func (t *TaSuperTrendPivot) GetTrendDuration() int {
	if len(t.Trend) < 2 {
		return 0
	}
	lastIndex := len(t.Trend) - 1
	currentTrend := t.Trend[lastIndex]
	duration := 1
	for i := lastIndex - 1; i >= 0; i-- {
		if t.Trend[i] != currentTrend {
			break
		}
		duration++
	}
	return duration
}

// GetBandwidth 获取带宽(Get Bandwidth)
// 返回值：
//   - float64: 上下轨之间的带宽
//
// 说明：
//
//	计算上下轨之间的价格区间大小
//
// 示例：
//
//	bandwidth := stp.GetBandwidth()
//	fmt.Printf("当前带宽：%.2f\n", bandwidth)
func (t *TaSuperTrendPivot) GetBandwidth() float64 {
	lastIndex := len(t.Upper) - 1
	return t.Upper[lastIndex] - t.Lower[lastIndex]
}

// IsBreakoutPossible 判断是否可能发生突破(Check Possible Breakout)
// 参数：
//   - threshold: 突破阈值(可选)
//
// 返回值：
//   - bool: 是否可能发生突破
//
// 说明：
//
//	通过分析价格与轨道的关系判断是否可能发生趋势突破
//
// 示例：
//
//	if stp.IsBreakoutPossible(0.1) {
//	    fmt.Println("可能即将发生趋势突破")
//	}
func (t *TaSuperTrendPivot) IsBreakoutPossible(threshold ...float64) bool {
	if len(t.Upper) < 2 || len(t.Lower) < 2 {
		return false
	}
	th := 0.1
	if len(threshold) > 0 {
		th = threshold[0]
	}
	lastIndex := len(t.Upper) - 1
	currentBandwidth := t.Upper[lastIndex] - t.Lower[lastIndex]
	previousBandwidth := t.Upper[lastIndex-1] - t.Lower[lastIndex-1]
	return math.Abs(currentBandwidth-previousBandwidth)/previousBandwidth > th
}

// GetTrendQuality 获取趋势质量(Get Trend Quality)
// 返回值：
//   - float64: 趋势质量得分，范围[0,1]
//
// 说明：
//
//	评估当前趋势的质量，考虑趋势持续时间、强度等因素
//
// 示例：
//
//	quality := stp.GetTrendQuality()
//	fmt.Printf("当前趋势质量：%.2f\n", quality)
func (t *TaSuperTrendPivot) GetTrendQuality() float64 {
	duration := t.GetTrendDuration()
	strength := t.GetTrendStrength()
	return float64(duration) * strength
}

// IsPivotBreakout 判断是否发生轴点突破(Check Pivot Breakout)
// 参数：
//   - klineData: K线数据集合
//
// 返回值：
//   - bool: 是否发生轴点突破
//
// 说明：
//
//	判断价格是否突破重要的轴点位置
//
// 示例：
//
//	if stp.IsPivotBreakout(klineData) {
//	    fmt.Println("发生轴点突破")
//	}
func (t *TaSuperTrendPivot) IsPivotBreakout(klineData KlineDatas) bool {
	if len(t.Trend) < t.PivotPeriod {
		return false
	}
	lastIndex := len(t.Trend) - 1
	pivotHigh := FindPivotHighPoint(klineData, lastIndex, t.PivotPeriod)
	pivotLow := FindPivotLowPoint(klineData, lastIndex, t.PivotPeriod)

	if !math.IsNaN(pivotHigh) && klineData[lastIndex].Close > pivotHigh {
		return true
	}
	if !math.IsNaN(pivotLow) && klineData[lastIndex].Close < pivotLow {
		return true
	}
	return false
}

// GetPivotStrength 获取轴点强度(Get Pivot Strength)
// 参数：
//   - klineData: K线数据集合
//
// 返回值：
//   - float64: 轴点强度值
//
// 说明：
//
//	计算当前轴点的强度，用于评估支撑或阻力的可靠性
//
// 示例：
//
//	strength := stp.GetPivotStrength(klineData)
//	fmt.Printf("当前轴点强度：%.2f\n", strength)
func (t *TaSuperTrendPivot) GetPivotStrength(klineData KlineDatas) float64 {
	if len(t.Trend) < t.PivotPeriod {
		return 0
	}
	lastIndex := len(t.Trend) - 1
	pivotHigh := FindPivotHighPoint(klineData, lastIndex, t.PivotPeriod)
	pivotLow := FindPivotLowPoint(klineData, lastIndex, t.PivotPeriod)

	if !math.IsNaN(pivotHigh) && !math.IsNaN(pivotLow) {
		return pivotHigh - pivotLow
	}
	return 0
}
