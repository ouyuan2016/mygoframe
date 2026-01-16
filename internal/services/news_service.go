package services

import (
	"context"
	"errors"

	"mygoframe/internal/models"
	"mygoframe/internal/repositories"

	"gorm.io/gorm"
)

type NewsService interface {
	GetNewsByID(ctx context.Context, id uint) (*models.News, error)
	GetNewsList(ctx context.Context, page, pageSize int) ([]*models.News, int64, error)
}

type newsService struct {
	repo repositories.NewsRepository
}

func NewNewsService(db *gorm.DB) NewsService {
	return &newsService{repo: repositories.NewNewsRepository(db)}
}

func (s *newsService) GetNewsByID(ctx context.Context, id uint) (*models.News, error) {
	news, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("快讯不存在")
	}
	return news, nil
}

func (s *newsService) GetNewsList(ctx context.Context, page, pageSize int) ([]*models.News, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	list, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
