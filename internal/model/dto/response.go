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
	ID          uint64            `json:"id"`
	UUID        string            `json:"uuid"`
	Amount      decimal.Decimal   `json:"amount"`
	BillType    int               `json:"bill_type"`
	Platform    string            `json:"platform"`
	Merchant    string            `json:"merchant"`
	Category    *CategoryResponse `json:"category"`
	PayTime     time.Time         `json:"pay_time"`
	PayMethod   string            `json:"pay_method"`
	OrderNo     string            `json:"order_no"`
	Remark      string            `json:"remark"`
	Confidence  float64           `json:"confidence"`
	IsConfirmed bool              `json:"is_confirmed"`
	CreatedAt   time.Time         `json:"created_at"`
}

// BillListResponse 账单列表响应
type BillListResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	List     []BillResponse `json:"list"`
}

// BillImportResponse 账单导入响应
type BillImportResponse struct {
	Total  int           `json:"total"`
	Failed int           `json:"failed"`
	Errors []ImportError `json:"errors"`
}

// ImportError 导入错误详情
type ImportError struct {
	Row     int               `json:"row"`
	RowData map[string]string `json:"row_data,omitempty"`
	Column  string            `json:"column,omitempty"`
	Message string            `json:"message"`
}

// =============== AI 识别相关 ===============

// AIRecognizeResponse AI识别响应
type AIRecognizeResponse struct {
	Platform    string          `json:"platform"`
	Amount      decimal.Decimal `json:"amount"`
	Merchant    string          `json:"merchant"`
	Category    string          `json:"category"`
	SubCategory string          `json:"sub_category"`
	PayTime     string          `json:"pay_time"`
	PayMethod   string          `json:"pay_method"`
	OrderNo     string          `json:"order_no"`
	Confidence  float64         `json:"confidence"`
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
	Date    DateOnly        `json:"date"`
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
	Children  []CategoryResponse `json:"children,omitempty"`
}

type DateOnly time.Time

func (d *DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*d).Format("2006-01-02") + `"`), nil
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"2006-01-02"`, string(data))
	if err != nil {
		return err
	}
	*d = DateOnly(t)
	return nil
}
