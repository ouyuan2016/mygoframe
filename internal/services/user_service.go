package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"mygoframe/internal/dto"
	"mygoframe/internal/models"
	"mygoframe/internal/repositories"
	"mygoframe/pkg/cache"
	"mygoframe/pkg/utils"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	Register(ctx context.Context, req dto.UserRegisterRequest) (*models.User, error)
	Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserLoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, accessToken string) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	SendSMSCode(ctx context.Context, req dto.SendSMSCodeRequest) (*dto.SendSMSCodeResponse, error)
	VerifySMSCode(ctx context.Context, req dto.VerifySMSCodeRequest) (*dto.VerifySMSCodeResponse, error)
	SendEmailCode(ctx context.Context, req dto.SendEmailCodeRequest) (*dto.SendEmailCodeResponse, error)
	VerifyEmailCode(ctx context.Context, req dto.VerifyEmailCodeRequest) (*dto.VerifyEmailCodeResponse, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repositories.UserRepository
	jwtUtil  *utils.JWTUtil
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB) UserService {
	jwtUtil, _ := utils.GetJWTUtil()
	return &userService{
		userRepo: repositories.NewUserRepository(db),
		jwtUtil:  jwtUtil,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req dto.UserRegisterRequest) (*models.User, error) {
	existingUser, err := s.userRepo.FindByEmail(req.Email) // FindByEmail暂时不传context，保持现状
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("邮箱已被注册")
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Email:    req.Email,
		Password: s.hashPassword(req.Password),
		Name:     req.Name,
		Status:   "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserLoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email) // FindByEmail暂时不传context，保持现状
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	if !user.IsActive() {
		return nil, errors.New("用户已被禁用")
	}

	if !s.verifyPassword(req.Password, user.Password) {
		return nil, errors.New("密码错误")
	}

	accessToken, expiresIn, err := s.jwtUtil.GenerateToken(utils.UserInfo{
		Id:          user.ID,
		DisplayName: user.Name,
		Email:       user.Email,
		Avatar:      user.Avatar,
	})
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &dto.UserLoginResponse{
		User: dto.UserInfoResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Avatar:    user.Avatar,
			Status:    user.Status,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	claims, err := s.jwtUtil.ParseToken(refreshToken)
	if err != nil {
		return nil, errors.New("刷新令牌无效")
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserInfo.Id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	if user.Status != "active" {
		return nil, errors.New("用户状态异常")
	}

	accessToken, expiresIn, err := s.jwtUtil.GenerateToken(utils.UserInfo{
		Id:          user.ID,
		DisplayName: user.Name,
		Email:       user.Email,
		Avatar:      user.Avatar,
	})
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	return &dto.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

// Logout 用户登出
func (s *userService) Logout(ctx context.Context, accessToken string) error {
	_, err := s.jwtUtil.ParseToken(accessToken)
	if err != nil {
		return errors.New("访问令牌无效")
	}

	return nil
}

// GetUserByID 根据ID获取用户信息
func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// SendSMSCode 发送短信验证码
func (s *userService) SendSMSCode(ctx context.Context, req dto.SendSMSCodeRequest) (*dto.SendSMSCodeResponse, error) {
	// 生成6位随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 模拟发送短信（实际项目中这里调用短信服务商API）
	// 这里只是记录日志，表示"发送"了验证码
	fmt.Printf("模拟发送短信验证码到 %s: %s\n", req.Phone, code)

	// 将验证码存储到本地缓存中，有效期5分钟
	cacheKey := fmt.Sprintf("sms_code:%s", req.Phone)
	localCache := cache.Local()
	if localCache == nil {
		return nil, fmt.Errorf("本地缓存未初始化")
	}

	err := localCache.Put(ctx, cacheKey, code, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("存储验证码失败: %w", err)
	}

	expiredAt := time.Now().Add(5 * time.Minute).Unix()

	return &dto.SendSMSCodeResponse{
		Message:   "验证码发送成功",
		ExpiredAt: expiredAt,
	}, nil
}

// VerifySMSCode 验证短信验证码
func (s *userService) VerifySMSCode(ctx context.Context, req dto.VerifySMSCodeRequest) (*dto.VerifySMSCodeResponse, error) {
	cacheKey := fmt.Sprintf("sms_code:%s", req.Phone)

	// 从本地缓存中获取验证码
	localCache := cache.Local()
	if localCache == nil {
		return nil, fmt.Errorf("本地缓存未初始化")
	}

	storedCode, err := localCache.Get(ctx, cacheKey)
	if err != nil {
		return &dto.VerifySMSCodeResponse{
			Valid:   false,
			Message: "验证码已过期或不存在",
		}, nil
	}

	// 验证验证码是否正确
	if storedCode != req.Code {
		return &dto.VerifySMSCodeResponse{
			Valid:   false,
			Message: "验证码错误",
		}, nil
	}

	// 验证成功后删除验证码（防止重复使用）
	localCache.Forget(ctx, cacheKey)

	return &dto.VerifySMSCodeResponse{
		Valid:   true,
		Message: "验证码正确",
	}, nil
}

// SendEmailCode 发送邮箱验证码
func (s *userService) SendEmailCode(ctx context.Context, req dto.SendEmailCodeRequest) (*dto.SendEmailCodeResponse, error) {
	// 生成6位随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 模拟发送邮件（实际项目中这里调用邮件服务商API）
	// 这里只是记录日志，表示"发送"了验证码
	fmt.Printf("模拟发送邮箱验证码到 %s: %s\n", req.Email, code)

	// 将验证码存储到本地缓存中，有效期5分钟
	cacheKey := fmt.Sprintf("email_code:%s", req.Email)
	localCache := cache.Local()
	if localCache == nil {
		return nil, fmt.Errorf("本地缓存未初始化")
	}

	err := localCache.Put(ctx, cacheKey, code, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("存储验证码失败: %w", err)
	}

	expiredAt := time.Now().Add(5 * time.Minute).Unix()

	return &dto.SendEmailCodeResponse{
		Message:   "验证码发送成功",
		ExpiredAt: expiredAt,
	}, nil
}

// VerifyEmailCode 验证邮箱验证码
func (s *userService) VerifyEmailCode(ctx context.Context, req dto.VerifyEmailCodeRequest) (*dto.VerifyEmailCodeResponse, error) {
	cacheKey := fmt.Sprintf("email_code:%s", req.Email)

	// 从本地缓存中获取验证码
	localCache := cache.Local()
	if localCache == nil {
		return nil, fmt.Errorf("本地缓存未初始化")
	}

	storedCode, err := localCache.Get(ctx, cacheKey)
	if err != nil {
		return &dto.VerifyEmailCodeResponse{
			Valid:   false,
			Message: "验证码已过期或不存在",
		}, nil
	}

	// 验证验证码是否正确
	if storedCode != req.Code {
		return &dto.VerifyEmailCodeResponse{
			Valid:   false,
			Message: "验证码错误",
		}, nil
	}

	// 验证成功后删除验证码（防止重复使用）
	localCache.Forget(ctx, cacheKey)

	return &dto.VerifyEmailCodeResponse{
		Valid:   true,
		Message: "验证码正确",
	}, nil
}

// hashPassword 密码加密
func (s *userService) hashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

// verifyPassword 验证密码
func (s *userService) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
