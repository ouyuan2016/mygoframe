package handlers

import (
	"strings"

	"mygoframe/internal/dto"
	"mygoframe/internal/services"
	"mygoframe/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(db),
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数验证失败")
		return
	}

	user, err := h.userService.Register(c, req) // 传递context
	if err != nil {
		utils.BadRequest(c, "注册失败: "+err.Error())
		return
	}

	response := dto.UserRegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}

	utils.Success(c, response)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数验证失败")
		return
	}

	resp, err := h.userService.Login(c, req) // 传递context
	if err != nil {
		utils.Unauthorized(c, "登录失败: "+err.Error())
		return
	}

	utils.Success(c, resp)
}

// RefreshToken 刷新访问令牌
func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("Authorization")
	if refreshToken == "" {
		utils.Unauthorized(c, "缺少刷新令牌")
		return
	}

	refreshToken = strings.TrimPrefix(refreshToken, "Bearer ")

	resp, err := h.userService.RefreshToken(c, refreshToken) // 传递context
	if err != nil {
		utils.Unauthorized(c, "刷新令牌失败: "+err.Error())
		return
	}

	utils.Success(c, resp)
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		utils.Unauthorized(c, "缺少访问令牌")
		return
	}

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	if err := h.userService.Logout(c, accessToken); err != nil { // 传递context
		utils.BadRequest(c, "登出失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetProfile 获取用户信息（需要登录）
func (h *UserHandler) GetProfile(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		utils.Unauthorized(c, "未授权，请先登录")
		return
	}

	userClaims, ok := claims.(*utils.UserInfo)
	if !ok {
		utils.Unauthorized(c, "令牌格式错误")
		return
	}

	user, err := h.userService.GetUserByID(c, userClaims.Id) // 传递context
	if err != nil {
		utils.ServerError(c, "获取用户信息失败: "+err.Error())
		return
	}

	if user == nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	response := dto.UserInfoResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	utils.Success(c, response)
}
