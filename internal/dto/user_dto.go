package dto

import "time"

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	User         UserInfoResponse `json:"user"`
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int              `json:"expires_in"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// UserRegisterResponse 用户注册响应
type UserRegisterResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// SendSMSCodeRequest 发送短信验证码请求
type SendSMSCodeRequest struct {
	Phone string `json:"phone" binding:"required,min=11,max=11"`
}

// SendSMSCodeResponse 发送短信验证码响应
type SendSMSCodeResponse struct {
	Message   string `json:"message"`
	ExpiredAt int64  `json:"expired_at"`
}

// VerifySMSCodeRequest 验证短信验证码请求
type VerifySMSCodeRequest struct {
	Phone string `json:"phone" binding:"required,min=11,max=11"`
	Code  string `json:"code" binding:"required,min=4,max=6"`
}

// VerifySMSCodeResponse 验证短信验证码响应
type VerifySMSCodeResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// SendEmailCodeRequest 发送邮箱验证码请求
type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// SendEmailCodeResponse 发送邮箱验证码响应
type SendEmailCodeResponse struct {
	Message   string `json:"message"`
	ExpiredAt int64  `json:"expired_at"`
}

// VerifyEmailCodeRequest 验证邮箱验证码请求
type VerifyEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,min=4,max=6"`
}

// VerifyEmailCodeResponse 验证邮箱验证码响应
type VerifyEmailCodeResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}
