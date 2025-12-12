package ai

import (
	"context"
	"mime/multipart"
	"sync"
	"time"

	"smart-ledger-server/internal/model/dto"
)

// Task 表示单个图片识别任务
type Task struct {
	Index  int                   // 图片索引（用于结果排序）
	File   *multipart.FileHeader // 图片文件
	Prompt string                // 识别提示词
}

// TaskResult 任务执行结果
type TaskResult struct {
	Index    int                      // 图片索引
	FileName string                   // 文件名
	Success  bool                     // 是否成功
	Data     *dto.AIRecognizeResponse // 识别结果（成功时）
	Error    string                   // 错误信息（失败时）
	Duration int64                    // 处理耗时(毫秒)
}

// WorkerPool 任务池
type WorkerPool struct {
	workerCount  int           // Worker数量
	limiter      *RPMLimiter   // RPM限流器
	taskTimeout  time.Duration // 单任务超时
	client       Client        // AI客户端
	maxImageSize int64         // 最大图片大小
}

// NewWorkerPool 创建任务池
func NewWorkerPool(workerCount int, limiter *RPMLimiter, taskTimeout time.Duration, client Client, maxImageSize int64) *WorkerPool {
	return &WorkerPool{
		workerCount:  workerCount,
		limiter:      limiter,
		taskTimeout:  taskTimeout,
		client:       client,
		maxImageSize: maxImageSize,
	}
}

// Execute 执行批量任务
// 返回与输入任务顺序一致的结果数组
func (p *WorkerPool) Execute(ctx context.Context, tasks []Task) []TaskResult {
	results := make([]TaskResult, len(tasks))

	// 任务通道
	taskChan := make(chan Task, len(tasks))

	// 结果通道
	resultChan := make(chan TaskResult, len(tasks))

	// 启动 Workers
	var wg sync.WaitGroup
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go p.worker(ctx, &wg, taskChan, resultChan)
	}

	// 发送任务
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	// 等待所有 Worker 完成后关闭结果通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for result := range resultChan {
		results[result.Index] = result
	}

	return results
}

// worker 工作协程
func (p *WorkerPool) worker(ctx context.Context, wg *sync.WaitGroup, taskChan <-chan Task, resultChan chan<- TaskResult) {
	defer wg.Done()

	for task := range taskChan {
		result := p.processTask(ctx, task)
		resultChan <- result
	}
}

// processTask 处理单个任务
func (p *WorkerPool) processTask(ctx context.Context, task Task) TaskResult {
	startTime := time.Now()
	result := TaskResult{
		Index:    task.Index,
		FileName: task.File.Filename,
	}

	// 创建带超时的context
	taskCtx, cancel := context.WithTimeout(ctx, p.taskTimeout)
	defer cancel()

	// 等待RPM限流
	if err := p.limiter.Wait(taskCtx); err != nil {
		result.Error = "请求过于频繁，请稍后重试"
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// 校验文件大小
	if task.File.Size > p.maxImageSize {
		result.Error = "图片过大"
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// 校验文件类型
	contentType := task.File.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		result.Error = "图片格式无效"
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// 读取图片数据
	imageData, mimeType, err := ReadImageFromFile(task.File)
	if err != nil {
		result.Error = "读取图片失败"
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// 调用AI识别
	aiResult, err := p.client.RecognizePayment(taskCtx, imageData, mimeType, task.Prompt)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	result.Success = true
	result.Data = aiResult
	result.Duration = time.Since(startTime).Milliseconds()
	return result
}

// isValidImageType 检查图片类型
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}
