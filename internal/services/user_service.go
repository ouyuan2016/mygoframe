package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mygoframe/internal/dto"
	"mygoframe/internal/models"
	"mygoframe/internal/repositories"
	"mygoframe/pkg/config"
	"mygoframe/pkg/utils"

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
}

// userService 用户服务实现
type userService struct {
	userRepo repositories.UserRepository
	jwtUtil  *utils.JWTUtil
	config   *config.Config
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB) UserService {
	jwtUtil, _ := utils.GetJWTUtil()
	cfg := config.GetConfig()
	return &userService{
		userRepo: repositories.NewUserRepository(db),
		jwtUtil:  jwtUtil,
		config:   cfg,
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

	if user.Status != "active" {
		return nil, errors.New("用户状态异常")
	}

	if !s.verifyPassword(req.Password, user.Password) {
		return nil, errors.New("密码错误")
	}

	accessToken, err := s.jwtUtil.GenerateToken(utils.UserInfo{
		Id:          user.ID,
		DisplayName: user.Name,
		Email:       user.Email,
		Avatar:      user.Avatar,
	})
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := s.jwtUtil.GenerateTokenWithExpire(
		user.ID,
		user.Name,
		user.Email,
		"",
		user.Avatar,
		time.Now().Add(7*24*time.Hour),
	)
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
		ExpiresIn:    s.config.JWT.AccessTokenExpire * 60,
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

	accessToken, err := s.jwtUtil.GenerateToken(utils.UserInfo{
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
		ExpiresIn:   s.config.JWT.AccessTokenExpire,
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
