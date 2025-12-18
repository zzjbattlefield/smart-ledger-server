package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// BillType 账单类型
type BillType int

const (
	BillTypeExpense BillType = 1 // 支出
	BillTypeIncome  BillType = 2 // 收入
)

// Bill 账单模型
type Bill struct {
	BaseModel
	UUID          string          `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"` // 账单唯一标识（UUID格式）
	UserID        uint64          `gorm:"index;not null" json:"user_id"`                     // 所属用户ID
	Amount        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`         // 账单总金额
	BillType      BillType        `gorm:"default:1" json:"bill_type"`                        // 账单类型：1-支出，2-收入
	Platform      string          `gorm:"type:varchar(50)" json:"platform"`                  // 支付平台（如：微信、支付宝）
	Merchant      string          `gorm:"type:varchar(255)" json:"merchant"`                 // 商户名称
	CategoryID    *uint64         `gorm:"index" json:"category_id"`                          // 分类ID（可为空）
	PayTime       time.Time       `gorm:"type:datetime;not null;index" json:"pay_time"`      // 支付时间
	PayMethod     string          `gorm:"type:varchar(50)" json:"pay_method"`                // 支付方式（如：余额、银行卡）
	OrderNo       string          `gorm:"type:varchar(100)" json:"order_no"`                 // 订单号
	Remark        string          `gorm:"type:varchar(500)" json:"remark"`                   // 备注信息
	ImagePath     string          `gorm:"type:varchar(255)" json:"image_path"`               // 支付截图路径
	AIRawResponse string          `gorm:"type:text" json:"-"`                                // AI识别原始响应（不输出到JSON）
	Confidence    float64         `gorm:"type:decimal(3,2)" json:"confidence"`               // AI识别置信度（0-1）
	IsConfirmed   bool            `gorm:"default:false" json:"is_confirmed"`                 // 是否已确认（用户确认AI识别结果）

	// 关联
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`         // 所属用户
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"` // 所属分类
}

// TableName 指定表名
func (Bill) TableName() string {
	return "bills"
}
