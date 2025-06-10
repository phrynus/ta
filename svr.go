package ta

import (
	"fmt"
	"math"
)

// TaSVR SVR指标结构体(Support Vector Regression)
type TaSVR struct {
	Values    []float64 `json:"values"`     // SVR预测值序列
	UpperBand []float64 `json:"upper_band"` // 上轨带
	LowerBand []float64 `json:"lower_band"` // 下轨带
	Period    int       `json:"period"`     // 计算周期
	Kernel    string    `json:"kernel"`     // 核函数类型
	Epsilon   float64   `json:"epsilon"`    // ε参数
	C         float64   `json:"c"`          // 惩罚参数
	Bandwidth float64   `json:"bandwidth"`  // 核函数带宽
}

// kernelRBF 径向基核函数(Radial Basis Function)
func kernelRBF(x1, x2 []float64, bandwidth float64) float64 {
	var sum float64
	for i := range x1 {
		diff := x1[i] - x2[i]
		sum += diff * diff
	}
	return math.Exp(-sum / (2 * bandwidth * bandwidth))
}

// kernelLinear 线性核函数
func kernelLinear(x1, x2 []float64, _ float64) float64 {
	var sum float64
	for i := range x1 {
		sum += x1[i] * x2[i]
	}
	return sum
}

// createFeatureVector 创建特征向量
func createFeatureVector(data []float64, period int, index int) []float64 {
	feature := make([]float64, period)
	for i := 0; i < period; i++ {
		if index-i >= 0 {
			feature[period-1-i] = data[index-i]
		}
	}
	return feature
}

// CalculateSVR 计算支持向量回归
// 参数：
//   - prices: 价格序列
//   - period: 计算周期
//   - kernel: 核函数类型 ("rbf" 或 "linear")
//   - epsilon: ε参数，允许的误差范围
//   - c: 惩罚参数
//   - bandwidth: 核函数带宽(仅用于RBF核)
//
// 返回值：
//   - *TaSVR: SVR指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	SVR是一种回归分析方法，用于预测时间序列的未来走势
//	计算步骤：
//	1. 构建特征向量和目标值
//	2. 使用核函数计算核矩阵
//	3. 求解对偶问题得到支持向量
//	4. 计算预测值和置信区间
//
// 示例：
//
//	svr, err := CalculateSVR(prices, 14, "rbf", 0.1, 1.0, 0.5)
func CalculateSVR(prices []float64, period int, kernel string, epsilon, c, bandwidth float64) (*TaSVR, error) {
	length := len(prices)
	if length < period {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 预分配切片
	slices := preallocateSlices(length, 3) // [values, upperBand, lowerBand]
	values, upperBand, lowerBand := slices[0], slices[1], slices[2]

	// 选择核函数
	var kernelFunc func([]float64, []float64, float64) float64
	switch kernel {
	case "rbf":
		kernelFunc = kernelRBF
	case "linear":
		kernelFunc = kernelLinear
	default:
		return nil, fmt.Errorf("不支持的核函数类型: %s", kernel)
	}

	// 对每个时间点进行预测
	for i := period; i < length; i++ {
		// 构建训练数据
		var trainX [][]float64
		var trainY []float64
		for j := i - period; j < i; j++ {
			feature := createFeatureVector(prices, period, j)
			trainX = append(trainX, feature)
			trainY = append(trainY, prices[j])
		}

		// 计算核矩阵
		n := len(trainX)
		kernelMatrix := make([][]float64, n)
		for j := range kernelMatrix {
			kernelMatrix[j] = make([]float64, n)
			for k := range kernelMatrix[j] {
				kernelMatrix[j][k] = kernelFunc(trainX[j], trainX[k], bandwidth)
			}
		}

		// 简化的SMO算法求解
		alphas := make([]float64, n)
		b := 0.0

		// 迭代优化（简化版）
		for iter := 0; iter < 100; iter++ {
			changed := false
			for j := 0; j < n; j++ {
				var sum float64
				for k := 0; k < n; k++ {
					sum += alphas[k] * kernelMatrix[j][k]
				}
				error := sum + b - trainY[j]

				if (trainY[j]*error < -epsilon && alphas[j] < c) ||
					(trainY[j]*error > epsilon && alphas[j] > -c) {
					alphas[j] = math.Max(-c, math.Min(c,
						alphas[j]-error/kernelMatrix[j][j]))
					changed = true
				}
			}
			if !changed {
				break
			}
		}

		// 预测当前值
		var prediction float64
		testFeature := createFeatureVector(prices, period, i)
		for j := 0; j < n; j++ {
			prediction += alphas[j] * kernelFunc(trainX[j], testFeature, bandwidth)
		}
		prediction += b

		// 计算预测区间
		std := 0.0
		for j := i - period; j < i; j++ {
			diff := prices[j] - values[j]
			std += diff * diff
		}
		std = math.Sqrt(std / float64(period))

		values[i] = prediction
		upperBand[i] = prediction + 2*std
		lowerBand[i] = prediction - 2*std
	}

	return &TaSVR{
		Values:    values,
		UpperBand: upperBand,
		LowerBand: lowerBand,
		Period:    period,
		Kernel:    kernel,
		Epsilon:   epsilon,
		C:         c,
		Bandwidth: bandwidth,
	}, nil
}

// SVR 计算K线数据的SVR指标
// 参数：
//   - source: 数据来源，可以是 "open"、"high"、"low"、"close"、"volume"
//   - period: 计算周期
//   - kernel: 核函数类型
//   - epsilon: ε参数
//   - c: 惩罚参数
//   - bandwidth: 核函数带宽
//
// 返回值：
//   - *TaSVR: SVR指标结构体
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	svr, err := kline.SVR("close", 14, "rbf", 0.1, 1.0, 0.5)
func (k *KlineDatas) SVR(source string, period int, kernel string, epsilon, c, bandwidth float64) (*TaSVR, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateSVR(prices, period, kernel, epsilon, c, bandwidth)
}

// SVR_ 获取最新的SVR预测值
// 参数：
//   - source: 数据来源
//   - period: 计算周期
//   - kernel: 核函数类型
//   - epsilon: ε参数
//   - c: 惩罚参数
//   - bandwidth: 核函数带宽
//
// 返回值：
//   - prediction: 预测值
//   - upper: 上轨带值
//   - lower: 下轨带值
//
// 示例：
//
//	pred, upper, lower := kline.SVR_("close", 14, "rbf", 0.1, 1.0, 0.5)
func (k *KlineDatas) SVR_(source string, period int, kernel string, epsilon, c, bandwidth float64) (prediction, upper, lower float64) {
	_k, err := k.Keep(period * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0, 0
	}
	svr, err := CalculateSVR(prices, period, kernel, epsilon, c, bandwidth)
	if err != nil {
		return 0, 0, 0
	}
	return svr.Value()
}

// Value 返回最新的SVR预测值和区间
// 返回值：
//   - prediction: 预测值
//   - upper: 上轨带值
//   - lower: 下轨带值
//
// 示例：
//
//	pred, upper, lower := svr.Value()
func (t *TaSVR) Value() (prediction, upper, lower float64) {
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex], t.UpperBand[lastIndex], t.LowerBand[lastIndex]
}

// IsTrendUp 判断趋势是否向上
// 返回值：
//   - bool: 如果趋势向上返回true，否则返回false
//
// 示例：
//
//	isUp := svr.IsTrendUp()
func (t *TaSVR) IsTrendUp() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] > t.Values[lastIndex-1]
}

// IsTrendDown 判断趋势是否向下
// 返回值：
//   - bool: 如果趋势向下返回true，否则返回false
//
// 示例：
//
//	isDown := svr.IsTrendDown()
func (t *TaSVR) IsTrendDown() bool {
	if len(t.Values) < 2 {
		return false
	}
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex] < t.Values[lastIndex-1]
}

// GetConfidenceInterval 获取预测的置信区间宽度
// 返回值：
//   - float64: 置信区间宽度
//
// 示例：
//
//	width := svr.GetConfidenceInterval()
func (t *TaSVR) GetConfidenceInterval() float64 {
	lastIndex := len(t.Values) - 1
	return t.UpperBand[lastIndex] - t.LowerBand[lastIndex]
}

// IsOutsideBands 判断价格是否超出预测区间
// 参数：
//   - price: 当前价格
//
// 返回值：
//   - above: 是否突破上轨
//   - below: 是否跌破下轨
//
// 示例：
//
//	above, below := svr.IsOutsideBands(price)
func (t *TaSVR) IsOutsideBands(price float64) (above, below bool) {
	lastIndex := len(t.Values) - 1
	above = price > t.UpperBand[lastIndex]
	below = price < t.LowerBand[lastIndex]
	return
}
