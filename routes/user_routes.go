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
		public.POST("/send-sms-code", userHandler.SendSMSCode)         // 发送短信验证码
		public.POST("/verify-sms-code", userHandler.VerifySMSCode)     // 验证短信验证码
		public.POST("/send-email-code", userHandler.SendEmailCode)     // 发送邮箱验证码
		public.POST("/verify-email-code", userHandler.VerifyEmailCode) // 验证邮箱验证码

		// 队列测试
		public.POST("/test_queue", userHandler.TestQueue)
		public.POST("/delayed_task", userHandler.EnqueueDelayedTask)
	}

	protected := router.Group("/users")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/logout", userHandler.Logout)
		protected.GET("/profile", userHandler.GetProfile)
	}
}
