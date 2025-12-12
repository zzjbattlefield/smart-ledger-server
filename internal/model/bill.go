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
	UUID          string          `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UserID        uint64          `gorm:"index;not null" json:"user_id"`
	Amount        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	BillType      BillType        `gorm:"default:1" json:"bill_type"`
	Platform      string          `gorm:"type:varchar(50)" json:"platform"`
	Merchant      string          `gorm:"type:varchar(255)" json:"merchant"`
	CategoryID    *uint64         `gorm:"index" json:"category_id"`
	PayTime       time.Time       `gorm:"type:datetime;not null;index" json:"pay_time"`
	PayMethod     string          `gorm:"type:varchar(50)" json:"pay_method"`
	OrderNo       string          `gorm:"type:varchar(100)" json:"order_no"`
	Remark        string          `gorm:"type:varchar(500)" json:"remark"`
	ImagePath     string          `gorm:"type:varchar(255)" json:"image_path"`
	AIRawResponse string          `gorm:"type:text" json:"-"`
	Confidence    float64         `gorm:"type:decimal(3,2)" json:"confidence"`
	IsConfirmed   bool            `gorm:"default:false" json:"is_confirmed"`

	// 关联
	User     *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Items    []BillItem `gorm:"foreignKey:BillID" json:"items,omitempty"`
}

// TableName 指定表名
func (Bill) TableName() string {
	return "bills"
}

// BillItem 账单明细
type BillItem struct {
	BaseModel
	BillID   uint64          `gorm:"index;not null" json:"bill_id"`
	Name     string          `gorm:"type:varchar(255);not null" json:"name"`
	Price    decimal.Decimal `gorm:"type:decimal(10,2)" json:"price"`
	Quantity int             `gorm:"default:1" json:"quantity"`
}

// TableName 指定表名
func (BillItem) TableName() string {
	return "bill_items"
}
