package services

import (
	"context"
	"errors"

	"mygoframe/internal/models"
	"mygoframe/internal/repositories"
)

type NewsService interface {
	GetNewsByID(ctx context.Context, id uint) (*models.News, error)
	GetNewsList(ctx context.Context, page, pageSize int) ([]*models.News, int64, error)
}

type newsService struct {
	repo repositories.NewsRepository
}

func NewNewsService(repo repositories.NewsRepository) NewsService {
	return &newsService{repo: repo}
}

func (s *newsService) GetNewsByID(ctx context.Context, id uint) (*models.News, error) {
	news, err := s.repo.GetByID(id)
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

	total, err := s.repo.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	list, err := s.repo.List(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
