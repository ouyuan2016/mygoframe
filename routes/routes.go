package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mygoframe/pkg/config"
	"mygoframe/pkg/logger"
	"mygoframe/routes/middleware"
)

func SetupRoutes(db *gorm.DB) *gin.Engine {
	if config.GetBuildMode() == "pro" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(logger.GinLogger())
	r.Use(middleware.Cors())
	r.Use(gin.Recovery())

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		InitNewsRoutes(apiGroup, db)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Not Found"})
	})

	return r
}
