package ta

import (
	"fmt"
	"math"
)

// TaMLFactor 机器学习因子系统结构体
// 用于整合多个机器学习模型的预测结果
// 包含：
//   - 预测值序列
//   - 预测概率
//   - 预测置信度
//   - 方向预测
//   - 信号强度
//   - 止损位
//   - 止盈位
//   - 特征矩阵
//   - 模型配置信息
type TaMLFactor struct {
	Values        []float64   `json:"values"`        // 预测值序列
	Probabilities []float64   `json:"probabilities"` // 预测概率
	Confidence    []float64   `json:"confidence"`    // 预测置信度
	Direction     []int       `json:"direction"`     // 方向预测 (1=多, -1=空, 0=震荡)
	Strength      []float64   `json:"strength"`      // 信号强度
	StopLoss      []float64   `json:"stop_loss"`     // 止损位
	TakeProfit    []float64   `json:"take_profit"`   // 止盈位
	Features      [][]float64 `json:"features"`      // 特征矩阵
	ModelType     string      `json:"model_type"`    // 模型类型
	WindowSize    int         `json:"window_size"`   // 时间窗口
}

// FeatureExtractor 特征提取器
// 用于从K线数据中提取各类技术指标特征
// 包含：
//   - 趋势指标：MACD、SuperTrend、ADX、T3
//   - 动量指标：RSI、StochRSI、WilliamsR、CCI、KDJ
//   - 波动指标：ATR、VolatilityRatio、Bollinger
//   - 成交量指标：OBV
//   - 机器学习预测：SVR、DTR、DeepTS
type FeatureExtractor struct {
	// 趋势指标
	macd       *TaMacd       // MACD指标
	superTrend *TaSuperTrend // 超级趋势
	adx        *TaADX        // 平均趋向指标
	t3         *TaT3         // T3移动平均线

	// 动量指标
	rsi      *TaRSI       // 相对强弱指标
	stochRsi *TaStochRSI  // 随机相对强弱
	wr       *TaWilliamsR // 威廉指标
	cci      *TaCCI       // 商品通道指标
	kdj      *TaKDJ       // 随机指标

	// 波动指标
	atr  *TaATR             // 平均真实波幅
	vr   *TaVolatilityRatio // 波动率比率
	boll *TaBoll            // 布林带

	// 成交量指标
	obv *TaOBV // 能量潮

	// 机器学习预测
	svr    *TaSVR    // 支持向量回归
	dtr    *TaDTR    // 决策树回归
	deepts *TaDeepTS // 深度时间序列
}

// CalculateMLFactor 计算机器学习因子
// 参数：
//   - klineData: K线数据
//   - windowSize: 时间窗口大小
//   - modelType: 模型类型 ("ensemble", "deep", "hybrid")
//
// 返回值：
//   - *TaMLFactor: 机器学习因子结构体指针
//   - error: 计算过程中可能发生的错误
//
// 说明：
//
//	该函数执行以下步骤：
//	1. 初始化特征提取器
//	2. 提取技术指标特征
//	3. 根据模型类型进行预测
//	4. 计算预测概率和置信度
//
// 示例：
//
//	mlFactor, err := CalculateMLFactor(klineData, 20, "ensemble")
func CalculateMLFactor(klineData KlineDatas, windowSize int, modelType string) (*TaMLFactor, error) {
	if len(klineData) < windowSize {
		return nil, fmt.Errorf("计算数据不足")
	}

	// 初始化特征提取器
	extractor, err := initFeatureExtractor(klineData)
	if err != nil {
		return nil, err
	}

	// 提取特征
	features, err := extractFeatures(klineData, extractor)
	if err != nil {
		return nil, err
	}

	length := len(klineData)
	// 预分配切片
	values := make([]float64, length)
	probabilities := make([]float64, length)
	confidence := make([]float64, length)

	// 根据不同模型类型进行预测
	switch modelType {
	case "ensemble":
		// 集成学习方法
		err = calculateEnsemblePrediction(features, values, probabilities, confidence, extractor)
	case "deep":
		// 深度学习方法
		err = calculateDeepPrediction(features, values, probabilities, confidence, extractor)
	case "hybrid":
		// 混合模型方法
		err = calculateHybridPrediction(features, values, probabilities, confidence, extractor)
	default:
		return nil, fmt.Errorf("不支持的模型类型: %s", modelType)
	}

	if err != nil {
		return nil, err
	}

	return &TaMLFactor{
		Values:        values,
		Probabilities: probabilities,
		Confidence:    confidence,
		Features:      features,
		ModelType:     modelType,
		WindowSize:    windowSize,
	}, nil
}

// initFeatureExtractor 初始化特征提取器
// 参数：
//   - klineData: K线数据
//
// 返回值：
//   - *FeatureExtractor: 特征提取器指针
//   - error: 初始化过程中可能发生的错误
//
// 说明：
//
//	初始化所有技术指标，包括：
//	1. 趋势指标
//	2. 动量指标
//	3. 波动指标
//	4. 成交量指标
//	5. 机器学习模型
func initFeatureExtractor(klineData KlineDatas) (*FeatureExtractor, error) {
	extractor := &FeatureExtractor{}
	var err error

	// 初始化趋势指标
	extractor.macd, err = klineData.MACD("close", 12, 26, 9)
	if err != nil {
		return nil, err
	}

	extractor.superTrend, err = klineData.SuperTrend(10, 3)
	if err != nil {
		return nil, err
	}

	extractor.adx, err = klineData.ADX(14)
	if err != nil {
		return nil, err
	}

	extractor.t3, err = klineData.T3(5, 0.7, "close")
	if err != nil {
		return nil, err
	}

	// 初始化动量指标
	extractor.rsi, err = klineData.RSI(14, "close")
	if err != nil {
		return nil, err
	}

	extractor.stochRsi, err = klineData.StochRSI(14, 14, 3, 3, "close")
	if err != nil {
		return nil, err
	}

	extractor.wr, err = klineData.WilliamsR(14)
	if err != nil {
		return nil, err
	}

	extractor.cci, err = klineData.CCI(20)
	if err != nil {
		return nil, err
	}

	extractor.kdj, err = klineData.KDJ(9, 3, 3)
	if err != nil {
		return nil, err
	}

	// 初始化波动指标
	extractor.atr, err = klineData.ATR(14)
	if err != nil {
		return nil, err
	}

	extractor.vr, err = CalculateVolatilityRatio(klineData, 1, 10)
	if err != nil {
		return nil, err
	}

	extractor.boll, err = klineData.Boll(20, 2, "close")
	if err != nil {
		return nil, err
	}

	// 初始化成交量指标
	extractor.obv, err = klineData.OBV("close")
	if err != nil {
		return nil, err
	}

	// 初始化机器学习预测
	prices, err := klineData.ExtractSlice("close")
	if err != nil {
		return nil, err
	}

	extractor.svr, err = CalculateSVR(prices, 20, "rbf", 0.1, 1.0, 0.1)
	if err != nil {
		return nil, err
	}

	extractor.dtr, err = CalculateDTR(prices, 20, 5, 3)
	if err != nil {
		return nil, err
	}

	extractor.deepts, err = CalculateDeepTS(prices, "lstm", 20, 32)
	if err != nil {
		return nil, err
	}

	return extractor, nil
}

// extractFeatures 提取特征
// 参数：
//   - klineData: K线数据
//   - extractor: 特征提取器
//
// 返回值：
//   - [][]float64: 特征矩阵
//   - error: 特征提取过程中可能发生的错误
//
// 说明：
//
//	提取的特征包括：
//	1. 基础价格特征
//	2. 技术指标特征
//	3. 机器学习预测特征
func extractFeatures(klineData KlineDatas, extractor *FeatureExtractor) ([][]float64, error) {
	length := len(klineData)
	features := make([][]float64, length)
	for i := range features {
		features[i] = make([]float64, 0)
	}

	for i := 0; i < length; i++ {
		// 1. 基础价格特征
		features[i] = append(features[i],
			klineData[i].Open,
			klineData[i].High,
			klineData[i].Low,
			klineData[i].Close,
			klineData[i].Volume,
		)

		// 2. MACD特征
		if i < len(extractor.macd.Macd) {
			features[i] = append(features[i],
				extractor.macd.Macd[i],
				extractor.macd.Dif[i],
				extractor.macd.Dea[i],
			)
		}

		// 3. SuperTrend特征
		if i < len(extractor.superTrend.Upper) {
			var trendValue float64
			if extractor.superTrend.Trend[i] {
				trendValue = 1.0
			}
			features[i] = append(features[i],
				extractor.superTrend.Upper[i],
				extractor.superTrend.Lower[i],
				trendValue,
			)
		}

		// 4. ADX特征
		if i < len(extractor.adx.ADX) {
			features[i] = append(features[i],
				extractor.adx.ADX[i],
				extractor.adx.PlusDI[i],
				extractor.adx.MinusDI[i],
			)
		}

		// 5. T3特征
		if i < len(extractor.t3.Values) {
			features[i] = append(features[i],
				extractor.t3.Values[i],
				extractor.t3.GetDeviation(klineData[i].Close),
			)
		}

		// 6. RSI特征
		if i < len(extractor.rsi.Values) {
			features[i] = append(features[i],
				extractor.rsi.Values[i],
			)
		}

		// 7. StochRSI特征
		if i < len(extractor.stochRsi.K) {
			features[i] = append(features[i],
				extractor.stochRsi.K[i],
				extractor.stochRsi.D[i],
			)
		}

		// 8. Williams %R特征
		if i < len(extractor.wr.Values) {
			features[i] = append(features[i],
				extractor.wr.Values[i],
			)
		}

		// 9. CCI特征
		if i < len(extractor.cci.Values) {
			features[i] = append(features[i],
				extractor.cci.Values[i],
			)
		}

		// 10. KDJ特征
		if i < len(extractor.kdj.K) {
			features[i] = append(features[i],
				extractor.kdj.K[i],
				extractor.kdj.D[i],
				extractor.kdj.J[i],
			)
		}

		// 11. ATR特征
		if i < len(extractor.atr.Values) {
			features[i] = append(features[i],
				extractor.atr.Values[i],
				extractor.atr.GetVolatilityRatio(),
			)
		}

		// 12. 波动率比率特征
		if i < len(extractor.vr.Values) {
			features[i] = append(features[i],
				extractor.vr.Values[i],
			)
		}

		// 13. 布林带特征
		if i < len(extractor.boll.Upper) {
			features[i] = append(features[i],
				extractor.boll.Upper[i],
				extractor.boll.Mid[i],
				extractor.boll.Lower[i],
				(klineData[i].Close-extractor.boll.Mid[i])/extractor.boll.Mid[i],
			)
		}

		// 14. OBV特征
		if i < len(extractor.obv.Values) {
			features[i] = append(features[i],
				extractor.obv.Values[i],
			)
		}

		// 15. SVR预测特征
		if i < len(extractor.svr.Values) {
			features[i] = append(features[i],
				extractor.svr.Values[i],
				extractor.svr.UpperBand[i],
				extractor.svr.LowerBand[i],
			)
		}

		// 16. DTR预测特征
		if i < len(extractor.dtr.Values) {
			features[i] = append(features[i],
				extractor.dtr.Values[i],
			)
		}

		// 17. DeepTS预测特征
		if i < len(extractor.deepts.Values) {
			features[i] = append(features[i],
				extractor.deepts.Values[i],
				0.8, // 使用固定置信度
			)
		}
	}

	return features, nil
}

// calculateEnsemblePrediction 计算集成学习预测
// 参数：
//   - features: 特征矩阵
//   - values: 预测值数组
//   - probabilities: 预测概率数组
//   - confidence: 预测置信度数组
//   - extractor: 特征提取器
//
// 返回值：
//   - error: 预测过程中可能发生的错误
//
// 说明：
//
//	集成预测权重分配：
//	1. SVR: 40%
//	2. DTR: 30%
//	3. DeepTS: 30%
func calculateEnsemblePrediction(features [][]float64, values, probabilities, confidence []float64, extractor *FeatureExtractor) error {
	length := len(features)
	if length == 0 {
		return fmt.Errorf("特征数据为空")
	}

	// 初始化预测结果数组
	values = make([]float64, length)
	probabilities = make([]float64, length)
	confidence = make([]float64, length)

	for i := 0; i < length; i++ {
		// 1. SVR预测权重(40%)
		svrWeight := 0.4
		svrPred := extractor.svr.Values[i]
		svrProb := (extractor.svr.UpperBand[i] - extractor.svr.LowerBand[i]) / extractor.svr.Values[i]

		// 2. DTR预测权重(30%)
		dtrWeight := 0.3
		dtrPred := extractor.dtr.Values[i]
		dtrProb := 0.8 // 使用固定置信度

		// 3. DeepTS预测权重(30%)
		deeptsWeight := 0.3
		deeptsPred := extractor.deepts.Values[i]
		deeptsProb := 0.8 // 使用固定置信度

		// 计算加权预测值
		values[i] = svrPred*svrWeight + dtrPred*dtrWeight + deeptsPred*deeptsWeight

		// 计算预测概率
		probabilities[i] = svrProb*svrWeight + dtrProb*dtrWeight + deeptsProb*deeptsWeight

		// 计算预测置信度
		// 基于技术指标的一致性评分
		technicalConf := calculateTechnicalConfidence(i, extractor)

		// 基于波动性评分
		volatilityConf := calculateVolatilityConfidence(i, extractor)

		// 基于趋势强度评分
		trendConf := calculateTrendConfidence(i, extractor)

		// 综合置信度
		confidence[i] = (technicalConf + volatilityConf + trendConf) / 3.0
	}

	return nil
}

// calculateTechnicalConfidence 计算技术指标一致性置信度
func calculateTechnicalConfidence(index int, extractor *FeatureExtractor) float64 {
	var technicalConf float64

	// 1. MACD趋势一致性
	if index < len(extractor.macd.Macd) {
		if extractor.macd.Macd[index] > extractor.macd.Dea[index] {
			technicalConf += 0.1
		}
	}

	// 2. RSI超买超卖
	if index < len(extractor.rsi.Values) {
		if extractor.rsi.Values[index] > 70 || extractor.rsi.Values[index] < 30 {
			technicalConf += 0.1
		}
	}

	// 3. KDJ金叉死叉
	if index < len(extractor.kdj.K) {
		if extractor.kdj.K[index] > extractor.kdj.D[index] {
			technicalConf += 0.1
		}
	}

	// 4. 布林带位置
	if index < len(extractor.boll.Upper) {
		if extractor.boll.IsBreakoutPossible() {
			technicalConf += 0.1
		}
	}

	// 5. ADX趋势强度
	if index < len(extractor.adx.ADX) {
		if extractor.adx.ADX[index] > 25 {
			technicalConf += 0.1
		}
	}

	return technicalConf
}

// calculateVolatilityConfidence 计算波动性置信度
func calculateVolatilityConfidence(index int, extractor *FeatureExtractor) float64 {
	var volatilityConf float64

	// 1. ATR波动率
	if index < len(extractor.atr.Values) {
		volatilityConf += math.Min(extractor.atr.GetVolatilityRatio(), 0.2)
	}

	// 2. 布林带宽度
	if index < len(extractor.boll.Upper) {
		bandwidth := (extractor.boll.Upper[index] - extractor.boll.Lower[index]) / extractor.boll.Mid[index]
		volatilityConf += math.Min(bandwidth, 0.2)
	}

	// 3. 波动率比率
	if index < len(extractor.vr.Values) {
		volatilityConf += math.Min(extractor.vr.Values[index], 0.2)
	}

	return volatilityConf
}

// calculateTrendConfidence 计算趋势强度置信度
func calculateTrendConfidence(index int, extractor *FeatureExtractor) float64 {
	var trendConf float64

	// 1. SuperTrend方向
	if index < len(extractor.superTrend.Trend) {
		if extractor.superTrend.Trend[index] {
			trendConf += 0.2
		}
	}

	// 2. ADX趋势强度
	if index < len(extractor.adx.ADX) {
		trendConf += math.Min(extractor.adx.ADX[index]/100.0, 0.4)
	}

	// 3. OBV趋势确认
	if index < len(extractor.obv.Values) {
		if extractor.obv.IsTrendUp() {
			trendConf += 0.2
		}
	}

	// 4. T3趋势偏离度
	if index < len(extractor.t3.Values) {
		deviation := math.Abs(extractor.t3.GetDeviation(extractor.t3.Values[index]))
		trendConf += math.Min(deviation/100.0, 0.2)
	}

	// 5. MACD趋势强度
	if index < len(extractor.macd.Macd) {
		trendConf += math.Min(math.Abs(extractor.macd.Macd[index])/100.0, 0.2)
	}

	return trendConf
}

// calculateDeepPrediction 计算深度学习预测
// 参数：
//   - features: 特征矩阵
//   - values: 预测值数组
//   - probabilities: 预测概率数组
//   - confidence: 预测置信度数组
//   - extractor: 特征提取器
//
// 返回值：
//   - error: 预测过程中可能发生的错误
//
// 说明：
//
//	深度学习预测基于：
//	1. DeepTS模型预测
//	2. 特征组合置信度
//	3. 趋势动量确认
func calculateDeepPrediction(features [][]float64, values, probabilities, confidence []float64, extractor *FeatureExtractor) error {
	length := len(features)
	for i := 0; i < length; i++ {
		if i < len(extractor.deepts.Values) {
			// 使用DeepTS的预测值
			values[i] = extractor.deepts.Values[i]

			// 计算预测概率
			probabilities[i] = calculateProbability(features[i])

			// 使用特征组合计算置信度
			trendConf := getTrendConfidence(features[i], extractor, i)
			momentumConf := getMomentumConfidence(features[i], extractor, i)
			volatilityConf := getVolatilityConfidence(features[i], extractor, i)

			confidence[i] = (trendConf + momentumConf + volatilityConf) / 3.0
		}
	}
	return nil
}

// calculateHybridPrediction 计算混合模型预测
// 参数：
//   - features: 特征矩阵
//   - values: 预测值数组
//   - probabilities: 预测概率数组
//   - confidence: 预测置信度数组
//   - extractor: 特征提取器
//
// 返回值：
//   - error: 预测过程中可能发生的错误
//
// 说明：
//
//	混合预测方法：
//	1. 动态权重分配
//	2. 趋势强度调整
//	3. 置信度融合
func calculateHybridPrediction(features [][]float64, values, probabilities, confidence []float64, extractor *FeatureExtractor) error {
	length := len(features)

	// 预分配数组
	direction := make([]int, length)
	strength := make([]float64, length)
	stopLoss := make([]float64, length)
	takeProfit := make([]float64, length)

	for i := 0; i < length; i++ {
		if i < len(extractor.svr.Values) && i < len(extractor.deepts.Values) {
			// 计算趋势强度
			trendStrength := getTrendStrength(features[i], extractor, i)

			// 根据趋势强度动态调整权重
			svrWeight := 0.4 + 0.2*trendStrength
			deepsWeight := 1.0 - svrWeight

			// 混合预测值
			values[i] = svrWeight*extractor.svr.Values[i] +
				deepsWeight*extractor.deepts.Values[i]

			// 计算预测概率
			probabilities[i] = calculateProbability(features[i])

			// 混合置信度
			confidence[i] = svrWeight*extractor.svr.GetConfidenceInterval() +
				deepsWeight*extractor.deepts.GetConfidence()

			// 计算方向和强度
			direction[i], strength[i] = calculateDirectionAndStrength(features[i], extractor, i)

			// 计算止损止盈位
			stopLoss[i], takeProfit[i] = calculateStopLossTakeProfit(values[i], direction[i], extractor, i)
		}
	}

	return nil
}

// calculateDirectionAndStrength 计算交易方向和强度
func calculateDirectionAndStrength(features []float64, extractor *FeatureExtractor, index int) (int, float64) {
	// 趋势信号
	trendSignal := getTrendSignal(extractor, index)

	// 动量信号
	momentumSignal := getMomentumSignal(extractor, index)

	// 波动信号
	volatilitySignal := getVolatilitySignal(extractor, index)

	// 综合信号
	compositeSignal := (trendSignal + momentumSignal + volatilitySignal) / 3.0

	// 确定方向
	var direction int
	if compositeSignal > 0.2 {
		direction = 1 // 多头信号
	} else if compositeSignal < -0.2 {
		direction = -1 // 空头信号
	} else {
		direction = 0 // 震荡信号
	}

	// 计算强度
	strength := math.Abs(compositeSignal)

	return direction, strength
}

// calculateStopLossTakeProfit 计算止损止盈位
func calculateStopLossTakeProfit(currentPrice float64, direction int, extractor *FeatureExtractor, index int) (float64, float64) {
	// 获取ATR值作为波动参考
	atr := extractor.atr.Values[index]

	var stopLoss, takeProfit float64

	switch direction {
	case 1: // 多头
		stopLoss = currentPrice - 2*atr
		takeProfit = currentPrice + 3*atr
	case -1: // 空头
		stopLoss = currentPrice + 2*atr
		takeProfit = currentPrice - 3*atr
	default: // 震荡
		stopLoss = currentPrice - 1.5*atr
		takeProfit = currentPrice + 1.5*atr
	}

	return stopLoss, takeProfit
}

// getTrendSignal 获取趋势信号
func getTrendSignal(extractor *FeatureExtractor, index int) float64 {
	var signal float64

	// MACD信号
	if index < len(extractor.macd.Macd) {
		if extractor.macd.Macd[index] > extractor.macd.Dea[index] {
			signal += 0.3
		} else {
			signal -= 0.3
		}
	}

	// SuperTrend信号
	if index < len(extractor.superTrend.Trend) {
		if extractor.superTrend.Trend[index] {
			signal += 0.4
		} else {
			signal -= 0.4
		}
	}

	// ADX趋势强度
	if index < len(extractor.adx.ADX) {
		if extractor.adx.ADX[index] > 25 {
			if extractor.adx.PlusDI[index] > extractor.adx.MinusDI[index] {
				signal += 0.3
			} else {
				signal -= 0.3
			}
		}
	}

	return signal
}

// getMomentumSignal 获取动量信号
func getMomentumSignal(extractor *FeatureExtractor, index int) float64 {
	var signal float64

	// RSI信号
	if index < len(extractor.rsi.Values) {
		rsi := extractor.rsi.Values[index]
		if rsi > 70 {
			signal -= 0.3
		} else if rsi < 30 {
			signal += 0.3
		}
	}

	// KDJ信号
	if index < len(extractor.kdj.K) {
		if extractor.kdj.K[index] > extractor.kdj.D[index] {
			signal += 0.2
		} else {
			signal -= 0.2
		}
	}

	// CCI信号
	if index < len(extractor.cci.Values) {
		if extractor.cci.Values[index] > 100 {
			signal -= 0.2
		} else if extractor.cci.Values[index] < -100 {
			signal += 0.2
		}
	}

	return signal
}

// getVolatilitySignal 获取波动信号
func getVolatilitySignal(extractor *FeatureExtractor, index int) float64 {
	var signal float64

	// 布林带信号
	if index < len(extractor.boll.Upper) {
		price := extractor.boll.Mid[index]
		if price > extractor.boll.Upper[index] {
			signal -= 0.3
		} else if price < extractor.boll.Lower[index] {
			signal += 0.3
		}
	}

	// ATR波动信号
	if index < len(extractor.atr.Values) {
		atrValue := extractor.atr.Values[index]
		atrRatio := atrValue / extractor.atr.Values[index-1]
		if atrRatio > 1.5 {
			signal *= 0.8 // 高波动降低信号强度
		}
	}

	return signal
}

// getTrendConfidence 获取趋势置信度
func getTrendConfidence(features []float64, extractor *FeatureExtractor, index int) float64 {
	var trendConf float64

	// MACD趋势
	if index < len(extractor.macd.Macd) {
		if extractor.macd.Macd[index] > extractor.macd.Dea[index] {
			trendConf += 0.2
		}
	}

	// ADX趋势强度
	if index < len(extractor.adx.ADX) {
		trendConf += math.Min(extractor.adx.ADX[index]/100.0, 0.4)
	}

	// SuperTrend趋势
	if index < len(extractor.superTrend.Trend) {
		if extractor.superTrend.Trend[index] {
			trendConf += 0.2
		}
	}

	// Ichimoku云层
	if index < len(extractor.macd.Macd) {
		if extractor.macd.Macd[index] > extractor.macd.Dea[index] {
			trendConf += 0.2
		}
	}

	return trendConf
}

// getMomentumConfidence 获取动量置信度
func getMomentumConfidence(features []float64, extractor *FeatureExtractor, index int) float64 {
	var momentumConf float64

	// RSI动量
	if index < len(extractor.rsi.Values) {
		rsi := extractor.rsi.Values[index]
		if rsi > 70 || rsi < 30 {
			momentumConf += 0.25
		}
	}

	// CCI动量
	if index < len(extractor.cci.Values) {
		cci := math.Abs(extractor.cci.Values[index])
		momentumConf += math.Min(cci/200.0, 0.25)
	}

	// Williams %R
	if index < len(extractor.wr.Values) {
		wr := math.Abs(extractor.wr.Values[index])
		if wr > 80 {
			momentumConf += 0.25
		}
	}

	// KDJ
	if index < len(extractor.kdj.K) {
		if math.Abs(extractor.kdj.K[index]-extractor.kdj.D[index]) > 20 {
			momentumConf += 0.25
		}
	}

	return momentumConf
}

// getVolatilityConfidence 获取波动率置信度
func getVolatilityConfidence(features []float64, extractor *FeatureExtractor, index int) float64 {
	var volatilityConf float64

	// ATR波动
	if index < len(extractor.atr.Values) {
		volatilityConf += math.Min(extractor.atr.Values[index]/100.0, 0.25)
	}

	// 波动率
	if index < len(extractor.vr.Values) {
		vr := extractor.vr.Values[index]
		volatilityConf += math.Min(vr, 0.25)
	}

	// 布林带宽度
	if index < len(extractor.boll.Upper) {
		bandwidth := (extractor.boll.Upper[index] - extractor.boll.Lower[index]) / extractor.boll.Mid[index]
		volatilityConf += math.Min(bandwidth, 0.25)
	}

	// CMF资金流
	if index < len(extractor.obv.Values) {
		cmf := math.Abs(extractor.obv.Values[index])
		volatilityConf += math.Min(cmf, 0.25)
	}

	return volatilityConf
}

// getTrendStrength 获取趋势强度
func getTrendStrength(features []float64, extractor *FeatureExtractor, index int) float64 {
	var strength float64

	// ADX趋势强度
	if index < len(extractor.adx.ADX) {
		strength += extractor.adx.ADX[index] / 100.0
	}

	// MACD趋势
	if index < len(extractor.macd.Macd) {
		strength += math.Abs(extractor.macd.Macd[index]) / 100.0
	}

	// SuperTrend趋势
	if index < len(extractor.superTrend.Trend) {
		if extractor.superTrend.Trend[index] {
			strength += 0.5
		}
	}

	return math.Min(strength/2.0, 1.0)
}

// calculateProbability 计算预测概率
func calculateProbability(features []float64) float64 {
	// 使用Softmax函数计算概率
	sum := 0.0
	maxFeature := -math.MaxFloat64
	for _, f := range features {
		if f > maxFeature {
			maxFeature = f
		}
	}

	for _, f := range features {
		sum += math.Exp(f - maxFeature)
	}

	return math.Exp(features[0]-maxFeature) / sum
}

// MLFactor 计算K线数据的机器学习因子
// 参数：
//   - windowSize: 时间窗口大小
//   - modelType: 模型类型 ("ensemble", "deep", "hybrid")
//
// 返回值：
//   - *TaMLFactor: 机器学习因子结构体指针
//   - error: 计算过程中可能发生的错误
//
// 示例：
//
//	mlFactor, err := kline.MLFactor(20, "ensemble")
func (k *KlineDatas) MLFactor(windowSize int, modelType string) (*TaMLFactor, error) {
	return CalculateMLFactor(*k, windowSize, modelType)
}

// Value 返回最新的预测值、概率和置信度
// 返回值：
//   - prediction: 预测值
//   - probability: 预测概率
//   - confidence: 预测置信度
//
// 示例：
//
//	pred, prob, conf := mlFactor.Value()
func (t *TaMLFactor) Value() (prediction, probability, confidence float64) {
	lastIndex := len(t.Values) - 1
	if lastIndex < 0 {
		return 0, 0, 0
	}
	return t.Values[lastIndex], t.Probabilities[lastIndex], t.Confidence[lastIndex]
}

// GetTrend 获取预测趋势
// 返回值：
//   - int: 趋势值（1=上涨，-1=下跌，0=盘整）
//   - float64: 趋势强度（0-1之间）
//
// 说明：
//
//	趋势判断基于：
//	1. 预测值变化
//	2. 置信度调整
//
// 示例：
//
//	trend, strength := mlFactor.GetTrend()
func (t *TaMLFactor) GetTrend() (int, float64) {
	lastIndex := len(t.Values) - 1
	if lastIndex < 1 {
		return 0, 0
	}

	diff := t.Values[lastIndex] - t.Values[lastIndex-1]
	strength := math.Abs(diff) * t.Confidence[lastIndex]

	if diff > 0 {
		return 1, strength
	} else if diff < 0 {
		return -1, strength
	}
	return 0, strength
}

// GetFeatureImportance 获取特征重要性
// 返回值：
//   - map[string]float64: 特征重要性得分
//
// 说明：
//
//	特征重要性计算：
//	1. 计算特征贡献度
//	2. 归一化处理
//
// 示例：
//
//	importance := mlFactor.GetFeatureImportance()
func (t *TaMLFactor) GetFeatureImportance() map[string]float64 {
	importance := make(map[string]float64)
	featureNames := []string{
		"price", "volume", "macd", "rsi", "stoch_rsi",
		"kdj", "atr", "adx", "supertrend", "svr", "dtr", "deepts",
	}

	// 计算每个特征的重要性
	for i, name := range featureNames {
		if i < len(t.Features[0]) {
			var sum float64
			for j := range t.Features {
				sum += math.Abs(t.Features[j][i])
			}
			importance[name] = sum / float64(len(t.Features))
		}
	}

	// 归一化特征重要性
	total := 0.0
	for _, v := range importance {
		total += v
	}
	if total > 0 {
		for k := range importance {
			importance[k] /= total
		}
	}

	return importance
}

// GetTradeSignal 获取交易信号
// 返回值：
//   - direction: 交易方向 (1=多, -1=空, 0=震荡)
//   - strength: 信号强度 (0-1)
//   - stopLoss: 建议止损价格
//   - takeProfit: 建议止盈价格
//   - confidence: 信号置信度
//
// 说明：
//
//	该方法综合考虑多个因素：
//	1. 趋势方向和强度
//	2. 波动率和动量
//	3. 支撑阻力位
//	4. 风险收益比
func (t *TaMLFactor) GetTradeSignal() (direction int, strength, stopLoss, takeProfit, confidence float64) {
	lastIndex := len(t.Values) - 1
	if lastIndex < 1 {
		return 0, 0, 0, 0, 0
	}

	// 获取最新数据
	currentValue := t.Values[lastIndex]
	currentConf := t.Confidence[lastIndex]

	// 计算方向和强度
	direction = t.Direction[lastIndex]
	strength = t.Strength[lastIndex]

	// 获取止损止盈位
	stopLoss = t.StopLoss[lastIndex]
	takeProfit = t.TakeProfit[lastIndex]

	// 计算风险收益比
	var riskRewardRatio float64
	if direction == 1 { // 多头
		riskRewardRatio = (takeProfit - currentValue) / (currentValue - stopLoss)
	} else if direction == -1 { // 空头
		riskRewardRatio = (currentValue - takeProfit) / (stopLoss - currentValue)
	}

	// 根据风险收益比调整置信度
	confidence = currentConf
	if riskRewardRatio < 2.0 {
		confidence *= 0.8 // 降低低风险收益比的信号置信度
	} else if riskRewardRatio > 3.0 {
		confidence *= 1.2 // 提高高风险收益比的信号置信度
	}

	// 确保置信度在0-1之间
	confidence = math.Max(0, math.Min(1, confidence))

	return direction, strength, stopLoss, takeProfit, confidence
}
