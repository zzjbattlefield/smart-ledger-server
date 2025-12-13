package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	pkgerrors "github.com/pkg/errors"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/model/dto"
)

// OpenAIClient OpenAI 兼容客户端
type OpenAIClient struct {
	client openai.Client
	model  shared.ChatModel
}

// NewOpenAIClient 创建 OpenAI 兼容客户端
func NewOpenAIClient(cfg *config.AIConfig) (*OpenAIClient, error) {
	// 准备选项
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}

	// 如果配置了 BaseURL，使用自定义端点
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	// 创建客户端
	client := openai.NewClient(opts...)

	// 模型名称
	model := shared.ChatModel(cfg.Model)
	if cfg.Model == "" {
		model = openai.ChatModelGPT4o
	}

	return &OpenAIClient{
		client: client,
		model:  model,
	}, nil
}

// RecognizePayment 识别支付截图
func (c *OpenAIClient) RecognizePayment(ctx context.Context, imageData []byte, mimeType string, prompt string) (*dto.AIRecognizeResponse, error) {
	// 构建图片 data URL
	base64Image := ImageToBase64(imageData)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	// 构建消息内容（使用 helper 函数）
	contentParts := []openai.ChatCompletionContentPartUnionParam{
		openai.TextContentPart(prompt),
		openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL:    imageURL,
			Detail: "high",
		}),
	}

	// 调用 API
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:    []openai.ChatCompletionMessageParamUnion{openai.UserMessage(contentParts)},
		Model:       c.model,
		MaxTokens:   openai.Int(1000),
		Temperature: openai.Float(0.1),
	})

	if err != nil {
		return nil, fmt.Errorf("OpenAI API 调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI 返回空结果")
	}

	content := resp.Choices[0].Message.Content

	return ParseAiResponse(content)
}

func ParseAiResponse(content string) (result *dto.AIRecognizeResponse, err error) {
	result = &dto.AIRecognizeResponse{}
	err = json.Unmarshal([]byte(content), result)
	if err != nil {
		err = pkgerrors.Wrap(err, "解析 AI 返回结果失败")
	}
	return
}
