package model

import (
	"time"
)

// User 用户模型
type User struct {
	BaseModel
	Phone       string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	Password    string     `gorm:"type:varchar(255);not null" json:"-"`
	Nickname    string     `gorm:"type:varchar(50)" json:"nickname"`
	AvatarURL   string     `gorm:"type:varchar(255)" json:"avatar_url"`
	LastLoginAt *time.Time `gorm:"type:datetime" json:"last_login_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
