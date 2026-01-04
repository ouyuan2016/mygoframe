package models

import (
	"time"

	"gorm.io/gorm"
)

// News 快讯模型
type News struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Source      string         `gorm:"size:200;not null" json:"source"`
	Category    int            `gorm:"type:int;default:1" json:"category"` // 1:快讯
	IsImportant bool           `gorm:"default:false" json:"is_important"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (News) TableName() string {
	return "news"
}
