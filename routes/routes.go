package routes

import (
	"mygoframe/pkg/config"
	"mygoframe/routes/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB) *gin.Engine {
	if config.GetBuildMode() == "pro" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger()) // 请求日志
	r.Use(middleware.Cors())
	r.Use(middleware.Recovery())

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		InitNewsRoutes(apiGroup, db)
		SetupUserRoutes(apiGroup, db)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Not Found", "data": nil})
	})

	return r
}
