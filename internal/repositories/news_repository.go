package repositories

import (
	"github.com/ouyuan2016/mygoframe/internal/models"
	"github.com/ouyuan2016/mygoframe/pkg/database"
	"gorm.io/gorm"
)

// NewsRepository 快讯仓库接口
type NewsRepository interface {
	Create(news *models.News) error
	GetByID(id uint) (*models.News, error)
	List(category int, isImportant bool, offset, limit int) ([]models.News, error)
	Update(news *models.News) error
	Delete(id uint) error
	Count(category int, isImportant bool) (int64, error)
}

// newsRepository 快讯仓库实现
type newsRepository struct {
	db *gorm.DB
}

// NewNewsRepository 创建快讯仓库实例
func NewNewsRepository() NewsRepository {
	return &newsRepository{db: database.DB}
}

// Create 创建新快讯
func (r *newsRepository) Create(news *models.News) error {
	return r.db.Create(news).Error
}

// GetByID 根据ID获取快讯
func (r *newsRepository) GetByID(id uint) (*models.News, error) {
	var news models.News
	if err := r.db.First(&news, id).Error; err != nil {
		return nil, err
	}
	return &news, nil
}

// List 获取快讯列表
func (r *newsRepository) List(category int, isImportant bool, offset, limit int) ([]models.News, error) {
	var newsList []models.News
	query := r.db.Order("created_at DESC")

	if category > 0 {
		query = query.Where("category = ?", category)
	}

	if isImportant {
		query = query.Where("is_important = ?", true)
	}

	if err := query.Offset(offset).Limit(limit).Find(&newsList).Error; err != nil {
		return nil, err
	}

	return newsList, nil
}

// Update 更新快讯
func (r *newsRepository) Update(news *models.News) error {
	return r.db.Save(news).Error
}

// Delete 删除快讯
func (r *newsRepository) Delete(id uint) error {
	return r.db.Delete(&models.News{}, id).Error
}

// Count 统计快讯数量
func (r *newsRepository) Count(category int, isImportant bool) (int64, error) {
	var count int64
	query := r.db.Model(&models.News{})

	if category > 0 {
		query = query.Where("category = ?", category)
	}

	if isImportant {
		query = query.Where("is_important = ?", true)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
