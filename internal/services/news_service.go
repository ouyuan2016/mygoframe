package services

import (
	"github.com/ouyuan2016/mygoframe/internal/models"
	"github.com/ouyuan2016/mygoframe/internal/repositories"
)

// NewsService 快讯服务接口
type NewsService interface {
	GetNewsByID(id uint) (*models.News, error)
	GetNewsList(category int, isImportant bool, page, pageSize int) ([]models.News, int64, error)
}

// newsService 快讯服务实现
type newsService struct {
	newsRepo repositories.NewsRepository
}

// NewNewsService 创建快讯服务实例
func NewNewsService(newsRepo repositories.NewsRepository) NewsService {
	return &newsService{newsRepo: newsRepo}
}

// GetNewsByID 根据ID获取快讯
func (s *newsService) GetNewsByID(id uint) (*models.News, error) {
	return s.newsRepo.GetByID(id)
}

// GetNewsList 获取快讯列表
func (s *newsService) GetNewsList(category int, isImportant bool, page, pageSize int) ([]models.News, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	newsList, err := s.newsRepo.List(category, isImportant, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.newsRepo.Count(category, isImportant)
	if err != nil {
		return nil, 0, err
	}

	return newsList, count, nil
}
