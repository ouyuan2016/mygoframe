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
	Id          string `xorm:"varchar(100) index" json:"id"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Avatar      string `xorm:"varchar(500)" json:"avatar"`
	Email       string `xorm:"varchar(100) index" json:"email"`
	Phone       string `xorm:"varchar(100) index" json:"phone"`
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

func (j *JWTUtil) GenerateToken(userInfo UserInfo) (string, error) {
	if userInfo.Id == "" {
		return "", errors.New("用户ID不能为空")
	}
	if userInfo.DisplayName == "" {
		return "", errors.New("显示名称不能为空")
	}
	if userInfo.Email == "" {
		return "", errors.New("邮箱不能为空")
	}

	if j.config.AccessTokenExpire <= 0 {
		return "", errors.New("访问令牌过期时间必须大于0")
	}

	expireTime := time.Now().Add(time.Duration(j.config.AccessTokenExpire) * time.Second)

	claims := Claims{
		UserInfo: &userInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
		},
	}

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
		return "", fmt.Errorf("生成令牌失败: %w", err)
	}

	return tokenString, nil
}

func (j *JWTUtil) ParseToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("令牌字符串不能为空")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
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

func (j *JWTUtil) validateRSAConfig() error {
	if j.config.PrivateKeyPath == "" && j.config.PublicKeyPath == "" {
		return errors.New("RSA算法需要配置密钥，请提供私钥和公钥文件路径")
	}

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

func (j *JWTUtil) parsePrivateKey() error {
	if j.config.PrivateKeyPath == "" {
		return nil
	}

	privateKeyData, err := os.ReadFile(j.config.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("读取私钥文件失败 [%s]: %w", j.config.PrivateKeyPath, err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	j.privateKey = privateKey
	return nil
}

func (j *JWTUtil) parsePublicKey() error {
	if j.config.PublicKeyPath == "" {
		return nil
	}

	publicKeyData, err := os.ReadFile(j.config.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("读取公钥文件失败 [%s]: %w", j.config.PublicKeyPath, err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return fmt.Errorf("解析公钥失败: %w", err)
	}

	j.publicKey = publicKey
	return nil
}

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
		return fmt.Errorf("密钥对验证失败 - 签名错误: %w", err)
	}

	err = rsa.VerifyPKCS1v15(j.publicKey, crypto.SHA256, digest, signature)
	if err != nil {
		return fmt.Errorf("密钥对验证失败 - 公钥私钥不匹配: %w", err)
	}

	return nil
}

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
