package repositories

import (
	"gorm.io/gorm"
	"mygoframe/internal/models"
)

type NewsRepository interface {
	Create(news *models.News) error
	GetByID(id uint) (*models.News, error)
	List(offset, limit int) ([]*models.News, error)
	Update(news *models.News) error
	Delete(id uint) error
	Count() (int64, error)
}

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Create(news *models.News) error {
	return r.db.Create(news).Error
}

func (r *newsRepository) GetByID(id uint) (*models.News, error) {
	var news models.News
	err := r.db.First(&news, id).Error
	if err != nil {
		return nil, err
	}
	return &news, nil
}

func (r *newsRepository) List(offset, limit int) ([]*models.News, error) {
	var newsList []*models.News
	err := r.db.Offset(offset).Limit(limit).Find(&newsList).Error
	return newsList, err
}

func (r *newsRepository) Update(news *models.News) error {
	return r.db.Save(news).Error
}

func (r *newsRepository) Delete(id uint) error {
	return r.db.Delete(&models.News{}, id).Error
}

func (r *newsRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.News{}).Count(&count).Error
	return count, err
}
