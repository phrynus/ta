package ta

import (
	"fmt"
	"math"
)

// TaDeepTS 深度学习时间序列指标结构体(Deep Learning Time Series)
type TaDeepTS struct {
	Values     []float64   `json:"values"`      // 预测值序列
	ModelType  string      `json:"model_type"`  // 模型类型 (LSTM/GRU/TCN)
	WindowSize int         `json:"window_size"` // 时间窗口
	Hidden     [][]float64 `json:"hidden"`      // 隐藏状态
	Weights    [][]float64 `json:"weights"`     // 模型权重
}

// 门控单元结构
type gateUnit struct {
	input  []float64
	forget []float64
	cell   []float64
	output []float64
}

// sigmoid 激活函数
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// tanh 激活函数
func tanh(x float64) float64 {
	return math.Tanh(x)
}

// matrixMultiply 矩阵乘法
func matrixMultiply(a, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

// CalculateDeepTS 计算深度学习时间序列预测
// 参数：
//   - prices: 价格序列
//   - modelType: 模型类型 ("lstm", "gru" 或 "tcn")
//   - windowSize: 时间窗口大小 10-20
//   - hiddenSize: 隐藏层大小 32-128
//
// 返回值：
//   - *TaDeepTS: DeepTS指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	DeepTS使用深度学习模型进行时间序列预测
//	支持的模型类型：
//	- LSTM: 长短期记忆网络，适合处理长期依赖
//	- GRU: 门控循环单元，LSTM的简化版本
//	- TCN: 时间卷积网络，适合并行计算和长期预测
//
// 示例：
//
//	deepts, err := CalculateDeepTS(prices, "tcn", 10, 32)
func CalculateDeepTS(prices []float64, modelType string, windowSize, hiddenSize int) (*TaDeepTS, error) {
	length := len(prices)
	if length < windowSize {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 预分配切片
	slices := preallocateSlices(length, 2) // [values, hidden[0]]
	values, hidden := slices[0], slices[1]

	// 初始化权重
	weights := make([][]float64, 4)
	for i := range weights {
		weights[i] = make([]float64, hiddenSize)
		for j := range weights[i] {
			weights[i][j] = 0.1
		}
	}

	// 初始化状态
	hiddenState := make([]float64, hiddenSize)

	// 对每个时间点进行预测
	for i := windowSize; i < length; i++ {
		// 准备输入窗口
		window := make([]float64, windowSize)
		copy(window, prices[i-windowSize:i])

		switch modelType {
		case "lstm":
			// LSTM计算
			cellState := make([]float64, hiddenSize)
			gates := gateUnit{
				input:  make([]float64, hiddenSize),
				forget: make([]float64, hiddenSize),
				cell:   make([]float64, hiddenSize),
				output: make([]float64, hiddenSize),
			}

			for j := 0; j < hiddenSize; j++ {
				gates.input[j] = sigmoid(matrixMultiply(window, weights[0]))
				gates.forget[j] = sigmoid(matrixMultiply(window, weights[1]))
				gates.cell[j] = tanh(matrixMultiply(window, weights[2]))
				gates.output[j] = sigmoid(matrixMultiply(window, weights[3]))

				cellState[j] = gates.forget[j]*cellState[j] + gates.input[j]*gates.cell[j]
				hiddenState[j] = gates.output[j] * tanh(cellState[j])
			}

		case "gru":
			// GRU计算
			resetGate := make([]float64, hiddenSize)
			updateGate := make([]float64, hiddenSize)
			newMemory := make([]float64, hiddenSize)

			for j := 0; j < hiddenSize; j++ {
				resetGate[j] = sigmoid(matrixMultiply(window, weights[0]))
				updateGate[j] = sigmoid(matrixMultiply(window, weights[1]))
				newMemory[j] = tanh(matrixMultiply(window, weights[2]))
				hiddenState[j] = updateGate[j]*hiddenState[j] + (1-updateGate[j])*newMemory[j]
			}

		case "tcn":
			// TCN计算
			// 使用扩张卷积进行时间序列建模
			dilations := []int{1, 2, 4, 8} // 扩张率序列
			tempHidden := make([]float64, hiddenSize)

			for d, dilation := range dilations {
				for j := 0; j < hiddenSize; j++ {
					// 计算扩张卷积
					var conv float64
					for k := 0; k < windowSize; k++ {
						if k*dilation < windowSize {
							conv += window[k*dilation] * weights[d][j]
						}
					}
					// 残差连接
					if d == 0 {
						tempHidden[j] = tanh(conv)
					} else {
						tempHidden[j] = tanh(conv + hiddenState[j])
					}
				}
				copy(hiddenState, tempHidden)
			}

		default:
			return nil, fmt.Errorf("不支持的模型类型: %s", modelType)
		}

		// 生成预测值
		var prediction float64
		for j := 0; j < hiddenSize; j++ {
			prediction += hiddenState[j]
		}
		values[i] = prediction / float64(hiddenSize)
		hidden[i] = hiddenState[0]
	}

	return &TaDeepTS{
		Values:     values,
		ModelType:  modelType,
		WindowSize: windowSize,
		Hidden:     [][]float64{hidden},
		Weights:    weights,
	}, nil
}

// DeepTS 计算K线数据的DeepTS指标
// 参数：
//   - source: 数据来源
//   - modelType: 模型类型 ("lstm" 或 "gru")
//   - windowSize: 时间窗口大小 10-20
//   - hiddenSize: 隐藏层大小 32-128
//
// 返回值：
//   - *TaDeepTS: DeepTS指标结构体
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	deepts, err := kline.DeepTS("close", "lstm", 10, 32)
func (k *KlineDatas) DeepTS(source string, modelType string, windowSize, hiddenSize int) (*TaDeepTS, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateDeepTS(prices, modelType, windowSize, hiddenSize)
}

// DeepTS_ 获取最新的DeepTS预测值
// 参数：
//   - source: 数据来源
//   - modelType: 模型类型
//   - windowSize: 时间窗口大小
//   - hiddenSize: 隐藏层大小
//
// 返回值：
//   - float64: 预测值
//   - float64: 隐藏状态值
//
// 示例：
//
//	pred, hidden := kline.DeepTS_("close", "lstm", 10, 32)
func (k *KlineDatas) DeepTS_(source string, modelType string, windowSize, hiddenSize int) (prediction, hidden float64) {
	_k, err := k.Keep(windowSize * 12)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0
	}
	deepts, err := CalculateDeepTS(prices, modelType, windowSize, hiddenSize)
	if err != nil {
		return 0, 0
	}
	return deepts.Value()
}

// Value 返回最新的预测值和隐藏状态
// 返回值：
//   - float64: 预测值
//   - float64: 隐藏状态值
//
// 示例：
//
//	pred, hidden := deepts.Value()
func (t *TaDeepTS) Value() (prediction, hidden float64) {
	lastIndex := len(t.Values) - 1
	return t.Values[lastIndex], t.Hidden[0][lastIndex]
}

// GetPredictionTrend 获取预测趋势
// 返回值：
//   - int: 趋势值，1=上涨，-1=下跌，0=盘整
//
// 示例：
//
//	trend := deepts.GetPredictionTrend()
func (t *TaDeepTS) GetPredictionTrend() int {
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

// GetConfidence 获取预测置信度
// 返回值：
//   - float64: 置信度值（0-1之间）
//
// 示例：
//
//	confidence := deepts.GetConfidence()
func (t *TaDeepTS) GetConfidence() float64 {
	if len(t.Hidden) == 0 || len(t.Hidden[0]) == 0 {
		return 0
	}
	lastHidden := t.Hidden[0][len(t.Hidden[0])-1]
	return sigmoid(lastHidden) // 使用sigmoid将隐藏状态映射到0-1区间
}

// IsStable 判断预测是否稳定
// 参数：
//   - threshold: 稳定性阈值
//
// 返回值：
//   - bool: 如果预测稳定返回true，否则返回false
//
// 示例：
//
//	isStable := deepts.IsStable(0.1)
func (t *TaDeepTS) IsStable(threshold float64) bool {
	if len(t.Values) < t.WindowSize {
		return false
	}
	lastIndex := len(t.Values) - 1
	var variance float64
	mean := t.Values[lastIndex]
	for i := 0; i < t.WindowSize; i++ {
		diff := t.Values[lastIndex-i] - mean
		variance += diff * diff
	}
	variance /= float64(t.WindowSize)
	return variance < threshold
}
