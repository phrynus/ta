package ta

import (
	"fmt"
	"math"
)

// Package ta 提供技术分析指标的计算功能

// TaVolatilityRatio 波动率比率指标结构体(Volatility Ratio)
// 说明：
//
//	波动率比率是一个衡量市场波动性的技术指标
//	它通过比较当前波动率与历史波动率来判断市场状态
//	主要应用场景：
//	1. 波动性分析：识别市场波动是否处于高位或低位
//	2. 趋势强度：评估当前趋势的强弱程度
//	3. 市场状态：判断市场是否处于盘整或突破状态
//
// 计算公式：
//  1. 计算N周期的真实波幅(TR)
//  2. 计算当前波动率和历史波动率
//  3. 波动率比率 = 当前波动率 / 历史波动率
type TaVolatilityRatio struct {
	Values []float64 `json:"values"` // 波动率比率值序列
	Period int       `json:"period"` // 计算周期
}

// CalculateVolatilityRatio 计算波动率比率
// 参数：
//   - prices: 价格序列
//   - period: 计算周期
//
// 返回值：
//   - *TaVolatilityRatio: 波动率比率指标结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	波动率比率通过比较不同时期的波动性来判断市场状态
//	计算步骤：
//	1. 计算每个周期的真实波幅
//	2. 计算当前周期和历史周期的平均波动率
//	3. 计算波动率比率
//
// 示例：
//
//	vr, err := CalculateVolatilityRatio(prices, 14)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("当前波动率比率：%.2f\n", vr.Values[len(vr.Values)-1])

// CalculateVolatilityRatio 计算波动比率指标(Volatility Ratio)
// 参数：
//   - klineData: K线数据集合
//   - shortPeriod: 短期周期，通常为1
//   - longPeriod: 长期周期，通常为10
//
// 返回值：
//   - *TaVolatilityRatio: 包含波动比率值的结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	波动比率是一个用于衡量市场波动强度变化的技术指标
//	计算步骤：
//	1. 计算短期真实波幅(TR_short)
//	2. 计算长期真实波幅的移动平均(TR_long_MA)
//	3. 计算波动比率 = TR_short / TR_long_MA
//	4. 当比率>1时表示波动加剧，<1时表示波动减弱
//
// 示例：
//
//	vr, err := CalculateVolatilityRatio(klineData, 1, 10)
func CalculateVolatilityRatio(klineData KlineDatas, shortPeriod, longPeriod int) (*TaVolatilityRatio, error) {
	if len(klineData) < longPeriod {
		return nil, fmt.Errorf("计算数据不足")
	}

	length := len(klineData)
	// 预分配所有需要的切片
	slices := preallocateSlices(length, 2) // [trueRange, ratio]
	trueRange, ratio := slices[0], slices[1]

	// 计算真实波幅
	for i := 1; i < length; i++ {
		high := klineData[i].High
		low := klineData[i].Low
		prevClose := klineData[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)
		trueRange[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	// 使用滑动窗口计算波动比率
	for i := longPeriod; i < length; i++ {
		// 计算短期TR
		var shortTR float64
		for j := i - shortPeriod + 1; j <= i; j++ {
			shortTR += trueRange[j]
		}
		shortTR /= float64(shortPeriod)

		// 计算长期TR
		var longTR float64
		for j := i - longPeriod + 1; j <= i; j++ {
			longTR += trueRange[j]
		}
		longTR /= float64(longPeriod)

		// 计算波动比率
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

// VolatilityRatio 计算K线数据的波动率比率
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - *TaVolatilityRatio: 波动率比率指标结构体指针
//   - error: 可能的错误信息
//
// 说明：
//
//	波动率比率可用于：
//	1. 判断市场是否处于高波动或低波动状态
//	2. 预测可能的趋势变化
//	3. 识别市场突破机会
//
// 示例：
//
//	vr, err := kline.VolatilityRatio(14, "close")
//	if err != nil {
//	    return err
//	}

// VolatilityRatio_ 获取最新的波动率比率值
// 参数：
//   - period: 计算周期
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 最新的波动率比率值
//
// 示例：
//
//	value := kline.VolatilityRatio_(14, "close")

// Value 返回最新的波动率比率值
// 返回值：
//   - float64: 最新的波动率比率值
//
// 示例：
//
//	value := vr.Value()

// IsHighVolatility 判断是否处于高波动状态
// 参数：
//   - threshold: 高波动阈值，默认为1.5
//
// 返回值：
//   - bool: 如果波动率比率超过阈值返回true，否则返回false
//
// 说明：
//
//	高波动通常意味着市场可能出现大幅波动
//
// 示例：
//
//	isHigh := vr.IsHighVolatility(1.5)

// IsLowVolatility 判断是否处于低波动状态
// 参数：
//   - threshold: 低波动阈值，默认为0.5
//
// 返回值：
//   - bool: 如果波动率比率低于阈值返回true，否则返回false
//
// 说明：
//
//	低波动通常意味着市场可能即将突破
//
// 示例：
//
//	isLow := vr.IsLowVolatility(0.5)

// IsIncreasing 判断波动率是否在增加
// 返回值：
//   - bool: 如果波动率在增加返回true，否则返回false
//
// 说明：
//
//	波动率增加可能预示市场即将出现大幅波动
//
// 示例：
//
//	isIncreasing := vr.IsIncreasing()

// IsDecreasing 判断波动率是否在减少
// 返回值：
//   - bool: 如果波动率在减少返回true，否则返回false
//
// 说明：
//
//	波动率减少可能预示市场即将进入盘整期
//
// 示例：
//
//	isDecreasing := vr.IsDecreasing()

// GetTrend 获取波动率趋势
// 返回值：
//   - int: 趋势值，1=上升，0=盘整，-1=下降
//
// 说明：
//
//	通过比较多个周期的波动率来判断趋势方向
//
// 示例：
//
//	trend := vr.GetTrend()

// GetStrength 获取波动强度
// 返回值：
//   - float64: 波动强度，相对于1的偏离程度
//
// 说明：
//
//	波动强度反映了当前波动相对于正常水平的程度
//
// 示例：
//
//	strength := vr.GetStrength()

// IsDiverging 判断是否与价格出现背离
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - bool: 如果出现背离返回true，否则返回false
//
// 说明：
//
//	波动率与价格的背离可能预示趋势即将改变
//
// 示例：
//
//	isDiverging := vr.IsDiverging(prices)

// GetOptimalPeriod 获取最优计算周期
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - int: 建议的最优计算周期
//
// 说明：
//
//	根据价格波动特征自动计算最优的计算周期
//
// 示例：
//
//	optPeriod := vr.GetOptimalPeriod(prices)

// IsBreakoutLikely 判断是否可能出现突破
// 返回值：
//   - bool: 如果可能出现突破返回true，否则返回false
//
// 说明：
//
//	通过分析波动率的变化模式来预测可能的突破
//
// 示例：
//
//	isLikely := vr.IsBreakoutLikely()

// GetVolatilityZone 获取波动区间
// 返回值：
//   - string: 波动区间，"high"=高波动，"normal"=正常，"low"=低波动
//
// 说明：
//
//	根据波动率的位置判断当前所处的波动区间
//
// 示例：
//
//	zone := vr.GetVolatilityZone()

// IsConsolidating 判断是否处于盘整状态
// 返回值：
//   - bool: 如果处于盘整状态返回true，否则返回false
//
// 说明：
//
//	通过分析波动率的稳定性来判断是否处于盘整
//
// 示例：
//
//	isConsolidating := vr.IsConsolidating()

// GetVolatilityScore 获取波动性评分
// 返回值：
//   - float64: 波动性评分，0-100之间
//
// 说明：
//
//	将波动率转换为0-100之间的评分，便于比较
//
// 示例：
//
//	score := vr.GetVolatilityScore()
