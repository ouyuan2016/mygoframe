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

// UserInfo 定义了存储在 JWT 中的用户信息
type UserInfo struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

// Claims 定义了 JWT 的声明，包括自定义的用户信息和标准的注册声明
type Claims struct {
	*UserInfo
	jwt.RegisteredClaims
}

// Signer 定义了 JWT 签名和验证的接口
type Signer interface {
	Sign(claims Claims) (string, error)
	Verify(tokenString string) (*Claims, error)
	GetAlgorithm() string
}

// JWTUtil 是 JWT 操作的核心结构
type JWTUtil struct {
	config config.JWT
	signer Signer
}

// GetJWTUtil 使用单例模式获取 JWTUtil 实例
func GetJWTUtil() (*JWTUtil, error) {
	var initErr error
	jwtUtilOnce.Do(func() {
		cfg := config.GetConfig().JWT
		if cfg.SigningMethod == "" {
			cfg.SigningMethod = "HS256" // 默认签名算法
		}

		signer, err := newSigner(cfg)
		if err != nil {
			initErr = fmt.Errorf("创建签名器失败: %w", err)
			return
		}

		jwtUtil = &JWTUtil{
			config: cfg,
			signer: signer,
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

// newSigner 根据配置创建对应的 Signer
func newSigner(cfg config.JWT) (Signer, error) {
	method := strings.ToUpper(cfg.SigningMethod)
	switch method {
	case "RS256", "RS384", "RS512":
		return newRSASigner(method, cfg.PrivateKeyPath, cfg.PublicKeyPath)
	case "HS256", "HS384", "HS512":
		return newHMACSigner(method, cfg.SecretKey)
	default:
		return nil, fmt.Errorf("不支持的签名方法: %s", cfg.SigningMethod)
	}
}

// GenerateToken 生成访问令牌
func (j *JWTUtil) GenerateToken(userInfo UserInfo) (string, int, error) {
	if j.config.AccessTokenExpire <= 0 {
		return "", 0, errors.New("访问令牌过期时间必须大于0")
	}
	expiresIn := j.config.AccessTokenExpire * 60 // 分钟转秒
	expireTime := time.Now().Add(time.Duration(j.config.AccessTokenExpire) * time.Minute)

	token, err := j.generateToken(userInfo, expireTime)
	if err != nil {
		return "", 0, err
	}
	return token, expiresIn, nil
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTUtil) GenerateRefreshToken(userId string) (string, error) {
	if j.config.RefreshTokenExpire <= 0 {
		return "", errors.New("刷新令牌过期时间必须大于0")
	}
	expireTime := time.Now().Add(time.Duration(j.config.RefreshTokenExpire) * time.Minute)

	// 刷新令牌只需要用户ID
	userInfo := UserInfo{Id: userId}
	return j.generateToken(userInfo, expireTime)
}

// ParseToken 解析并验证令牌
func (j *JWTUtil) ParseToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("令牌字符串不能为空")
	}
	return j.signer.Verify(tokenString)
}

// generateToken 是内部通用的令牌生成函数
func (j *JWTUtil) generateToken(userInfo UserInfo, expireTime time.Time) (string, error) {
	claims := Claims{
		UserInfo: &userInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
		},
	}
	return j.signer.Sign(claims)
}

// --- HMAC Signer ---

type hmacSigner struct {
	algorithm jwt.SigningMethod
	secretKey []byte
}

func newHMACSigner(method, secretKey string) (Signer, error) {
	if secretKey == "" {
		return nil, errors.New("HMAC 算法需要配置 secretKey")
	}

	var alg jwt.SigningMethod
	switch method {
	case "HS256":
		alg = jwt.SigningMethodHS256
	case "HS384":
		alg = jwt.SigningMethodHS384
	case "HS512":
		alg = jwt.SigningMethodHS512
	default:
		// 这个分支在 newSigner 中已经处理，但为了稳健性保留
		return nil, fmt.Errorf("不支持的 HMAC 算法: %s", method)
	}

	return &hmacSigner{
		algorithm: alg,
		secretKey: []byte(secretKey),
	}, nil
}

func (s *hmacSigner) GetAlgorithm() string {
	return s.algorithm.Alg()
}

func (s *hmacSigner) Sign(claims Claims) (string, error) {
	token := jwt.NewWithClaims(s.algorithm, claims)
	return token.SignedString(s.secretKey)
}

func (s *hmacSigner) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("令牌无效")
}

// --- RSA Signer ---

type rsaSigner struct {
	algorithm  jwt.SigningMethod
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func newRSASigner(method, privateKeyPath, publicKeyPath string) (Signer, error) {
	var alg jwt.SigningMethod
	switch method {
	case "RS256":
		alg = jwt.SigningMethodRS256
	case "RS384":
		alg = jwt.SigningMethodRS384
	case "RS512":
		alg = jwt.SigningMethodRS512
	default:
		return nil, fmt.Errorf("不支持的 RSA 算法: %s", method)
	}

	if privateKeyPath == "" || publicKeyPath == "" {
		return nil, errors.New("RSA 算法需要配置私钥和公钥文件路径")
	}

	privateKey, err := parseRSAPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	publicKey, err := parseRSAPublicKey(publicKeyPath)
	if err != nil {
		return nil, err
	}

	if err := validateKeyPair(privateKey, publicKey); err != nil {
		return nil, fmt.Errorf("RSA 密钥对验证失败: %w", err)
	}

	return &rsaSigner{
		algorithm:  alg,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (s *rsaSigner) GetAlgorithm() string {
	return s.algorithm.Alg()
}

func (s *rsaSigner) Sign(claims Claims) (string, error) {
	token := jwt.NewWithClaims(s.algorithm, claims)
	return token.SignedString(s.privateKey)
}

func (s *rsaSigner) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("令牌无效")
}

func parseRSAPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败 [%s]: %w", filePath, err)
	}
	return jwt.ParseRSAPrivateKeyFromPEM(keyData)
}

func parseRSAPublicKey(filePath string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败 [%s]: %w", filePath, err)
	}
	return jwt.ParseRSAPublicKeyFromPEM(keyData)
}

func validateKeyPair(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) error {
	testData := []byte("key-pair-validation")
	hash := crypto.SHA256.New()
	hash.Write(testData)
	digest := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest)
	if err != nil {
		return fmt.Errorf("使用私钥签名失败: %w", err)
	}

	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest, signature); err != nil {
		return fmt.Errorf("使用公钥验证签名失败: %w", err)
	}

	return nil
}
