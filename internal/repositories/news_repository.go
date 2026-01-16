package repositories

import (
	"context"
	"mygoframe/internal/models"

	"gorm.io/gorm"
)

type NewsRepository interface {
	Create(ctx context.Context, news *models.News) error
	GetByID(ctx context.Context, id uint) (*models.News, error)
	List(ctx context.Context, offset, limit int) ([]*models.News, error)
	Update(ctx context.Context, news *models.News) error
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
}

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Create(ctx context.Context, news *models.News) error {
	return r.db.WithContext(ctx).Create(news).Error
}

func (r *newsRepository) GetByID(ctx context.Context, id uint) (*models.News, error) {
	var news models.News
	err := r.db.WithContext(ctx).First(&news, id).Error
	if err != nil {
		return nil, err
	}
	return &news, nil
}

func (r *newsRepository) List(ctx context.Context, offset, limit int) ([]*models.News, error) {
	var newsList []*models.News
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&newsList).Error
	return newsList, err
}

func (r *newsRepository) Update(ctx context.Context, news *models.News) error {
	return r.db.WithContext(ctx).Save(news).Error
}

func (r *newsRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.News{}, id).Error
}

func (r *newsRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.News{}).Count(&count).Error
	return count, err
}
