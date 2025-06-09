package ta

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

// KlineData K线数据结构 (Candlestick Data Structure)
// 包含开盘时间、开盘价、最高价、最低价、收盘价和成交量等基本信息
type KlineData struct {
	StartTime int64   `json:"startTime"` // 开盘时间戳
	Open      float64 `json:"open"`      // 开盘价
	High      float64 `json:"high"`      // 最高价
	Low       float64 `json:"low"`       // 最低价
	Close     float64 `json:"close"`     // 收盘价
	Volume    float64 `json:"volume"`    // 成交量
}

// KlineDatas K线数据集合 (Candlestick Data Collection)
type KlineDatas []*KlineData

// fieldCache 字段缓存结构 (Field Cache Structure)
// 用于缓存结构体字段的索引信息，提高反射性能
type fieldCache struct {
	timeFieldIndex   []int // 时间字段索引
	openFieldIndex   []int // 开盘价字段索引
	highFieldIndex   []int // 最高价字段索引
	lowFieldIndex    []int // 最低价字段索引
	closeFieldIndex  []int // 收盘价字段索引
	volumeFieldIndex []int // 成交量字段索引
	isTimeInt64      bool  // 时间字段是否为int64类型
	isStringPrice    bool  // 价格是否为字符串类型
}

// 支持的字段名称常量
var (
	timeFields    = []string{"StartTime", "OpenTime", "Time", "t", "T", "Timestamp", "OpenAt", "EventTime"}
	openFields    = []string{"Open", "OpenPrice", "O", "o"}
	highFields    = []string{"High", "HighPrice", "H", "h"}
	lowFields     = []string{"Low", "LowPrice", "L", "l"}
	closeFields   = []string{"Close", "ClosePrice", "C", "c"}
	volumeFields  = []string{"Volume", "Vol", "V", "v", "Amount", "Quantity"}
	fieldCacheMap = make(map[reflect.Type]*fieldCache)
	cacheMutex    sync.RWMutex
)

// findAndCacheFields 解析并缓存结构体的字段信息 (Parse and Cache Struct Fields)
// 参数：
//   - t: 需要解析的结构体类型
//
// 返回值：
//   - *fieldCache: 包含字段缓存信息的结构体
//   - error: 解析过程中可能发生的错误
//
// 说明：
//
//	使用反射解析结构体字段，并将结果缓存以提高性能
//	支持多种常见的字段命名方式
//
// 示例：
//
//	cache, err := findAndCacheFields(reflect.TypeOf(klineData))
func findAndCacheFields(t reflect.Type) (*fieldCache, error) {
	cacheMutex.RLock()
	if cache, ok := fieldCacheMap[t]; ok {
		cacheMutex.RUnlock()
		return cache, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 双重检查
	if cache, ok := fieldCacheMap[t]; ok {
		return cache, nil
	}

	cache := &fieldCache{}

	// 查找时间字段
	for _, field := range timeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.timeFieldIndex = f.Index
			cache.isTimeInt64 = f.Type.Kind() == reflect.Int64
			break
		}
	}
	if cache.timeFieldIndex == nil {
		return nil, fmt.Errorf("未找到时间字段，支持的字段名：%v", timeFields)
	}

	// 查找价格字段
	for _, field := range openFields {
		if f, ok := t.FieldByName(field); ok {
			cache.openFieldIndex = f.Index
			cache.isStringPrice = f.Type.Kind() == reflect.String
			break
		}
	}
	for _, field := range highFields {
		if f, ok := t.FieldByName(field); ok {
			cache.highFieldIndex = f.Index
			break
		}
	}
	for _, field := range lowFields {
		if f, ok := t.FieldByName(field); ok {
			cache.lowFieldIndex = f.Index
			break
		}
	}
	for _, field := range closeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.closeFieldIndex = f.Index
			break
		}
	}
	for _, field := range volumeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.volumeFieldIndex = f.Index
			break
		}
	}

	fieldCacheMap[t] = cache
	return cache, nil
}

// extractKlineData 从反射值中提取K线数据的各个字段 (Extract Kline Data Fields)
// 参数：
//   - item: 包含K线数据的反射值
//   - cache: 字段缓存信息
//
// 返回值：
//   - startTime: K线的起始时间戳
//   - open: 开盘价
//   - high: 最高价
//   - low: 最低价
//   - close: 收盘价
//   - volume: 成交量
//   - err: 提取过程中可能发生的错误
//
// 说明：
//
//	使用缓存的字段索引快速提取数据
//	支持多种数据类型的自动转换
//
// 示例：
//
//	startTime, open, high, low, close, volume, err := extractKlineData(reflect.ValueOf(kline), cache)
func extractKlineData(item reflect.Value, cache *fieldCache) (startTime int64, open, high, low, close, volume string, err error) {
	if item.Kind() == reflect.Ptr {
		item = item.Elem()
	}

	// 使用字段索引直接访问
	timeField := item.FieldByIndex(cache.timeFieldIndex)
	if cache.isTimeInt64 {
		startTime = timeField.Int()
	} else {
		// 处理其他类型的时间字段
		switch timeField.Kind() {
		case reflect.String:
			if t, err := strconv.ParseInt(timeField.String(), 10, 64); err == nil {
				startTime = t
			}
		case reflect.Float64:
			startTime = int64(timeField.Float())
		}
	}

	getValue := func(fieldIndex []int) string {
		if fieldIndex == nil {
			return ""
		}
		field := item.FieldByIndex(fieldIndex)
		if cache.isStringPrice {
			return field.String()
		}
		// 如果是数值类型，直接转换为字符串
		switch field.Kind() {
		case reflect.Float64:
			return strconv.FormatFloat(field.Float(), 'f', -1, 64)
		case reflect.Int64:
			return strconv.FormatInt(field.Int(), 10)
		}
		return field.String()
	}

	open = getValue(cache.openFieldIndex)
	high = getValue(cache.highFieldIndex)
	low = getValue(cache.lowFieldIndex)
	close = getValue(cache.closeFieldIndex)
	volume = getValue(cache.volumeFieldIndex)

	return
}

// NewKlineDatas 创建新的K线数据集合 (Create New Kline Data Collection)
// 参数：
//   - klines: 原始K线数据接口
//   - l: 是否进行日志记录
//
// 返回值：
//   - KlineDatas: 处理后的K线数据集合
//   - error: 处理过程中可能发生的错误
//
// 说明：
//
//	支持多种格式的K线数据转换
//	可以并行处理大量数据以提高性能
//	自动处理字段类型转换
//
// 示例：
//
//	data, err := NewKlineDatas(rawKlines, true)
func NewKlineDatas(klines interface{}, l bool) (KlineDatas, error) {
	v := reflect.ValueOf(klines)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("输入必须是切片类型")
	}

	length := v.Len()
	if l && length > 0 {
		length--
	}
	if length == 0 {
		return nil, errors.New("没有K线数据")
	}

	// 预分配内存
	klineDataList := make(KlineDatas, length)

	// 获取字段缓存
	firstItem := v.Index(0)
	if firstItem.Kind() == reflect.Ptr {
		firstItem = firstItem.Elem()
	}
	cache, err := findAndCacheFields(firstItem.Type())
	if err != nil {
		return nil, err
	}

	// 使用工作池处理大量数据
	if length > 1000 {
		var wg sync.WaitGroup
		errChan := make(chan error, length)
		workers := runtime.NumCPU()
		batchSize := length / workers
		if batchSize == 0 {
			batchSize = 1
		}

		for i := 0; i < workers; i++ {
			start := i * batchSize
			end := start + batchSize
			if i == workers-1 {
				end = length
			}

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for i := start; i < end; i++ {
					startTime, open, high, low, close, volume, err := extractKlineData(v.Index(i), cache)
					if err != nil {
						errChan <- fmt.Errorf("处理第%d条数据时出错: %v", i+1, err)
						return
					}

					if open == "" || high == "" || low == "" || close == "" || volume == "" {
						errChan <- fmt.Errorf("第%d条数据缺少必要字段", i+1)
						return
					}

					o, err1 := strconv.ParseFloat(open, 64)
					h, err2 := strconv.ParseFloat(high, 64)
					l, err3 := strconv.ParseFloat(low, 64)
					c, err4 := strconv.ParseFloat(close, 64)
					v, err5 := strconv.ParseFloat(volume, 64)

					if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
						errChan <- fmt.Errorf("第%d条数据转换失败", i+1)
						return
					}

					klineDataList[i] = &KlineData{
						StartTime: startTime,
						Open:      o,
						High:      h,
						Low:       l,
						Close:     c,
						Volume:    v,
					}
				}
			}(start, end)
		}

		// 等待所有工作完成
		go func() {
			wg.Wait()
			close(errChan)
		}()

		// 检查错误
		for err := range errChan {
			if err != nil {
				return nil, err
			}
		}
	} else {
		// 对于小数据量，直接处理
		for i := 0; i < length; i++ {
			startTime, open, high, low, close, volume, err := extractKlineData(v.Index(i), cache)
			if err != nil {
				return nil, fmt.Errorf("处理第%d条数据时出错: %v", i+1, err)
			}

			if open == "" || high == "" || low == "" || close == "" || volume == "" {
				return nil, fmt.Errorf("第%d条数据缺少必要字段", i+1)
			}

			o, err1 := strconv.ParseFloat(open, 64)
			h, err2 := strconv.ParseFloat(high, 64)
			l, err3 := strconv.ParseFloat(low, 64)
			c, err4 := strconv.ParseFloat(close, 64)
			v, err5 := strconv.ParseFloat(volume, 64)

			if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
				return nil, fmt.Errorf("第%d条数据转换失败", i+1)
			}

			klineDataList[i] = &KlineData{
				StartTime: startTime,
				Open:      o,
				High:      h,
				Low:       l,
				Close:     c,
				Volume:    v,
			}
		}
	}

	return klineDataList, nil
}

// _ExtractSlice 从K线数据集合中提取指定类型的价格序列 (Extract Price Series)
// 参数：
//   - priceType: 价格类型（"open"/"high"/"low"/"close"/"volume"）
//
// 返回值：
//   - []float64: 提取的价格序列
//   - error: 提取过程中可能发生的错误
//
// 说明：
//
//	支持提取开盘价、最高价、最低价、收盘价和成交量
//	返回的序列长度与K线数据集合长度相同
//
// 示例：
//
//	prices, err := klines._ExtractSlice("close")
func (k *KlineDatas) _ExtractSlice(priceType string) ([]float64, error) {
	var prices []float64
	for _, kline := range *k {
		switch priceType {
		case "open":
			prices = append(prices, kline.Open)
		case "high":
			prices = append(prices, kline.High)
		case "low":
			prices = append(prices, kline.Low)
		case "close":
			prices = append(prices, kline.Close)
		case "volume":
			prices = append(prices, kline.Volume)
		}
	}
	return prices, nil
}

// _Add 添加新的K线数据 (Add New Kline Data)
// 参数：
//   - wsKline: 新的K线数据
//
// 返回值：
//   - error: 添加过程中可能发生的错误
//
// 说明：
//
//	支持添加单个K线数据
//	自动进行数据类型转换和验证
//
// 示例：
//
//	err := klines._Add(newKline)
func (k *KlineDatas) _Add(wsKline interface{}) error {
	v := reflect.ValueOf(wsKline)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("数据必须是结构体类型")
	}

	// 获取字段缓存
	cache, err := findAndCacheFields(v.Type())
	if err != nil {
		return err
	}

	startTime, open, high, low, close, volume, err := extractKlineData(v, cache)
	if err != nil {
		return err
	}

	if open == "" || high == "" || low == "" || close == "" || volume == "" {
		return fmt.Errorf("缺少必要字段")
	}

	o, err1 := strconv.ParseFloat(open, 64)
	h, err2 := strconv.ParseFloat(high, 64)
	l, err3 := strconv.ParseFloat(low, 64)
	c, err4 := strconv.ParseFloat(close, 64)
	v5, err5 := strconv.ParseFloat(volume, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return fmt.Errorf("数据转换失败")
	}

	*k = append(*k, &KlineData{
		StartTime: startTime,
		Open:      o,
		High:      h,
		Low:       l,
		Close:     c,
		Volume:    v5,
	})
	return nil
}

// _Remove 移除指定数量的K线数据 (Remove Kline Data)
// 参数：
//   - n: 需要移除的数据数量
//
// 返回值：
//   - error: 移除过程中可能发生的错误
//
// 说明：
//
//	从集合开头移除指定数量的K线数据
//	如果n大于集合长度，返回错误
//
// 示例：
//
//	err := klines._Remove(10)
func (k *KlineDatas) _Remove(n int) error {
	if n <= 0 {
		return fmt.Errorf("删除数量必须大于0")
	}

	if len(*k) < n {
		return fmt.Errorf("要删除的数量(%d)大于现有数据量(%d)", n, len(*k))
	}

	// 保留后面的数据
	*k = (*k)[n:]
	return nil
}

// _Keep 保留指定数量的最新K线数据 (Keep Latest Kline Data)
// 参数：
//   - n: 需要保留的数据数量
//
// 返回值：
//   - KlineDatas: 保留的K线数据集合
//   - error: 处理过程中可能发生的错误
//
// 说明：
//
//	保留集合末尾的n个K线数据
//	如果n大于集合长度，返回错误
//
// 示例：
//
//	kept, err := klines._Keep(100)
func (k *KlineDatas) _Keep(n int) (KlineDatas, error) {
	if n <= 0 {
		return nil, fmt.Errorf("保留数量必须大于0")
	}

	if len(*k) < n {
		return nil, fmt.Errorf("要保留的数量(%d)大于现有数据量(%d)", n, len(*k))
	}

	// 创建新的KlineDatas并复制最后n根K线数据
	newK := make(KlineDatas, n)
	copy(newK, (*k)[len(*k)-n:])
	return newK, nil
}

// _Keep_ 保留指定数量的最新K线数据（原地修改） (Keep Latest Kline Data In-Place)
// 参数：
//   - n: 需要保留的数据数量
//
// 返回值：
//   - error: 处理过程中可能发生的错误
//
// 说明：
//
//	直接修改原集合，保留末尾的n个K线数据
//	如果n大于集合长度，返回错误
//
// 示例：
//
//	err := klines._Keep_(100)
func (k *KlineDatas) _Keep_(n int) error {
	if n <= 0 {
		return fmt.Errorf("保留数量必须大于0")
	}

	if len(*k) < n {
		return fmt.Errorf("要保留的数量(%d)大于现有数据量(%d)", n, len(*k))
	}

	*k = (*k)[len(*k)-n:]
	return nil
}

// _GetLast 获取最后一个K线数据的指定价格 (Get Last Price)
// 参数：
//   - source: 价格类型（"open"/"high"/"low"/"close"/"volume"）
//
// 返回值：
//   - float64: 最后一个K线数据的指定价格
//
// 说明：
//
//	获取集合中最后一个K线数据的指定类型价格
//	如果集合为空或价格类型无效，返回0
//
// 示例：
//
//	lastClose := klines._GetLast("close")
func (k *KlineDatas) _GetLast(source string) float64 {
	if len(*k) == 0 {
		return -1
	}
	lastKline := (*k)[len(*k)-1]
	switch source {
	case "open":
		return lastKline.Open
	case "high":
		return lastKline.High
	case "low":
		return lastKline.Low
	case "close":
		return lastKline.Close
	case "volume":
		return lastKline.Volume
	default:
		return -1
	}
}

// preallocateSlices 预分配多个切片
// 参数：
//   - length: 每个切片的长度
//   - count: 需要预分配的切片数量
//
// 返回值：
//   - [][]float64: 预分配的切片数组
//
// 说明：
//
//	预分配内存可以提高性能，避免运行时的内存分配
//
// 示例：
//
//	slices := preallocateSlices(100, 3) // 预分配3个长度为100的切片
func preallocateSlices(length int, count int) [][]float64 {
	slices := make([][]float64, count)
	for i := range slices {
		slices[i] = make([]float64, length)
	}
	return slices
}

// max 返回两个数中的较大值 (Maximum Value)
// 参数：
//   - a: 第一个数
//   - b: 第二个数
//
// 返回值：
//   - float64: 较大的数
//
// 示例：
//
//	maxValue := max(1.2, 3.4)
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min 返回两个数中的较小值 (Minimum Value)
// 参数：
//   - a: 第一个数
//   - b: 第二个数
//
// 返回值：
//   - float64: 较小的数
//
// 示例：
//
//	minValue := min(1.2, 3.4)
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
