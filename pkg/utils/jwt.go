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

	"mygoframe/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtUtil     *JWTUtil
	jwtUtilOnce sync.Once
)

type UserInfo struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

type Claims struct {
	*UserInfo
	jwt.RegisteredClaims
}

type JWTUtil struct {
	config     config.JWT
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func GetJWTUtil() (*JWTUtil, error) {
	var initErr error
	jwtUtilOnce.Do(func() {
		cfg := config.GetConfig().JWT

		jwtUtil = &JWTUtil{
			config: cfg,
		}

		if cfg.SigningMethod == "" {
			cfg.SigningMethod = "HS256"
		}

		switch strings.ToUpper(cfg.SigningMethod) {
		case "RS256", "RS384", "RS512":
			if err := jwtUtil.parseRSAKeys(); err != nil {
				initErr = err
				jwtUtil = nil
				return
			}
		case "HS256", "HS384", "HS512":
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

// generateTokenInternal 内部通用的令牌生成方法
func (j *JWTUtil) generateTokenInternal(claims Claims, tokenType string) (string, error) {
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

	token := jwt.NewWithClaims(signingMethod, claims)

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
		return "", fmt.Errorf("生成%s失败: %w", tokenType, err)
	}

	return tokenString, nil
}

// GenerateToken 生成访问令牌，返回令牌和过期时间（秒）
func (j *JWTUtil) GenerateToken(userInfo UserInfo) (string, int, error) {
	if j.config.AccessTokenExpire <= 0 {
		return "", 0, errors.New("访问令牌过期时间必须大于0")
	}

	// 将分钟转换为秒
	expiresIn := j.config.AccessTokenExpire * 60 // 转换为秒
	expireTime := time.Now().Add(time.Duration(j.config.AccessTokenExpire) * time.Minute)

	tokenString, err := j.GenerateTokenWithExpire(
		userInfo.Id,
		userInfo.DisplayName,
		userInfo.Email,
		userInfo.Phone,
		userInfo.Avatar,
		expireTime,
	)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// GenerateRefreshToken 生成刷新令牌（使用配置的过期时间，只需要用户ID）
func (j *JWTUtil) GenerateRefreshToken(userId string) (string, error) {
	if j.config.RefreshTokenExpire <= 0 {
		return "", errors.New("刷新令牌过期时间必须大于0")
	}
	// 将分钟转换为秒
	expireTime := time.Now().Add(time.Duration(j.config.RefreshTokenExpire) * time.Minute)

	// 刷新令牌只需要用户ID，其他字段留空
	return j.GenerateTokenWithExpire(userId, "", "", "", "", expireTime)
}

// getKeyForParseToken 获取用于解析令牌的密钥
func (j *JWTUtil) getKeyForParseToken(token *jwt.Token) (interface{}, error) {
	actualAlg := "unknown"
	if token.Header["alg"] != nil {
		actualAlg = token.Header["alg"].(string)
	}

	switch strings.ToUpper(j.config.SigningMethod) {
	case "RS256", "RS384", "RS512":
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("意外的签名方法: %s (期望: %s)", actualAlg, j.config.SigningMethod)
		}
		if j.publicKey == nil {
			return nil, errors.New("RSA公钥未初始化")
		}
		return j.publicKey, nil
	case "HS256", "HS384", "HS512":
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %s (期望: %s)", actualAlg, j.config.SigningMethod)
		}
		if j.config.SecretKey == "" {
			return nil, errors.New("HMAC密钥未配置")
		}
		return []byte(j.config.SecretKey), nil
	default:
		return nil, fmt.Errorf("不支持的签名方法: %s", j.config.SigningMethod)
	}
}

func (j *JWTUtil) ParseToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("令牌字符串不能为空")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, j.getKeyForParseToken)

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("无法解析令牌声明")
	}

	return claims, nil
}

func (j *JWTUtil) parseRSAKeys() error {
	if err := j.validateRSAConfig(); err != nil {
		return err
	}

	if err := j.parsePrivateKey(); err != nil {
		return err
	}

	if err := j.parsePublicKey(); err != nil {
		return err
	}

	if j.privateKey != nil && j.publicKey != nil {
		if err := j.validateKeyPair(); err != nil {
			return err
		}
	}

	return nil
}

// validateFileExists 验证文件是否存在
func (j *JWTUtil) validateFileExists(filePath, fileType string) error {
	if filePath == "" {
		return nil
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("%s文件不存在: %s", fileType, filePath)
	}
	return nil
}

func (j *JWTUtil) validateRSAConfig() error {
	if j.config.PrivateKeyPath == "" && j.config.PublicKeyPath == "" {
		return errors.New("RSA算法需要配置密钥，请提供私钥和公钥文件路径")
	}

	if err := j.validateFileExists(j.config.PrivateKeyPath, "私钥"); err != nil {
		return err
	}

	if err := j.validateFileExists(j.config.PublicKeyPath, "公钥"); err != nil {
		return err
	}

	return nil
}

// parseRSAKeyFromFile 从文件解析RSA密钥
func (j *JWTUtil) parseRSAKeyFromFile(filePath string, isPrivateKey bool) (interface{}, error) {
	if filePath == "" {
		return nil, nil
	}

	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取%s文件失败 [%s]: %w", map[bool]string{true: "私钥", false: "公钥"}[isPrivateKey], filePath, err)
	}

	if isPrivateKey {
		return jwt.ParseRSAPrivateKeyFromPEM(keyData)
	}
	return jwt.ParseRSAPublicKeyFromPEM(keyData)
}

func (j *JWTUtil) parsePrivateKey() error {
	privateKey, err := j.parseRSAKeyFromFile(j.config.PrivateKeyPath, true)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}
	if privateKey != nil {
		j.privateKey = privateKey.(*rsa.PrivateKey)
	}
	return nil
}

func (j *JWTUtil) parsePublicKey() error {
	publicKey, err := j.parseRSAKeyFromFile(j.config.PublicKeyPath, false)
	if err != nil {
		return fmt.Errorf("解析公钥失败: %w", err)
	}
	if publicKey != nil {
		j.publicKey = publicKey.(*rsa.PublicKey)
	}
	return nil
}

// validateKeyPair 验证RSA密钥对是否匹配
func (j *JWTUtil) validateKeyPair() error {
	if j.privateKey == nil || j.publicKey == nil {
		return nil
	}

	testData := []byte("test-key-pair-validation")
	hash := crypto.SHA256.New()
	hash.Write(testData)
	digest := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, j.privateKey, crypto.SHA256, digest)
	if err != nil {
		return fmt.Errorf("密钥对验证失败: %w", err)
	}

	if err := rsa.VerifyPKCS1v15(j.publicKey, crypto.SHA256, digest, signature); err != nil {
		return fmt.Errorf("密钥对验证失败: %w", err)
	}

	return nil
}

// GenerateTokenWithExpire 生成指定过期时间的令牌（使用完整用户信息）
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

	tokenString, err := j.generateTokenInternal(claims, "令牌")
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
