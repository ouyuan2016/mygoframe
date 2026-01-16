package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"mygoframe/pkg/logger"
	"mygoframe/pkg/utils"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "Access token not provided")
			c.Abort()
			return
		}

		// 检查Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Unauthorized(c, "Invalid access token format")
			c.Abort()
			return
		}

		// 获取JWT工具实例
		jwtUtil, err := utils.GetJWTUtil()
		if err != nil {
			logger.Error("JWT工具初始化失败", logger.Err(err))
			utils.ServerError(c, "Authentication service error")
			c.Abort()
			return
		}

		// 解析令牌
		claims, err := jwtUtil.ParseToken(parts[1])
		if err != nil {
			logger.Warn("令牌解析失败", logger.Err(err))
			utils.Unauthorized(c, "Access token is invalid or expired")
			c.Abort()
			return
		}

		// 验证用户ID
		if claims.UserInfo.Id == "" {
			utils.Unauthorized(c, "Invalid user information")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		//c.Set("user_id", claims.UserInfo.Id)
		//c.Set("display_name", claims.UserInfo.DisplayName)
		//c.Set("avatar", claims.UserInfo.Avatar)
		//c.Set("email", claims.UserInfo.Email)
		//c.Set("phone", claims.UserInfo.Phone)
		//c.Set("claims", claims)

		// 继续处理请求
		c.Next()
	}
}

// OptionalJWTAuth 可选的JWT认证中间件
// 如果提供了有效的令牌，则解析用户信息；否则继续处理但不设置用户信息
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		jwtUtil, err := utils.GetJWTUtil()
		if err != nil {
			c.Next()
			return
		}

		claims, err := jwtUtil.ParseToken(parts[1])
		if err != nil {
			c.Next()
			return
		}

		if claims.UserInfo.Id != "" {
			c.Set("user_id", claims.UserInfo.Id)
			c.Set("display_name", claims.UserInfo.DisplayName)
			c.Set("avatar", claims.UserInfo.Avatar)
			c.Set("email", claims.UserInfo.Email)
			c.Set("phone", claims.UserInfo.Phone)
			c.Set("claims", claims)
		}

		c.Next()
	}
}

// GetCurrentUser 获取当前登录用户信息
func GetCurrentUser(c *gin.Context) (userID string, displayName string, email string, exists bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return "", "", "", false
	}

	displayNameInterface, _ := c.Get("display_name")
	emailInterface, _ := c.Get("email")

	userID, ok := userIDInterface.(string)
	if !ok {
		return "", "", "", false
	}

	displayName, _ = displayNameInterface.(string)
	email, _ = emailInterface.(string)

	return userID, displayName, email, true
}

// GetCurrentClaims 获取当前用户的JWT声明
func GetCurrentClaims(c *gin.Context) (*utils.Claims, bool) {
	claimsInterface, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	claims, ok := claimsInterface.(*utils.Claims)
	if !ok {
		return nil, false
	}

	return claims, true
}

// RequireRole 角色权限检查中间件（示例）
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetCurrentClaims(c)
		if !exists {
			utils.Unauthorized(c, "User information does not exist")
			c.Abort()
			return
		}

		// 这里可以根据需要扩展角色验证逻辑
		// 例如从数据库获取用户角色进行验证
		userRole := "user" // 示例角色

		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
