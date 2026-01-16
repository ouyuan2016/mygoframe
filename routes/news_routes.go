package routes

import (
	"mygoframe/internal/handlers"
	"mygoframe/internal/repositories"
	"mygoframe/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitNewsRoutes(router *gin.RouterGroup, db *gorm.DB) {
	newsRepo := repositories.NewNewsRepository(db)
	newsService := services.NewNewsService(newsRepo)
	newsHandler := handlers.NewNewsHandler(newsService)

	newsGroup := router.Group("/news")
	{
		newsGroup.GET("", newsHandler.GetNewsList)
		newsGroup.GET("/:id", newsHandler.GetNewsByID)
	}
}
