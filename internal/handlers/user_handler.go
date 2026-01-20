package handlers

import (
	"fmt"
	"strings"

	"mygoframe/internal/dto"
	"mygoframe/internal/services"
	"mygoframe/internal/task"
	"mygoframe/pkg/queue"
	"mygoframe/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
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
	// 从查询参数获取队列名称，默认为 "default"  default 队列权重为 3 ，critical 队列权重为 6 low 队列权重为 1
	queueName := c.DefaultQuery("queue", "default")

	// 模拟一个用户ID
	userID := 123
	// 创建一个新任务
	t, err := task.NewWelcomeEmailTask(userID)
	if err != nil {
		utils.ServerError(c, "创建任务失败: "+err.Error())
		return
	}
	// 入队时指定队列
	info, err := queue.Client.Enqueue(t, asynq.MaxRetry(3), asynq.Timeout(20*time.Minute), asynq.Queue(queueName))
	if err != nil {
		utils.ServerError(c, "入队失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{
		"task_id": info.ID,
		"queue":   queueName,
	})
}

// EnqueueDelayedTask 测试延迟队列
func (h *UserHandler) EnqueueDelayedTask(c *gin.Context) {
	// 模拟一个用户ID
	userID := "123"
	t, err := task.NewSendLaterEmailTask(userID, time.Now().Add(3*time.Second))
	if err != nil {
		utils.ServerError(c, "创建任务失败: "+err.Error())
		return
	}
	fmt.Println("start time:", time.Now().Format("2006-01-02 15:04:05"))
	// Enqueue the task to be processed after the specified delay.
	info, err := queue.Client.Enqueue(t, asynq.ProcessIn(3*time.Second))
	if err != nil {
		utils.ServerError(c, "入队失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"task_id":    info.ID,
		"process_in": "3s",
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
