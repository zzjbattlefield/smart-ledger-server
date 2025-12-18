package model

// CategoryType 分类类型
type CategoryType int

const (
	CategoryTypeExpense CategoryType = 1 // 支出
	CategoryTypeIncome  CategoryType = 2 // 收入
)

// Category 分类模型
type Category struct {
	BaseModel
	Name      string       `gorm:"type:varchar(50);not null" json:"name"`
	Type      CategoryType `gorm:"type:tinyint;not null;default:1" json:"type"`
	UserID    uint64       `gorm:"index;not null" json:"user_id"`
	ParentID  uint64       `gorm:"default:0;index" json:"parent_id"`
	Icon      string       `gorm:"type:varchar(100)" json:"icon"`
	SortOrder int          `gorm:"default:0" json:"sort_order"`

	// 关联
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}

// IsTopLevel 是否为一级分类
func (c *Category) IsTopLevel() bool {
	return c.ParentID == 0
}
