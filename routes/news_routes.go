package routes

import (
	"mygoframe/internal/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitNewsRoutes 初始化新闻路由 - 改为与NewUserHandler相同的模式
func InitNewsRoutes(router *gin.RouterGroup, db *gorm.DB) {
	newsHandler := handlers.NewNewsHandler(db)

	newsGroup := router.Group("/news")
	{
		newsGroup.GET("", newsHandler.GetNewsList)
		newsGroup.GET("/:id", newsHandler.GetNewsByID)
	}
}
