package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/pkg/ai"
	"smart-ledger-server/pkg/errcode"
)

// AIService AI识别服务
type AIService struct {
	client          ai.Client
	billService     BillServiceInterface
	categoryService CategoryServiceInterface
	maxImageSize    int64
	workerPool      *ai.WorkerPool
	batchConfig     *config.BatchConfig
}

// NewAIService 创建AI服务
func NewAIService(cfg *config.AIConfig, billService BillServiceInterface, categoryService CategoryServiceInterface) (*AIService, error) {
	client, err := ai.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// 创建RPM限流器
	limiter := ai.NewRPMLimiter(cfg.Batch.RPM)

	// 创建Worker Pool
	taskTimeout := time.Duration(cfg.Batch.TaskTimeout) * time.Second
	workerPool := ai.NewWorkerPool(
		cfg.Batch.WorkerCount,
		limiter,
		taskTimeout,
		client,
		cfg.MaxImageSize,
	)

	return &AIService{
		client:          client,
		billService:     billService,
		categoryService: categoryService,
		maxImageSize:    cfg.MaxImageSize,
		workerPool:      workerPool,
		batchConfig:     &cfg.Batch,
	}, nil
}

// RecognizeImage 识别图片
func (s *AIService) RecognizeImage(ctx context.Context, file *multipart.FileHeader) (*dto.AIRecognizeResponse, error) {
	// 检查文件大小
	if file.Size > s.maxImageSize {
		return nil, errcode.ErrImageTooLarge
	}

	// 检查文件类型
	contentType := file.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return nil, errcode.ErrImageFormatInvalid
	}

	// 读取图片数据
	imageData, mimeType, err := ai.ReadImageFromFile(file)
	if err != nil {
		return nil, errcode.ErrServer
	}

	// 获取分类数据，构建提示词
	var prompt string
	categories, err := s.categoryService.GetCategoriesForAI(ctx)
	if err != nil || len(categories) == 0 {
		// 降级方案：使用默认提示词
		prompt = ai.GetRecognitionPrompt()
	} else {
		prompt = ai.BuildRecognitionPrompt(categories)
	}

	// 调用AI识别
	result, err := s.client.RecognizePayment(ctx, imageData, mimeType, prompt)
	if err != nil {
		return nil, errcode.ErrAIRecognizeFailed.WithMessage(err.Error())
	}

	return result, nil
}

// RecognizeAndCreateBill 识别图片并创建账单
func (s *AIService) RecognizeAndCreateBill(ctx context.Context, userID uint64, file *multipart.FileHeader) (*dto.BillResponse, error) {
	// 识别图片
	aiResult, err := s.RecognizeImage(ctx, file)
	if err != nil {
		return nil, err
	}

	// TODO: 保存图片到对象存储，获取图片路径
	imagePath := ""

	// 创建账单
	return s.billService.CreateFromAI(ctx, userID, aiResult, imagePath)
}

// isValidImageType 检查是否为有效的图片类型
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

// BatchRecognizeImages 批量识别图片
func (s *AIService) BatchRecognizeImages(ctx context.Context, files []*multipart.FileHeader) (*dto.BatchRecognizeResponse, error) {
	// 校验图片数量
	if len(files) == 0 {
		return nil, errcode.ErrParams.WithMessage("请至少上传一张图片")
	}
	if len(files) > s.batchConfig.MaxImages {
		return nil, errcode.ErrParams.WithMessage(fmt.Sprintf("单次最多上传 %d 张图片", s.batchConfig.MaxImages))
	}

	// 获取分类数据，构建提示词
	var prompt string
	categories, err := s.categoryService.GetCategoriesForAI(ctx)
	if err != nil || len(categories) == 0 {
		prompt = ai.GetRecognitionPrompt()
	} else {
		prompt = ai.BuildRecognitionPrompt(categories)
	}

	// 构建任务列表
	tasks := make([]ai.Task, len(files))
	for i, file := range files {
		tasks[i] = ai.Task{
			Index:  i,
			File:   file,
			Prompt: prompt,
		}
	}

	// 执行批量识别
	results := s.workerPool.Execute(ctx, tasks)

	// 转换为响应格式
	return s.toBatchResponse(results), nil
}

// BatchRecognizeAndSave 批量识别并保存
func (s *AIService) BatchRecognizeAndSave(ctx context.Context, userID uint64, files []*multipart.FileHeader) (*dto.BatchRecognizeAndSaveResponse, error) {
	// 先批量识别
	batchResult, err := s.BatchRecognizeImages(ctx, files)
	if err != nil {
		return nil, err
	}

	// 逐个保存成功识别的结果
	response := &dto.BatchRecognizeAndSaveResponse{
		Total:        batchResult.Total,
		SuccessCount: 0,
		FailCount:    0,
		Results:      make([]dto.BatchItemSaveResult, len(batchResult.Results)),
	}

	for i, item := range batchResult.Results {
		saveResult := dto.BatchItemSaveResult{
			Index:    item.Index,
			FileName: item.FileName,
		}

		if !item.Success {
			saveResult.Success = false
			saveResult.Error = item.Error
			response.FailCount++
		} else {
			// 保存账单
			// TODO: 保存图片到对象存储
			imagePath := ""
			bill, err := s.billService.CreateFromAI(ctx, userID, item.Data, imagePath)
			if err != nil {
				saveResult.Success = false
				saveResult.Error = "保存账单失败"
				response.FailCount++
			} else {
				saveResult.Success = true
				saveResult.Bill = bill
				response.SuccessCount++
			}
		}

		response.Results[i] = saveResult
	}

	return response, nil
}

// toBatchResponse 转换为批量识别响应
func (s *AIService) toBatchResponse(results []ai.TaskResult) *dto.BatchRecognizeResponse {
	response := &dto.BatchRecognizeResponse{
		Total:        len(results),
		SuccessCount: 0,
		FailCount:    0,
		Results:      make([]dto.BatchItemResult, len(results)),
	}

	for i, result := range results {
		item := dto.BatchItemResult{
			Index:    result.Index,
			FileName: result.FileName,
			Success:  result.Success,
			Duration: result.Duration,
		}

		if result.Success {
			item.Data = result.Data
			response.SuccessCount++
		} else {
			item.Error = result.Error
			response.FailCount++
		}

		response.Results[i] = item
	}

	return response
}
