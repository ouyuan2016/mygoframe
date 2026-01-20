package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Email     string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_email" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Avatar    string         `gorm:"type:varchar(255)" json:"avatar"`
	Status    string         `gorm:"type:varchar(20);default:active" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (u *User) IsActive() bool {
	return u.Status == "active"
}
