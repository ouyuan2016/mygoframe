package routes

import (
	"log"
	"os"

	"github.com/ouyuan2016/mygoframe/pkg/utils"
	"github.com/ouyuan2016/mygoframe/routes/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes() *gin.Engine {
	if envMode := os.Getenv("BUILD_MODE"); envMode != "" {

		var configName string
		switch envMode {
		case "pro", "prod", "production":
			configName = gin.ReleaseMode
		case "test", "testing":
			configName = gin.TestMode
		default:
			configName = gin.DebugMode
		}
		log.Printf("设置Gin模式为: %s", configName)
		gin.SetMode(configName)
	}
	// 创建Gin引擎
	r := gin.Default()

	r.SetTrustedProxies(nil)
	// 添加中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.Use(middleware.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		utils.Success(c, "ok")
	})

	apirg := r.Group("/api/v1")
	InitNewsRoutes(apirg)
	InitMomentRoutes(apirg)

	// 404错误处理 - 放在最后
	r.NoRoute(func(c *gin.Context) {
		utils.NotFound(c, "Page not found")
	})

	return r
}
