package ta

import (
	"fmt"
	"math"
	"sort"
)

// TaDTR 决策树回归指标结构体(Decision Tree Regression)
type TaDTR struct {
	Values     []float64 `json:"values"`      // 预测值序列
	WindowSize int       `json:"window_size"` // 时间窗口
	MaxDepth   int       `json:"max_depth"`   // 最大树深度
	MinSplit   int       `json:"min_split"`   // 最小分裂样本数
	TreeNodes  []*Node   `json:"tree_nodes"`  // 树节点列表
}

// Node 决策树节点
type Node struct {
	Feature    int     `json:"feature"`     // 特征索引
	Threshold  float64 `json:"threshold"`   // 分裂阈值
	Value      float64 `json:"value"`       // 叶子节点值
	Left       *Node   `json:"left"`        // 左子树
	Right      *Node   `json:"right"`       // 右子树
	Depth      int     `json:"depth"`       // 节点深度
	SampleSize int     `json:"sample_size"` // 样本数量
}

// CalculateDTR 计算决策树回归指标
// 参数：
//   - prices: 价格序列
//   - windowSize: 时间窗口大小 5-20
//   - maxDepth: 最大树深度 3-10
//   - minSplit: 最小分裂样本数 2-10
//
// 返回值：
//   - *TaDTR: DTR指标结构体
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	DTR使用决策树进行时间序列预测
//	特点：
//	1. 可以处理非线性关系
//	2. 模型可解释性强
//	3. 不需要数据标准化
//	4. 适合发现价格区间突破
//
// 示例：
//
//	dtr, err := CalculateDTR(prices, 10, 5, 3)
func CalculateDTR(prices []float64, windowSize, maxDepth, minSplit int) (*TaDTR, error) {
	length := len(prices)
	if length < windowSize {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 预分配切片
	values := make([]float64, length)

	// 构建训练数据
	var treeNodes []*Node

	// 对每个时间点进行预测
	for i := windowSize; i < length; i++ {
		// 准备训练数据
		X := make([][]float64, windowSize)
		y := make([]float64, windowSize)

		for j := 0; j < windowSize; j++ {
			X[j] = make([]float64, 3) // 使用3个特征：价格、价格变化率、移动平均
			idx := i - windowSize + j
			X[j][0] = prices[idx]
			if idx > 0 {
				X[j][1] = (prices[idx] - prices[idx-1]) / prices[idx-1]
			}
			var ma float64
			for k := 0; k < 5 && idx-k >= 0; k++ {
				ma += prices[idx-k]
			}
			X[j][2] = ma / 5
			y[j] = prices[idx]
		}

		// 构建决策树
		root := buildTree(X, y, 0, maxDepth, minSplit)
		treeNodes = append(treeNodes, root)

		// 使用最新的树进行预测
		testX := []float64{
			prices[i-1],
			(prices[i-1] - prices[i-2]) / prices[i-2],
			(prices[i-1] + prices[i-2] + prices[i-3] + prices[i-4] + prices[i-5]) / 5,
		}
		values[i] = predict(root, testX)
	}

	return &TaDTR{
		Values:     values,
		WindowSize: windowSize,
		MaxDepth:   maxDepth,
		MinSplit:   minSplit,
		TreeNodes:  treeNodes,
	}, nil
}

// buildTree 构建决策树
func buildTree(X [][]float64, y []float64, depth, maxDepth, minSplit int) *Node {
	sampleSize := len(y)
	if sampleSize < minSplit || depth >= maxDepth {
		return &Node{
			Value:      mean(y),
			Depth:      depth,
			SampleSize: sampleSize,
		}
	}

	bestGain := 0.0
	var bestFeature int
	var bestThreshold float64
	var bestLeft, bestRight []int

	// 遍历所有特征和可能的分裂点
	for feature := 0; feature < len(X[0]); feature++ {
		values := make([]float64, sampleSize)
		for i := range X {
			values[i] = X[i][feature]
		}
		sort.Float64s(values)

		// 尝试每个可能的分裂点
		for i := 1; i < sampleSize; i++ {
			threshold := (values[i-1] + values[i]) / 2
			left, right := split(X, y, feature, threshold)
			if len(left) < minSplit || len(right) < minSplit {
				continue
			}

			gain := calculateGain(y, left, right)
			if gain > bestGain {
				bestGain = gain
				bestFeature = feature
				bestThreshold = threshold
				bestLeft = left
				bestRight = right
			}
		}
	}

	// 如果没有找到好的分裂点
	if bestGain == 0 {
		return &Node{
			Value:      mean(y),
			Depth:      depth,
			SampleSize: sampleSize,
		}
	}

	// 递归构建左右子树
	leftX, leftY := getSubset(X, y, bestLeft)
	rightX, rightY := getSubset(X, y, bestRight)

	return &Node{
		Feature:    bestFeature,
		Threshold:  bestThreshold,
		Left:       buildTree(leftX, leftY, depth+1, maxDepth, minSplit),
		Right:      buildTree(rightX, rightY, depth+1, maxDepth, minSplit),
		Depth:      depth,
		SampleSize: sampleSize,
	}
}

// predict 使用决策树进行预测
func predict(node *Node, x []float64) float64 {
	if node.Left == nil && node.Right == nil {
		return node.Value
	}
	if x[node.Feature] <= node.Threshold {
		return predict(node.Left, x)
	}
	return predict(node.Right, x)
}

// calculateGain 计算信息增益
func calculateGain(y []float64, left, right []int) float64 {
	totalVar := variance(y)
	leftY := make([]float64, len(left))
	rightY := make([]float64, len(right))
	for i, idx := range left {
		leftY[i] = y[idx]
	}
	for i, idx := range right {
		rightY[i] = y[idx]
	}
	weightedVar := float64(len(left))*variance(leftY) + float64(len(right))*variance(rightY)
	weightedVar /= float64(len(y))
	return totalVar - weightedVar
}

// split 数据分裂
func split(X [][]float64, y []float64, feature int, threshold float64) ([]int, []int) {
	var left, right []int
	for i := range X {
		if X[i][feature] <= threshold {
			left = append(left, i)
		} else {
			right = append(right, i)
		}
	}
	return left, right
}

// getSubset 获取数据子集
func getSubset(X [][]float64, y []float64, indices []int) ([][]float64, []float64) {
	subX := make([][]float64, len(indices))
	subY := make([]float64, len(indices))
	for i, idx := range indices {
		subX[i] = X[idx]
		subY[i] = y[idx]
	}
	return subX, subY
}

// mean 计算平均值
func mean(x []float64) float64 {
	sum := 0.0
	for _, v := range x {
		sum += v
	}
	return sum / float64(len(x))
}

// variance 计算方差
func variance(x []float64) float64 {
	m := mean(x)
	sum := 0.0
	for _, v := range x {
		d := v - m
		sum += d * d
	}
	return sum / float64(len(x))
}

// DTR 计算K线数据的DTR指标
// 参数：
//   - source: 数据来源
//   - windowSize: 时间窗口大小 5-20
//   - maxDepth: 最大树深度 3-10
//   - minSplit: 最小分裂样本数 2-10
//
// 返回值：
//   - *TaDTR: DTR指标结构体
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	dtr, err := kline.DTR("close", 10, 5, 3)
func (k *KlineDatas) DTR(source string, windowSize, maxDepth, minSplit int) (*TaDTR, error) {
	prices, err := k.ExtractSlice(source)
	if err != nil {
		return nil, err
	}
	return CalculateDTR(prices, windowSize, maxDepth, minSplit)
}

// DTR_ 获取最新的DTR预测值
// 参数：
//   - source: 数据来源
//   - windowSize: 时间窗口大小 5-20
//   - maxDepth: 最大树深度 3-10
//   - minSplit: 最小分裂样本数 2-10
//
// 返回值：
//   - float64: 预测值
//   - float64: 置信度（基于样本数量）
//
// 示例：
//
//	pred, confidence := kline.DTR_("close", 10, 5, 3)
func (k *KlineDatas) DTR_(source string, windowSize, maxDepth, minSplit int) (prediction, confidence float64) {
	_k, err := k.Keep(windowSize * 14)
	if err != nil {
		_k = *k
	}
	prices, err := _k.ExtractSlice(source)
	if err != nil {
		return 0, 0
	}
	dtr, err := CalculateDTR(prices, windowSize, maxDepth, minSplit)
	if err != nil {
		return 0, 0
	}
	return dtr.Value()
}

// Value 返回最新的预测值和置信度
// 返回值：
//   - float64: 预测值
//   - float64: 置信度（基于样本数量）
//
// 示例：
//
//	pred, confidence := dtr.Value()
func (t *TaDTR) Value() (prediction, confidence float64) {
	lastIndex := len(t.Values) - 1
	if lastIndex < 0 || len(t.TreeNodes) == 0 {
		return 0, 0
	}

	// 计算置信度（基于最后一个树节点的样本数量）
	lastNode := t.TreeNodes[len(t.TreeNodes)-1]
	confidence = math.Min(1.0, float64(lastNode.SampleSize)/float64(t.WindowSize*2))

	return t.Values[lastIndex], confidence
}

// GetTreeDepth 获取当前树的实际深度
// 返回值：
//   - int: 树的实际深度
//
// 示例：
//
//	depth := dtr.GetTreeDepth()
func (t *TaDTR) GetTreeDepth() int {
	if len(t.TreeNodes) == 0 {
		return 0
	}
	return getNodeDepth(t.TreeNodes[len(t.TreeNodes)-1])
}

// getNodeDepth 获取节点深度
func getNodeDepth(node *Node) int {
	if node == nil {
		return 0
	}
	leftDepth := getNodeDepth(node.Left)
	rightDepth := getNodeDepth(node.Right)
	if leftDepth > rightDepth {
		return leftDepth + 1
	}
	return rightDepth + 1
}

// GetImportantFeatures 获取重要特征
// 返回值：
//   - map[string]float64: 特征重要性得分
//
// 示例：
//
//	features := dtr.GetImportantFeatures()
func (t *TaDTR) GetImportantFeatures() map[string]float64 {
	if len(t.TreeNodes) == 0 {
		return nil
	}

	features := map[string]float64{
		"price":  0,
		"change": 0,
		"ma":     0,
	}

	// 统计每个特征的使用频率
	countFeatures(t.TreeNodes[len(t.TreeNodes)-1], features)

	// 归一化特征重要性
	total := 0.0
	for _, v := range features {
		total += v
	}
	if total > 0 {
		for k := range features {
			features[k] /= total
		}
	}

	return features
}

// countFeatures 统计特征使用频率
func countFeatures(node *Node, features map[string]float64) {
	if node == nil || node.Left == nil {
		return
	}

	switch node.Feature {
	case 0:
		features["price"] += float64(node.SampleSize)
	case 1:
		features["change"] += float64(node.SampleSize)
	case 2:
		features["ma"] += float64(node.SampleSize)
	}

	countFeatures(node.Left, features)
	countFeatures(node.Right, features)
}
