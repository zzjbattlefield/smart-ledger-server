package service

import (
	"context"
	"mime/multipart"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/pkg/ai"
	"smart-ledger-server/pkg/errcode"
)

// AIService AI识别服务
type AIService struct {
	client          ai.Client
	billService     *BillService
	categoryService *CategoryService
	maxImageSize    int64
}

// NewAIService 创建AI服务
func NewAIService(cfg *config.AIConfig, billService *BillService, categoryService *CategoryService) (*AIService, error) {
	client, err := ai.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &AIService{
		client:          client,
		billService:     billService,
		categoryService: categoryService,
		maxImageSize:    cfg.MaxImageSize,
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
