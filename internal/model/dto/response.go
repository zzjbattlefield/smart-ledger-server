package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// =============== 用户相关 ===============

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID          uint64     `json:"id"`
	Phone       string     `json:"phone"`
	Nickname    string     `json:"nickname"`
	AvatarURL   string     `json:"avatar_url"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// =============== 账单相关 ===============

// BillResponse 账单响应
type BillResponse struct {
	ID          uint64             `json:"id"`
	UUID        string             `json:"uuid"`
	Amount      decimal.Decimal    `json:"amount"`
	BillType    int                `json:"bill_type"`
	Platform    string             `json:"platform"`
	Merchant    string             `json:"merchant"`
	Category    *CategoryResponse  `json:"category"`
	PayTime     time.Time          `json:"pay_time"`
	PayMethod   string             `json:"pay_method"`
	OrderNo     string             `json:"order_no"`
	Remark      string             `json:"remark"`
	Confidence  float64            `json:"confidence"`
	IsConfirmed bool               `json:"is_confirmed"`
	Items       []BillItemResponse `json:"items,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

// BillItemResponse 账单明细响应
type BillItemResponse struct {
	ID       uint64          `json:"id"`
	Name     string          `json:"name"`
	Price    decimal.Decimal `json:"price"`
	Quantity int             `json:"quantity"`
}

// BillListResponse 账单列表响应
type BillListResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	List     []BillResponse `json:"list"`
}

// =============== AI 识别相关 ===============

// AIRecognizeResponse AI识别响应
type AIRecognizeResponse struct {
	Platform    string                    `json:"platform"`
	Amount      decimal.Decimal           `json:"amount"`
	Merchant    string                    `json:"merchant"`
	Category    string                    `json:"category"`
	SubCategory string                    `json:"sub_category"`
	PayTime     string                    `json:"pay_time"`
	PayMethod   string                    `json:"pay_method"`
	OrderNo     string                    `json:"order_no"`
	Items       []AIRecognizeItemResponse `json:"items,omitempty"`
	Confidence  float64                   `json:"confidence"`
}

// AIRecognizeItemResponse AI识别明细响应
type AIRecognizeItemResponse struct {
	Name     string          `json:"name"`
	Price    decimal.Decimal `json:"price"`
	Quantity int             `json:"quantity"`
}

// =============== 统计相关 ===============

// StatsSummaryResponse 统计摘要响应
type StatsSummaryResponse struct {
	Period        string              `json:"period"`
	TotalExpense  decimal.Decimal     `json:"total_expense"`
	TotalIncome   decimal.Decimal     `json:"total_income"`
	BillCount     int64               `json:"bill_count"`
	DailyAverage  decimal.Decimal     `json:"daily_average"`
	TopCategories []CategoryStatsItem `json:"top_categories"`
	Trend         []TrendItem         `json:"trend"`
}

// CategoryStatsItem 分类统计项
type CategoryStatsItem struct {
	ID      uint64          `json:"id"`
	Name    string          `json:"name"`
	Amount  decimal.Decimal `json:"amount"`
	Percent float64         `json:"percent"`
}

// TrendItem 趋势项
type TrendItem struct {
	Date    string          `json:"date"`
	Expense decimal.Decimal `json:"expense"`
	Income  decimal.Decimal `json:"income"`
}

// CategoryStatsResponse 分类统计响应
type CategoryStatsResponse struct {
	Period     string              `json:"period"`
	Categories []CategoryStatsItem `json:"categories"`
}

// =============== 分类相关 ===============

// CategoryResponse 分类响应
type CategoryResponse struct {
	ID        uint64             `json:"id"`
	Name      string             `json:"name"`
	ParentID  uint64             `json:"parent_id"`
	Icon      string             `json:"icon"`
	SortOrder int                `json:"sort_order"`
	IsSystem  bool               `json:"is_system"`
	Children  []CategoryResponse `json:"children,omitempty"`
}

// =============== 批量识别相关 ===============

// BatchRecognizeResponse 批量识别响应
type BatchRecognizeResponse struct {
	Total        int               `json:"total"`         // 总数
	SuccessCount int               `json:"success_count"` // 成功数
	FailCount    int               `json:"fail_count"`    // 失败数
	Results      []BatchItemResult `json:"results"`       // 结果列表
}

// BatchItemResult 批量识别单项结果
type BatchItemResult struct {
	Index    int                  `json:"index"`           // 图片索引
	FileName string               `json:"file_name"`       // 文件名
	Success  bool                 `json:"success"`         // 是否成功
	Data     *AIRecognizeResponse `json:"data,omitempty"`  // 识别结果
	Error    string               `json:"error,omitempty"` // 错误信息
	Duration int64                `json:"duration"`        // 处理耗时(ms)
}

// BatchRecognizeAndSaveResponse 批量识别并保存响应
type BatchRecognizeAndSaveResponse struct {
	Total        int                   `json:"total"`         // 总数
	SuccessCount int                   `json:"success_count"` // 成功数
	FailCount    int                   `json:"fail_count"`    // 失败数
	Results      []BatchItemSaveResult `json:"results"`       // 结果列表
}

// BatchItemSaveResult 批量保存单项结果
type BatchItemSaveResult struct {
	Index    int           `json:"index"`           // 图片索引
	FileName string        `json:"file_name"`       // 文件名
	Success  bool          `json:"success"`         // 是否成功
	Bill     *BillResponse `json:"bill,omitempty"`  // 账单（成功时）
	Error    string        `json:"error,omitempty"` // 错误信息
}
