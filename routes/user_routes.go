package routes

import (
	"mygoframe/internal/handlers"
	"mygoframe/routes/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupUserRoutes 设置用户相关路由
func SetupUserRoutes(router *gin.RouterGroup, db *gorm.DB) {
	userHandler := handlers.NewUserHandler(db)

	public := router.Group("/users")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", userHandler.Login)
		public.POST("/refresh-token", userHandler.RefreshToken)
	}

	protected := router.Group("/users")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/logout", userHandler.Logout)
		protected.GET("/profile", userHandler.GetProfile)
	}
}
