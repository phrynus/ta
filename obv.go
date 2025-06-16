package ta

import (
	"fmt"
)

// TaOBV 用于存储 OBV 指标计算结果的结构体
// 说明：
//
//	该结构体用于存储 OBV 指标的计算结果，`Values` 字段保存了每个时间点的 OBV 值。
//
// 字段：
//   - Values: 存储 OBV 指标值的切片 (float64 类型)
type TaOBV struct {
	Values []float64 `json:"values"`
}

// CalculateOBV 计算 OBV 指标值
// 参数：
//   - prices: 价格数据切片 (float64 类型)
//   - volumes: 成交量数据切片 (float64 类型)
//
// 返回值：
//   - *TaOBV: 存储 OBV 指标计算结果的结构体指针
//   - error: 计算过程中可能出现的错误
//
// 说明/注意事项：
//   - 输入的 `prices` 和 `volumes` 切片长度必须一致，否则会返回错误。
//   - 输入数据长度至少为 2，否则会返回错误。
//
// 示例：
//
//	obv, err := CalculateOBV(prices, volumes)
//	if err != nil {
//	    // 处理错误
//	}
func CalculateOBV(prices, volumes []float64) (*TaOBV, error) {
	if len(prices) != len(volumes) {
		return nil, fmt.Errorf("输入数据长度不一致")
	}
	if len(prices) < 2 {
		return nil, fmt.Errorf("计算数据不足")
	}

	obv := make([]float64, len(prices))
	obv[0] = volumes[0]

	for i := 1; i < len(prices); i++ {
		if prices[i] > prices[i-1] {
			obv[i] = obv[i-1] + volumes[i]
		} else if prices[i] < prices[i-1] {
			obv[i] = obv[i-1] - volumes[i]
		} else {
			obv[i] = obv[i-1]
		}
	}

	return &TaOBV{
		Values: obv,
	}, nil
}

// OBV 从 KlineDatas 中提取收盘价和成交量数据并计算 OBV 指标值
// 参数：
//   - source: 数据源标识 (string 类型)
//
// 返回值：
//   - *TaOBV: 存储 OBV 指标计算结果的结构体指针
//   - error: 提取数据或计算过程中可能出现的错误
//
// 说明/注意事项：
//   - 该方法会从 `KlineDatas` 中提取 `close` 和 `volume` 数据进行计算。
//   - 若提取数据过程中出现错误，会返回相应的错误信息。
func (k *KlineDatas) OBV(source string) (*TaOBV, error) {
	close, err := k.ExtractSlice("close")
	if err != nil {
		return nil, err
	}
	volume, err := k.ExtractSlice("volume")
	if err != nil {
		return nil, err
	}
	return CalculateOBV(close, volume)
}

// OBV_ 从 KlineDatas 中提取收盘价和成交量数据并计算 OBV 指标的最后一个值
// 参数：
//   - source: 数据源标识 (string 类型)
//
// 返回值：
//   - float64: OBV 指标的最后一个值
//
// 说明/注意事项：
//   - 该方法会先保留最近 50 条数据，若保留过程中出现错误，则使用原始数据。
//   - 若计算过程中出现错误，会返回 0。
func (k *KlineDatas) OBV_(source string) float64 {
	_k, err := k.Keep(50)
	if err != nil {
		_k = *k
	}
	obv, err := _k.OBV(source)
	if err != nil {
		return 0
	}
	return obv.Value()
}

// Value 获取 TaOBV 结构体中 OBV 指标的最后一个值
// 返回值：
//   - float64: OBV 指标的最后一个值
//
// 说明/注意事项：
//   - 若 `Values` 切片为空，可能会导致数组越界错误。
func (t *TaOBV) Value() float64 {
	return t.Values[len(t.Values)-1]
}

// IsTrendUp 判断 OBV 指标是否处于上升趋势
// 返回值：
//   - bool: 若处于上升趋势返回 true，否则返回 false
//
// 说明/注意事项：
//   - 该方法通过计算最近 5 个 OBV 值的线性回归斜率来判断趋势。
//   - 若 `Values` 切片长度小于 5，会直接返回 false。
func (t *TaOBV) IsTrendUp() bool {
	if len(t.Values) < 5 {
		return false
	}
	lastIndex := len(t.Values) - 1

	var sumX, sumY, sumXY, sumX2 float64
	n := 5
	for i := 0; i < n; i++ {
		x := float64(i)
		y := t.Values[lastIndex-i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	return slope < 0
}

// IsTrendDown 判断 OBV 指标是否处于下降趋势
// 返回值：
//   - bool: 若处于下降趋势返回 true，否则返回 false
//
// 说明/注意事项：
//   - 该方法通过计算最近 5 个 OBV 值的线性回归斜率来判断趋势。
//   - 若 `Values` 切片长度小于 5，会直接返回 false。
func (t *TaOBV) IsTrendDown() bool {
	if len(t.Values) < 5 {
		return false
	}
	lastIndex := len(t.Values) - 1

	var sumX, sumY, sumXY, sumX2 float64
	n := 5
	for i := 0; i < n; i++ {
		x := float64(i)
		y := t.Values[lastIndex-i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	return slope > 0
}

// IsVolumeExpanding 判断成交量是否在扩张
// 返回值：
//   - bool: 若成交量在扩张返回 true，否则返回 false
//
// 说明/注意事项：
//   - 若 `Values` 切片长度小于 2，会返回 false。
func (t *TaOBV) IsVolumeExpanding() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] > t.Values[len(t.Values)-2]
}

// IsVolumeContracting 判断成交量是否在收缩
// 返回值：
//   - bool: 若成交量在收缩返回 true，否则返回 false
//
// 说明/注意事项：
//   - 若 `Values` 切片长度小于 2，会返回 false。
func (t *TaOBV) IsVolumeContracting() bool {
	if len(t.Values) < 2 {
		return false
	}
	return t.Values[len(t.Values)-1] < t.Values[len(t.Values)-2]
}
