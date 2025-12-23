package ai

import (
	"context"
	"os"
	"smart-ledger-server/internal/config"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAIResponse_Normal(t *testing.T) {
	jsonStr := `{
                "platform": "微信支付",
                "amount": 25.50,
                "merchant": "星巴克",
                "category": "餐饮",
                "sub_category": "咖啡饮品",
                "pay_time": "2024-01-15T10:30:00+08:00",
                "pay_method": "零钱",
                "order_no": "123456789",
                "items": [{"name": "拿铁", "price": 25.50, "quantity": 1}],
                "confidence": 0.95
        }`
	result, err := ParseAiResponse(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, "微信支付", result.Platform)
	assert.True(t, result.Amount.Equal(decimal.NewFromFloat(25.50)))
	assert.NotNil(t, result.PayTime)
	assert.Equal(t, 0.95, result.Confidence)
}

func TestParseAIResponse_EmptyPayTime(t *testing.T) {

	tests := []struct {
		name      string
		json      string
		wantValue string
	}{
		{
			name: "empty pay time",
			json: `{
                "platform": "微信支付",
                "amount": 25.50,
                "merchant": "星巴克",
                "category": "餐饮",
                "sub_category": "咖啡饮品",
                "pay_time": "",
                "pay_method": "零钱",
                "order_no": "123456789",
                "items": [{"name": "拿铁", "price": 25.50, "quantity": 1}],
                "confidence": 0.95
        }`,
			wantValue: "",
		}, {
			name: "null pay time",
			json: `{
                "platform": "微信支付",
                "amount": 25.50,
                "merchant": "星巴克",
                "category": "餐饮",
                "sub_category": "咖啡饮品",
                "pay_time": null,
                "pay_method": "零钱",
                "order_no": "123456789",
                "items": [{"name": "拿铁", "price": 25.50, "quantity": 1}],
                "confidence": 0.95
        }`,
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAiResponse(tt.json)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, result.PayTime)
		})
	}
}

func TestOpenAIClient_RecognizePayment_Integration(t *testing.T) {
	apikey := os.Getenv("OPENAI_API_KEY")
	baseurl := os.Getenv("OPENAI_BASE_URL")

	if apikey == "" || baseurl == "" {
		t.Skip("OPENAI相关配置未设置，跳过集成测试")
	}
	testConfig := &config.AIConfig{
		APIKey:  apikey,
		BaseURL: baseurl,
		Model:   "qwen3-vl-8b-instruct",
	}

	tests := []struct {
		name         string
		imagePath    string
		wantAmount   decimal.Decimal
		wantPayTime  string
		wantBillType int
	}{
		{
			name:         "normal payment",
			imagePath:    "test_img/test_payment_normal.jpg",
			wantAmount:   decimal.NewFromFloat(17.3),
			wantPayTime:  "2025-12-11T12:24:23+08:00",
			wantBillType: 1,
		},
		{
			name:         "no pay time",
			imagePath:    "test_img/test_payment_miss_paytime.jpg",
			wantAmount:   decimal.NewFromFloat(36.53),
			wantPayTime:  "",
			wantBillType: 1,
		},
		{
			name:         "in come",
			imagePath:    "test_img/test_income.jpg",
			wantAmount:   decimal.NewFromFloat(1),
			wantPayTime:  "2025-12-17T22:26:04+08:00",
			wantBillType: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOpenAIClient(testConfig)
			require.NoError(t, err)
			// 读取测试图片
			testImagePath := tt.imagePath
			imageData, err := os.ReadFile(testImagePath)
			require.NoError(t, err)
			mimeType := "image/jpeg"

			// 调用识别
			result, err := client.RecognizePayment(context.Background(), imageData, mimeType, GetRecognitionPrompt())
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.True(t, result.Amount.Equal(tt.wantAmount))
			assert.Equal(t, tt.wantPayTime, result.PayTime)
			t.Logf("识别结果: %+v", result)
		})
	}

}
