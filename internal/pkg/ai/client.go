package ai

import (
	"context"
	"encoding/base64"
	"io"
	"mime/multipart"
	"strings"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/model/dto"
)

// Client AI客户端接口
type Client interface {
	// RecognizePayment 识别支付截图
	// prompt: 识别提示词（包含动态分类信息）
	RecognizePayment(ctx context.Context, imageData []byte, mimeType string, prompt string) (*dto.AIRecognizeResponse, error)
}

// NewClient 根据配置创建AI客户端
func NewClient(cfg *config.AIConfig) (Client, error) {
	return NewOpenAIClient(cfg)
}

// ReadImageFromFile 从上传的文件读取图片数据
func ReadImageFromFile(file *multipart.FileHeader) ([]byte, string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, "", err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return nil, "", err
	}

	// 获取MIME类型
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg" // 默认
	}

	return data, mimeType, nil
}

// ImageToBase64 将图片数据转换为base64
func ImageToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// GetRecognitionPrompt 获取识别提示词
func GetRecognitionPrompt() string {
	return `你是一个专业的支付截图识别助手。请分析这张支付截图，提取以下信息并以JSON格式返回：

{
  "platform": "支付平台（微信支付/支付宝/美团/京东/银行APP/其他）",
  "amount": 金额数字（不含货币符号）,
  "merchant": "商家名称",
  "category": "一级分类（餐饮/交通/购物/娱乐/生活服务/金融）",
  "sub_category": "二级分类",
  "pay_time": "支付时间（格式：2006-01-02T15:04:05+08:00）",
  "pay_method": "支付方式（零钱/银行卡/花呗/余额等）",
  "order_no": "订单号（如有）",
  "items": [
    {"name": "商品名", "price": 单价, "quantity": 数量}
  ],
  "confidence": 识别置信度（0-1之间的小数）
}

分类说明：
- 餐饮：正餐、小吃零食、咖啡饮品、水果生鲜、外卖配送费
- 交通：公共交通、打车、共享单车、加油停车
- 购物：日用百货、服饰鞋包、数码电器、美妆护肤
- 娱乐：电影演出、游戏充值、会员订阅、运动健身
- 生活服务：话费充值、水电燃气、医疗健康、快递物流、其他服务
- 金融：转账、还款、理财、保险

注意事项：
1. 金额必须是纯数字，不要包含货币符号
2. 如果无法识别某个字段，请使用空字符串或null
3. 时间格式必须是ISO 8601格式
4. 置信度反映识别结果的可靠程度
5. 只返回JSON，不要有其他文字说明`
}

// BuildRecognitionPrompt 根据分类数据构建识别提示词
func BuildRecognitionPrompt(categories []model.Category) string {
	// 构建一级分类列表
	var topLevelNames []string
	for _, cat := range categories {
		topLevelNames = append(topLevelNames, cat.Name)
	}
	topLevelList := strings.Join(topLevelNames, "/")

	// 构建分类说明
	var categoryDesc strings.Builder
	categoryDesc.WriteString("分类说明：\n")
	for _, cat := range categories {
		var childNames []string
		for _, child := range cat.Children {
			childNames = append(childNames, child.Name)
		}
		if len(childNames) > 0 {
			categoryDesc.WriteString("- ")
			categoryDesc.WriteString(cat.Name)
			categoryDesc.WriteString("：")
			categoryDesc.WriteString(strings.Join(childNames, "、"))
			categoryDesc.WriteString("\n")
		}
	}

	// 构建完整提示词
	var prompt strings.Builder
	prompt.WriteString(`你是一个专业的支付截图识别助手。请分析这张支付截图，提取以下信息并以JSON格式返回：

{
  "platform": "支付平台（微信支付/支付宝/美团/京东/银行APP/其他）",
  "amount": 金额数字（不含货币符号）,
  "merchant": "商家名称",
  "category": "一级分类（`)
	prompt.WriteString(topLevelList)
	prompt.WriteString(`）",
  "sub_category": "二级分类",
  "pay_time": "支付时间（格式：2006-01-02T15:04:05+08:00）",
  "pay_method": "支付方式（零钱/银行卡/花呗/余额等）",
  "order_no": "订单号（如有）",
  "items": [
    {"name": "商品名", "price": 单价, "quantity": 数量}
  ],
  "confidence": 识别置信度（0-1之间的小数）
}

`)
	prompt.WriteString(categoryDesc.String())
	prompt.WriteString(`
注意事项：
1. 金额必须是纯数字，不要包含货币符号
2. 如果无法识别某个字段，请使用空字符串或null
3. 时间格式必须是ISO 8601格式
4. 置信度反映识别结果的可靠程度
5. 只返回JSON，不要有其他文字说明`)

	return prompt.String()
}
