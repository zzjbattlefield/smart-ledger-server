package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// =============== 用户相关 ===============

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"max=50"`
}

// UpdateProfileRequest 更新用户信息请求
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" binding:"max=50"`
	AvatarURL string `json:"avatar_url" binding:"max=255"`
}

// =============== 账单相关 ===============

// CreateBillRequest 创建账单请求
type CreateBillRequest struct {
	Amount     decimal.Decimal `json:"amount" binding:"required"`
	BillType   int             `json:"bill_type" binding:"required,oneof=1 2"`
	Platform   string          `json:"platform" binding:"max=50"`
	Merchant   string          `json:"merchant" binding:"max=255"`
	CategoryID *uint64         `json:"category_id"`
	PayTime    time.Time       `json:"pay_time" binding:"required"`
	PayMethod  string          `json:"pay_method" binding:"max=50"`
	OrderNo    string          `json:"order_no" binding:"max=100"`
	Remark     string          `json:"remark" binding:"max=500"`
	Items      []BillItemDTO   `json:"items"`
}

// UpdateBillRequest 更新账单请求
type UpdateBillRequest struct {
	Amount      decimal.Decimal `json:"amount"`
	BillType    int             `json:"bill_type" binding:"omitempty,oneof=1 2"`
	Platform    string          `json:"platform" binding:"max=50"`
	Merchant    string          `json:"merchant" binding:"max=255"`
	CategoryID  *uint64         `json:"category_id"`
	PayTime     *time.Time      `json:"pay_time"`
	PayMethod   string          `json:"pay_method" binding:"max=50"`
	OrderNo     string          `json:"order_no" binding:"max=100"`
	Remark      string          `json:"remark" binding:"max=500"`
	IsConfirmed *bool           `json:"is_confirmed"`
}

// BillItemDTO 账单明细DTO
type BillItemDTO struct {
	Name     string          `json:"name" binding:"required,max=255"`
	Price    decimal.Decimal `json:"price"`
	Quantity int             `json:"quantity" binding:"min=1"`
}

// BillListRequest 账单列表请求
type BillListRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	PageSize   int    `form:"page_size" binding:"min=1,max=100"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	CategoryID uint64 `form:"category_id"`
	BillType   int    `form:"bill_type" binding:"omitempty,oneof=1 2"`
	Keyword    string `form:"keyword" binding:"max=100"`
}

// SetDefaults 设置默认值
func (r *BillListRequest) SetDefaults() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
}

// =============== 统计相关 ===============

// StatsSummaryRequest 统计摘要请求
type StatsSummaryRequest struct {
	Period string `form:"period" binding:"required,oneof=day week month year"`
	Date   string `form:"date" binding:"required"`
}

// StatsCategoryRequest 分类统计请求
type StatsCategoryRequest struct {
	Period string `form:"period" binding:"required,oneof=day week month year"`
	Date   string `form:"date" binding:"required"`
}

// =============== 分类相关 ===============

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name      string `json:"name" binding:"required,max=50"`
	ParentID  uint64 `json:"parent_id"`
	Icon      string `json:"icon" binding:"max=100"`
	SortOrder int    `json:"sort_order"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name      string `json:"name" binding:"max=50"`
	Icon      string `json:"icon" binding:"max=100"`
	SortOrder *int   `json:"sort_order"`
}
