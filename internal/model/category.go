package model

// Category 分类模型
type Category struct {
	BaseModel
	Name      string `gorm:"type:varchar(50);not null" json:"name"`
	ParentID  uint64 `gorm:"default:0;index" json:"parent_id"`
	Icon      string `gorm:"type:varchar(100)" json:"icon"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	IsSystem  bool   `gorm:"default:false" json:"is_system"`

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
