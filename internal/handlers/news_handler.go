package handlers

import (
	"strconv"

	"mygoframe/internal/services"
	"mygoframe/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NewsHandler struct {
	service services.NewsService
}

func NewNewsHandler(db *gorm.DB) *NewsHandler {
	return &NewsHandler{
		service: services.NewNewsService(db),
	}
}

func (h *NewsHandler) GetNewsByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效ID")
		return
	}

	news, err := h.service.GetNewsByID(c, uint(id))
	if err != nil {
		utils.NotFound(c, "快讯不存在")
		return
	}

	utils.Success(c, news)
}

func (h *NewsHandler) GetNewsList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	news, total, err := h.service.GetNewsList(c, page, pageSize)
	if err != nil {
		utils.ServerError(c, "获取快讯列表失败")
		return
	}

	utils.Success(c, gin.H{
		"list":     news,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
