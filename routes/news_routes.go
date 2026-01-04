package routes

import (
	"github.com/ouyuan2016/mygoframe/internal/handlers"
	"github.com/ouyuan2016/mygoframe/internal/repositories"
	"github.com/ouyuan2016/mygoframe/internal/services"
	"github.com/gin-gonic/gin"
)

// InitNewsRoutes 初始化快讯路由
func InitNewsRoutes(router *gin.RouterGroup) {
	// 初始化仓储
	newsRepo := repositories.NewNewsRepository()
	// 初始化服务
	newsService := services.NewNewsService(newsRepo)
	// 初始化处理器
	newsHandler := handlers.NewNewsHandler(newsService)

	// 快讯路由组
	newsGroup := router.Group("/news")
	{
		// 获取快讯列表
		newsGroup.GET("", newsHandler.GetNewsList)
		// 根据ID获取快讯
		newsGroup.GET("/:id", newsHandler.GetNewsByID)
	}
}
