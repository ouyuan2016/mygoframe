package handlers

import (
	"strings"

	"mygoframe/internal/dto"
	"mygoframe/internal/services"
	"mygoframe/internal/task"
	"mygoframe/pkg/utils"
	"time"

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

// TestQueue 测试队列
func (h *UserHandler) TestQueue(c *gin.Context) {
	userID := 123
	queueName := "critical"

	info, err := task.EnqueueWelcomeEmailTask(userID, queueName)
	if err != nil {
		utils.ServerError(c, "任务入队失败")
		return
	}

	utils.Success(c, gin.H{
		"message":   "任务已成功入队",
		"task_id":   info.ID,
		"queue":     info.Queue,
		"retention": info.Retention,
	})
}

// EnqueueDelayedTask handles enqueuing a delayed task.
func (h *UserHandler) EnqueueDelayedTask(c *gin.Context) {
	userID := "123"
	delay := 3 * time.Second

	info, err := task.EnqueueSendLaterEmailTask(userID, delay)
	if err != nil {
		utils.ServerError(c, "延迟任务入队失败")
		return
	}

	utils.Success(c, gin.H{
		"message":    "延迟任务已成功入队",
		"task_id":    info.ID,
		"process_in": delay.String(),
		"process_at": info.NextProcessAt.Format(time.RFC3339),
		"retention":  info.Retention,
	})
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

// SendEmailCode 发送邮箱验证码
func (h *UserHandler) SendEmailCode(c *gin.Context) {
	var req dto.SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数验证失败")
		return
	}

	resp, err := h.userService.SendEmailCode(c, req)
	if err != nil {
		utils.ServerError(c, "发送验证码失败: "+err.Error())
		return
	}

	utils.Success(c, resp)
}

// VerifyEmailCode 验证邮箱验证码
func (h *UserHandler) VerifyEmailCode(c *gin.Context) {
	var req dto.VerifyEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数验证失败")
		return
	}

	resp, err := h.userService.VerifyEmailCode(c, req)
	if err != nil {
		utils.ServerError(c, "验证验证码失败: "+err.Error())
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

// SendSMSCode 发送短信验证码
func (h *UserHandler) SendSMSCode(c *gin.Context) {
	var req dto.SendSMSCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数验证失败")
		return
	}

	resp, err := h.userService.SendSMSCode(c, req)
	if err != nil {
		utils.ServerError(c, "发送验证码失败: "+err.Error())
		return
	}

	utils.Success(c, resp)
}

// VerifySMSCode 验证短信验证码
func (h *UserHandler) VerifySMSCode(c *gin.Context) {
	var req dto.VerifySMSCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数验证失败")
		return
	}

	resp, err := h.userService.VerifySMSCode(c, req)
	if err != nil {
		utils.ServerError(c, "验证验证码失败: "+err.Error())
		return
	}

	utils.Success(c, resp)
}
