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
  "merchant": "商家名称或来源",
  "bill_type": 账单类型（1=支出，2=收入）,
  "category": "一级分类",
  "sub_category": "二级分类",
  "pay_time": "支付时间（格式：2006-01-02T15:04:05+08:00）(如果图片上缺少时间信息，请返回空字符串)",
  "pay_method": "支付方式（零钱/银行卡/花呗/余额等）",
  "order_no": "订单号（如有）",
  "items": [
    {"name": "商品名", "price": 单价, "quantity": 数量}
  ],
  "confidence": 识别置信度（0-1之间的小数）
}

【支出分类】（bill_type=1）：
- 餐饮：正餐、小吃零食、咖啡饮品、水果生鲜、外卖配送费
- 交通：公共交通、打车、共享单车、加油停车
- 购物：日用百货、服饰鞋包、数码电器、美妆护肤
- 娱乐：电影演出、游戏充值、会员订阅、运动健身
- 生活服务：话费充值、水电燃气、医疗健康、快递物流、其他服务
- 金融：转账、还款、理财、保险

【收入分类】（bill_type=2）：
- 薪资：工资、奖金、补贴
- 收红包：微信红包、支付宝红包
- 理财收益：利息、分红、投资收益
- 其他收入：退款、报销、意外来财

注意事项：
1. 金额必须是纯数字，不要包含货币符号
2. 如果无法识别某个字段，请使用空字符串或null
3. 时间格式必须是ISO 8601格式,如果图片上缺少时间信息，请返回空字符串
4. 置信度反映识别结果的可靠程度
5. 只返回JSON，不要有其他文字说明
6. bill_type判断规则：
   - 支出（1）：付款、消费、转账给他人、还款等减少资产的交易
   - 收入（2）：收款、收红包、工资到账、退款、转账收入等增加资产的交易
7. category和sub_category必须从上述对应类型的分类中选择`
}

// BuildRecognitionPrompt 根据分类数据构建识别提示词
func BuildRecognitionPrompt(categories []model.Category) string {
	// 按类型分组分类
	var expenseCategories, incomeCategories []model.Category
	for _, cat := range categories {
		if cat.Type == model.CategoryTypeExpense {
			expenseCategories = append(expenseCategories, cat)
		} else if cat.Type == model.CategoryTypeIncome {
			incomeCategories = append(incomeCategories, cat)
		}
	}

	// 构建支出分类说明
	var expenseDesc strings.Builder
	if len(expenseCategories) > 0 {
		expenseDesc.WriteString("【支出分类】（bill_type=1）：\n")
		for _, cat := range expenseCategories {
			var childNames []string
			for _, child := range cat.Children {
				childNames = append(childNames, child.Name)
			}
			expenseDesc.WriteString("- ")
			expenseDesc.WriteString(cat.Name)
			if len(childNames) > 0 {
				expenseDesc.WriteString("：")
				expenseDesc.WriteString(strings.Join(childNames, "、"))
			}
			expenseDesc.WriteString("\n")
		}
	}

	// 构建收入分类说明
	var incomeDesc strings.Builder
	if len(incomeCategories) > 0 {
		incomeDesc.WriteString("【收入分类】（bill_type=2）：\n")
		for _, cat := range incomeCategories {
			var childNames []string
			for _, child := range cat.Children {
				childNames = append(childNames, child.Name)
			}
			incomeDesc.WriteString("- ")
			incomeDesc.WriteString(cat.Name)
			if len(childNames) > 0 {
				incomeDesc.WriteString("：")
				incomeDesc.WriteString(strings.Join(childNames, "、"))
			}
			incomeDesc.WriteString("\n")
		}
	}

	// 构建完整提示词
	var prompt strings.Builder
	prompt.WriteString(`你是一个专业的支付截图识别助手。请分析这张支付截图，提取以下信息并以JSON格式返回：

{
  "platform": "支付平台（微信支付/支付宝/美团/京东/银行APP/其他）",
  "amount": 金额数字（不含货币符号）,
  "merchant": "商家名称或来源",
  "bill_type": 账单类型（1=支出，2=收入）,
  "category": "一级分类",
  "sub_category": "二级分类",
  "pay_time": "支付时间（格式：2006-01-02T15:04:05+08:00）(如果图片上缺少时间信息，请返回空字符串)",
  "pay_method": "支付方式（零钱/银行卡/花呗/余额等）",
  "order_no": "订单号（如有）",
  "items": [
    {"name": "商品名", "price": 单价, "quantity": 数量}
  ],
  "confidence": 识别置信度（0-1之间的小数）
}

`)
	prompt.WriteString(expenseDesc.String())
	prompt.WriteString("\n")
	prompt.WriteString(incomeDesc.String())
	prompt.WriteString(`
注意事项：
1. 金额必须是纯数字，不要包含货币符号
2. 如果无法识别某个字段，请使用空字符串或null
3. 时间格式必须是ISO 8601格式，如果图片上缺少支付时间信息才返回空字符串
4. 置信度反映识别结果的可靠程度
5. 只返回JSON，不要有其他文字说明
6. bill_type判断规则：
   - 支出（1）：付款、消费、转账给他人、还款等减少资产的交易
   - 收入（2）：收款、收红包、工资到账、退款、转账收入等增加资产的交易
7. category和sub_category必须从上述对应类型的分类中选择`)

	return prompt.String()
}
