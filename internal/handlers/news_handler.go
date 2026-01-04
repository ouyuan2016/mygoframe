package handlers

import (
	"strconv"

	"github.com/ouyuan2016/mygoframe/pkg/utils"
	"github.com/gin-gonic/gin"

	"github.com/ouyuan2016/mygoframe/internal/services"
)

// NewsHandler 快讯处理器
type NewsHandler struct {
	newsService services.NewsService
}

// NewNewsHandler 创建快讯处理器实例
func NewNewsHandler(newsService services.NewsService) *NewsHandler {
	return &NewsHandler{newsService: newsService}
}

// GetNewsByID 根据ID获取快讯
func (h *NewsHandler) GetNewsByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid news ID")
		return
	}

	news, err := h.newsService.GetNewsByID(uint(id))
	if err != nil {
		utils.NotFound(c, "News not found")
		return
	}

	utils.Success(c, news)
}

// GetNewsList 获取快讯列表
func (h *NewsHandler) GetNewsList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category, _ := strconv.Atoi(c.DefaultQuery("category", "1"))
	isImportantStr := c.Query("is_important")

	var isImportant bool
	if isImportantStr != "" {
		b, err := strconv.ParseBool(isImportantStr)
		if err != nil {
			utils.BadRequest(c, "Invalid is_important parameter")
			return
		}
		isImportant = b
	}

	newsList, total, err := h.newsService.GetNewsList(category, isImportant, page, pageSize)
	if err != nil {
		utils.ServerError(c, "Failed to retrieve news list")
		return
	}

	utils.Success(c, gin.H{
		"news_list": newsList,
		"total":     total,
	})
}
