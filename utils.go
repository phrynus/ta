// Package ta 技术分析工具包 (Technical Analysis Package)
// 提供各种技术分析指标的计算功能，包括移动平均线、趋势指标、波动率指标等
package ta

import (
	"runtime"
	"sync"
)

// parallelWorker 并行计算工具 (Parallel Computing Worker)
// 用于处理大规模数据的并行计算任务，提高计算效率
type parallelWorker struct {
	workerCount int            // CPU核心数
	tasks       chan func()    // 任务通道
	wg          sync.WaitGroup // 等待组
}

// newParallelWorker 创建新的并行计算工具
// 返回值：
//   - *parallelWorker: 并行计算工具实例
//
// 示例：
//
//	worker := newParallelWorker()
//	worker.start()
//	defer worker.wait()
func newParallelWorker() *parallelWorker {
	return &parallelWorker{
		workerCount: runtime.NumCPU(),
		tasks:       make(chan func(), runtime.NumCPU()),
	}
}

// start 启动工作池
func (w *parallelWorker) start() {
	for i := 0; i < w.workerCount; i++ {
		go func() {
			for task := range w.tasks {
				task()
				w.wg.Done()
			}
		}()
	}
}

// addTask 添加计算任务
func (w *parallelWorker) addTask(task func()) {
	w.wg.Add(1)
	w.tasks <- task
}

// wait 等待所有任务完成
func (w *parallelWorker) wait() {
	w.wg.Wait()
	close(w.tasks)
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

// calculateChange 计算变化率
// 参数：
//   - prices: 价格序列
//
// 返回值：
//   - []float64: 价格变化率序列
//
// 说明：
//
//	计算相邻价格之间的差值，第一个元素为0
//
// 示例：
//
//	changes := calculateChange([]float64{10, 11, 9, 12})
func calculateChange(prices []float64) []float64 {
	changes := make([]float64, len(prices))
	for i := 1; i < len(prices); i++ {
		changes[i] = prices[i] - prices[i-1]
	}
	return changes
}

// movingSum 计算移动求和
// 参数：
//   - values: 数值序列
//   - period: 计算周期
//
// 返回值：
//   - []float64: 移动求和序列
//
// 说明：
//
//	使用滑动窗口计算指定周期内的数值之和
//	结果序列的前period-1个元素为0
//
// 示例：
//
//	sums := movingSum([]float64{1, 2, 3, 4, 5}, 3)
func movingSum(values []float64, period int) []float64 {
	result := make([]float64, len(values))
	if len(values) < period {
		return result
	}

	// 计算第一个窗口的和
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[period-1] = sum

	// 使用滑动窗口计算后续值
	for i := period; i < len(values); i++ {
		sum = sum - values[i-period] + values[i]
		result[i] = sum
	}

	return result
}

// parallelProcess 并行处理数据 (Parallel Data Processing)
// 参数：
//   - data: 需要处理的数据切片
//   - processFunc: 处理函数，接收数据切片和处理范围
//
// 说明：
//
//	根据CPU核心数将数据分割成多个块并行处理
//	处理函数需要自己处理数据范围内的元素
//
// 示例：
//
//	parallelProcess(data, func(slice []float64, start, end int) {
//	    for i := start; i < end; i++ {
//	        // 处理数据
//	    }
//	})
func parallelProcess(data []float64, processFunc func([]float64, int, int)) {
	worker := newParallelWorker()
	worker.start()

	// 计算每个goroutine处理的数据量
	batchSize := len(data) / worker.workerCount
	if batchSize < 1 {
		batchSize = 1
	}

	// 分配任务
	for i := 0; i < worker.workerCount; i++ {
		start := i * batchSize
		end := start + batchSize
		if i == worker.workerCount-1 {
			end = len(data)
		}
		if start >= len(data) {
			break
		}

		worker.addTask(func() {
			processFunc(data, start, end)
		})
	}

	worker.wait()
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
