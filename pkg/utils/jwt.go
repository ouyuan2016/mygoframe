package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ouyuan2016/mygoframe/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtUtil     *JWTUtil
	jwtUtilOnce sync.Once
)

// UserInfo 用户信息结构体
type UserInfo struct {
	Id          string `xorm:"varchar(100) index" json:"id"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Avatar      string `xorm:"varchar(500)" json:"avatar"`
	Email       string `xorm:"varchar(100) index" json:"email"`
	Phone       string `xorm:"varchar(100) index" json:"phone"`
}

// Claims JWT声明结构体
type Claims struct {
	*UserInfo
	jwt.RegisteredClaims
}

// JWTUtil JWT工具类
type JWTUtil struct {
	config     config.JWT
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// GetJWTUtil 获取JWT工具实例（单例模式）
func GetJWTUtil() (*JWTUtil, error) {
	var initErr error
	jwtUtilOnce.Do(func() {
		// 获取全局配置
		cfg := config.GetConfig().JWT

		jwtUtil = &JWTUtil{
			config: cfg,
		}

		// 如果没有指定签名方法，默认为HS256
		if cfg.SigningMethod == "" {
			cfg.SigningMethod = "HS256"
		}

		// 根据签名方法初始化密钥
		switch strings.ToUpper(cfg.SigningMethod) {
		case "RS256", "RS384", "RS512":
			// 解析RSA密钥
			if err := jwtUtil.parseRSAKeys(); err != nil {
				initErr = err
				jwtUtil = nil
				return
			}
		case "HS256", "HS384", "HS512":
			// HMAC算法不需要额外处理
			if cfg.SecretKey == "" {
				initErr = errors.New("secret key is required for HMAC algorithms")
				jwtUtil = nil
				return
			}
		default:
			initErr = errors.New("unsupported signing method: " + cfg.SigningMethod)
			jwtUtil = nil
			return
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	if jwtUtil == nil {
		return nil, errors.New("JWT工具初始化失败")
	}

	return jwtUtil, nil
}

// GenerateToken 生成JWT令牌
func (j *JWTUtil) GenerateToken(userInfo UserInfo) (string, error) {
	// 验证输入参数
	if userInfo.Id == "" {
		return "", errors.New("用户ID不能为空")
	}
	if userInfo.DisplayName == "" {
		return "", errors.New("显示名称不能为空")
	}
	if userInfo.Email == "" {
		return "", errors.New("邮箱不能为空")
	}

	// 检查过期时间配置
	if j.config.AccessTokenExpire <= 0 {
		return "", errors.New("访问令牌过期时间必须大于0")
	}

	// 设置过期时间（使用AccessTokenExpire，单位为秒）
	expireTime := time.Now().Add(time.Duration(j.config.AccessTokenExpire) * time.Second)

	// 创建声明
	claims := Claims{
		UserInfo: &userInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
		},
	}

	// 根据配置选择签名方法
	var signingMethod jwt.SigningMethod
	switch strings.ToUpper(j.config.SigningMethod) {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	case "RS384":
		signingMethod = jwt.SigningMethodRS384
	case "RS512":
		signingMethod = jwt.SigningMethodRS512
	default:
		return "", errors.New("不支持的签名方法: " + j.config.SigningMethod)
	}

	// 创建令牌
	token := jwt.NewWithClaims(signingMethod, claims)

	// 根据算法类型进行签名
	var tokenString string
	var err error

	switch strings.ToUpper(j.config.SigningMethod) {
	case "RS256", "RS384", "RS512":
		if j.privateKey == nil {
			return "", errors.New("RSA私钥未初始化")
		}
		tokenString, err = token.SignedString(j.privateKey)
	default:
		if j.config.SecretKey == "" {
			return "", errors.New("HMAC密钥未配置")
		}
		tokenString, err = token.SignedString([]byte(j.config.SecretKey))
	}

	if err != nil {
		return "", fmt.Errorf("生成令牌失败: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func (j *JWTUtil) ParseToken(tokenString string) (*Claims, error) {
	// 验证输入参数
	if tokenString == "" {
		return nil, errors.New("令牌字符串不能为空")
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		switch strings.ToUpper(j.config.SigningMethod) {
		case "RS256", "RS384", "RS512":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				alg := "unknown"
				if token.Header["alg"] != nil {
					alg = token.Header["alg"].(string)
				}
				return nil, fmt.Errorf("意外的签名方法: %s (期望: %s)", alg, j.config.SigningMethod)
			}
			if j.publicKey == nil {
				return nil, errors.New("RSA公钥未初始化")
			}
			return j.publicKey, nil
		case "HS256", "HS384", "HS512":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				alg := "unknown"
				if token.Header["alg"] != nil {
					alg = token.Header["alg"].(string)
				}
				return nil, fmt.Errorf("意外的签名方法: %s (期望: %s)", alg, j.config.SigningMethod)
			}
			if j.config.SecretKey == "" {
				return nil, errors.New("HMAC密钥未配置")
			}
			return []byte(j.config.SecretKey), nil
		default:
			return nil, fmt.Errorf("不支持的签名方法: %s", j.config.SigningMethod)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	// 获取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("获取声明失败")
	}

	return claims, nil
}

// ValidateToken 验证JWT令牌
func (j *JWTUtil) ValidateToken(tokenString string) error {
	_, err := j.ParseToken(tokenString)
	return err
}

// RefreshToken 刷新JWT令牌
func (j *JWTUtil) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 生成新的令牌
	userInfo := UserInfo{
		Id:          claims.UserInfo.Id,
		DisplayName: claims.UserInfo.DisplayName,
		Avatar:      claims.UserInfo.Avatar,
		Email:       claims.UserInfo.Email,
		Phone:       claims.UserInfo.Phone,
	}
	return j.GenerateToken(userInfo)
}

// GetClaimsFromToken 从令牌字符串中获取声明（不验证）
func (j *JWTUtil) GetClaimsFromToken(tokenString string) (*Claims, error) {
	// 解析但不验证签名
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, errors.New("解析令牌失败: " + err.Error())
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("获取声明失败")
	}

	return claims, nil
}

// IsTokenExpired 检查令牌是否过期
func (j *JWTUtil) IsTokenExpired(tokenString string) bool {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return true
	}

	return claims.ExpiresAt.Time.Before(time.Now())
}

// GetTokenInfo 获取令牌详细信息
func (j *JWTUtil) GetTokenInfo(tokenString string) (map[string]interface{}, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"id":          claims.UserInfo.Id,
		"displayName": claims.UserInfo.DisplayName,
		"avatar":      claims.UserInfo.Avatar,
		"email":       claims.UserInfo.Email,
		"phone":       claims.UserInfo.Phone,
		"issuer":      claims.Issuer,
		"issued_at":   claims.IssuedAt.Time.Format("2006-01-02 15:04:05"),
		"expires_at":  claims.ExpiresAt.Time.Format("2006-01-02 15:04:05"),
		"not_before":  claims.NotBefore.Time.Format("2006-01-02 15:04:05"),
		"is_expired":  claims.ExpiresAt.Time.Before(time.Now()),
	}

	return info, nil
}

// ValidateConfig 验证JWT配置的有效性
func ValidateConfig(cfg config.JWT) error {
	if cfg.SigningMethod == "" {
		return errors.New("签名方法不能为空")
	}

	switch strings.ToUpper(cfg.SigningMethod) {
	case "RS256", "RS384", "RS512":
		// RSA算法验证
		hasPrivateKey := cfg.PrivateKeyPath != "" || cfg.PrivateKey != ""
		hasPublicKey := cfg.PublicKeyPath != "" || cfg.PublicKey != ""

		if !hasPrivateKey && !hasPublicKey {
			return errors.New("RSA算法需要配置私钥和公钥")
		}

		if cfg.PrivateKeyPath != "" && cfg.PrivateKey != "" {
			return errors.New("不能同时配置私钥文件路径和私钥内容")
		}

		if cfg.PublicKeyPath != "" && cfg.PublicKey != "" {
			return errors.New("不能同时配置公钥文件路径和公钥内容")
		}

		// 检查文件路径
		if cfg.PrivateKeyPath != "" {
			if _, err := os.Stat(cfg.PrivateKeyPath); os.IsNotExist(err) {
				return fmt.Errorf("私钥文件不存在: %s", cfg.PrivateKeyPath)
			}
		}

		if cfg.PublicKeyPath != "" {
			if _, err := os.Stat(cfg.PublicKeyPath); os.IsNotExist(err) {
				return fmt.Errorf("公钥文件不存在: %s", cfg.PublicKeyPath)
			}
		}

	case "HS256", "HS384", "HS512":
		// HMAC算法验证
		if cfg.SecretKey == "" {
			return errors.New("HMAC算法需要配置密钥")
		}

	default:
		return fmt.Errorf("不支持的签名方法: %s", cfg.SigningMethod)
	}

	// 验证过期时间
	if cfg.AccessTokenExpire <= 0 {
		return errors.New("访问令牌过期时间必须大于0")
	}

	if cfg.RefreshTokenExpire <= 0 {
		return errors.New("刷新令牌过期时间必须大于0")
	}

	if cfg.AccessTokenExpire >= cfg.RefreshTokenExpire {
		return errors.New("访问令牌过期时间必须小于刷新令牌过期时间")
	}

	return nil
}

// parseRSAKeys 解析RSA密钥
func (j *JWTUtil) parseRSAKeys() error {
	// 验证配置完整性
	if err := j.validateRSAConfig(); err != nil {
		return err
	}

	// 解析私钥
	if err := j.parsePrivateKey(); err != nil {
		return err
	}

	// 解析公钥
	if err := j.parsePublicKey(); err != nil {
		return err
	}

	// 验证密钥对匹配
	if j.privateKey != nil && j.publicKey != nil {
		if err := j.validateKeyPair(); err != nil {
			return err
		}
	}

	return nil
}

// validateRSAConfig 验证RSA配置
func (j *JWTUtil) validateRSAConfig() error {
	// 检查是否至少提供了一种密钥配置方式
	hasPrivateKey := j.config.PrivateKeyPath != "" || j.config.PrivateKey != ""
	hasPublicKey := j.config.PublicKeyPath != "" || j.config.PublicKey != ""

	if !hasPrivateKey && !hasPublicKey {
		return errors.New("RSA算法需要配置密钥，请提供私钥和公钥")
	}

	// 检查私钥配置
	if j.config.PrivateKeyPath != "" && j.config.PrivateKey != "" {
		return errors.New("不能同时配置私钥文件路径和私钥内容，请选择其中一种方式")
	}

	// 检查公钥配置
	if j.config.PublicKeyPath != "" && j.config.PublicKey != "" {
		return errors.New("不能同时配置公钥文件路径和公钥内容，请选择其中一种方式")
	}

	// 检查文件路径是否存在
	if j.config.PrivateKeyPath != "" {
		if _, err := os.Stat(j.config.PrivateKeyPath); os.IsNotExist(err) {
			return errors.New("私钥文件不存在: " + j.config.PrivateKeyPath)
		}
	}

	if j.config.PublicKeyPath != "" {
		if _, err := os.Stat(j.config.PublicKeyPath); os.IsNotExist(err) {
			return errors.New("公钥文件不存在: " + j.config.PublicKeyPath)
		}
	}

	return nil
}

// parsePrivateKey 解析私钥
func (j *JWTUtil) parsePrivateKey() error {
	var privateKeyData []byte
	var err error

	// 优先使用文件路径
	if j.config.PrivateKeyPath != "" {
		privateKeyData, err = os.ReadFile(j.config.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("读取私钥文件失败 [%s]: %w", j.config.PrivateKeyPath, err)
		}
	} else if j.config.PrivateKey != "" {
		privateKeyData = []byte(j.config.PrivateKey)
	} else {
		return nil // 未配置私钥，跳过解析
	}

	// 解析PEM格式的私钥
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	j.privateKey = privateKey
	return nil
}

// parsePublicKey 解析公钥
func (j *JWTUtil) parsePublicKey() error {
	var publicKeyData []byte
	var err error

	// 优先使用文件路径
	if j.config.PublicKeyPath != "" {
		publicKeyData, err = os.ReadFile(j.config.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("读取公钥文件失败 [%s]: %w", j.config.PublicKeyPath, err)
		}
	} else if j.config.PublicKey != "" {
		publicKeyData = []byte(j.config.PublicKey)
	} else {
		return nil // 未配置公钥，跳过解析
	}

	// 解析PEM格式的公钥
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return fmt.Errorf("解析公钥失败: %w", err)
	}

	j.publicKey = publicKey
	return nil
}

// validateKeyPair 验证密钥对是否匹配
func (j *JWTUtil) validateKeyPair() error {
	if j.privateKey == nil || j.publicKey == nil {
		return nil // 如果其中一个为nil，跳过验证
	}

	// 使用私钥签名一个测试数据
	testData := []byte("test-key-pair-validation")
	hash := crypto.SHA256.New()
	hash.Write(testData)
	digest := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, j.privateKey, crypto.SHA256, digest)
	if err != nil {
		return fmt.Errorf("密钥对验证失败 - 签名错误: %w", err)
	}

	// 使用公钥验证签名
	err = rsa.VerifyPKCS1v15(j.publicKey, crypto.SHA256, digest, signature)
	if err != nil {
		return fmt.Errorf("密钥对验证失败 - 公钥私钥不匹配: %w", err)
	}

	return nil
}

// GenerateTokenWithExpire 生成指定过期时间的JWT令牌
func (j *JWTUtil) GenerateTokenWithExpire(id, displayName, email, phone, avatar string, expireTime time.Time) (string, error) {
	userInfo := UserInfo{
		Id:          id,
		DisplayName: displayName,
		Email:       email,
		Phone:       phone,
		Avatar:      avatar,
	}
	claims := Claims{
		UserInfo: &userInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
		},
	}

	// 根据算法类型选择签名方法
	var signingMethod jwt.SigningMethod
	switch strings.ToUpper(j.config.SigningMethod) {
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	case "RS384":
		signingMethod = jwt.SigningMethodRS384
	case "RS512":
		signingMethod = jwt.SigningMethodRS512
	default:
		signingMethod = jwt.SigningMethodHS256
	}

	token := jwt.NewWithClaims(signingMethod, claims)

	// 根据算法类型进行签名
	var tokenString string
	var err error

	switch strings.ToUpper(j.config.SigningMethod) {
	case "RS256", "RS384", "RS512":
		if j.privateKey == nil {
			return "", errors.New("RSA私钥未配置")
		}
		tokenString, err = token.SignedString(j.privateKey)
	default:
		tokenString, err = token.SignedString([]byte(j.config.SecretKey))
	}

	if err != nil {
		return "", errors.New("生成令牌失败: " + err.Error())
	}

	return tokenString, nil
}
