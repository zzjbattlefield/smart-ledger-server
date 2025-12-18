package model

// CategoryTemplate 分类模板
type CategoryTemplate struct {
	BaseModel
	Name      string       `gorm:"type:varchar(50);not null" json:"name"`
	Type      CategoryType `gorm:"type:tinyint;not null;default:1" json:"type"`
	ParentID  uint64       `gorm:"default:0;index" json:"parent_id"`
	Icon      string       `gorm:"type:varchar(100)" json:"icon"`
	SortOrder int          `gorm:"default:0" json:"sort_order"`
}

// TableName 指定表名
func (CategoryTemplate) TableName() string {
	return "category_templates"
}
