package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

// Config 应用配置
type Config struct {
	System     System     `mapstructure:"system"`
	MySQL      Database   `mapstructure:"mysql"`
	Postgres   Database   `mapstructure:"pgsql"`
	SQLite     Database   `mapstructure:"sqlite"`
	Redis      Redis      `mapstructure:"redis"`
	Qiniu      Qiniu      `mapstructure:"qiniu"`
	TencentCOS TencentCOS `mapstructure:"tencent-cos"`
	AliyunOSS  AliyunOSS  `mapstructure:"aliyun-oss"`
	LocalOSS   LocalOSS   `mapstructure:"local-oss"`
	Zap        Zap        `mapstructure:"zap"`
	JWT        JWT        `mapstructure:"jwt"`
}

// System 系统配置
type System struct {
	DbType             string `mapstructure:"db-type"`
	OssType            string `mapstructure:"oss-type"`
	RouterPrefix       string `mapstructure:"router-prefix"`
	Addr               int    `mapstructure:"addr"`
	IplimitCount       int    `mapstructure:"iplimit-count"`
	IplimitTime        int    `mapstructure:"iplimit-time"`
	UseMultipoint      bool   `mapstructure:"use-multipoint"`
	UseRedis           bool   `mapstructure:"use-redis"`
	UseMongo           bool   `mapstructure:"use-mongo"`
	UseStrictAuth      bool   `mapstructure:"use-strict-auth"`
	DisableAutoMigrate bool   `mapstructure:"disable-auto-migrate"`
}

// Database 数据库配置
type Database struct {
	Prefix       string `mapstructure:"prefix"`
	Port         string `mapstructure:"port"`
	Config       string `mapstructure:"config"`
	DbName       string `mapstructure:"db-name"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Path         string `mapstructure:"path"`
	Engine       string `mapstructure:"engine"`
	LogMode      string `mapstructure:"log-mode"`
	MaxIdleConns int    `mapstructure:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns"`
	Singular     bool   `mapstructure:"singular"`
	LogZap       bool   `mapstructure:"log-zap"`
}

// LevelLog 返回GORM日志级别
func (d *Database) LevelLog() logger.LogLevel {
	switch strings.ToLower(d.LogMode) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

// Redis Redis配置
type Redis struct {
	Name         string   `mapstructure:"name"`
	Addr         string   `mapstructure:"addr"`
	Password     string   `mapstructure:"password"`
	DB           int      `mapstructure:"db"`
	UseCluster   bool     `mapstructure:"use-cluster"`
	ClusterAddrs []string `mapstructure:"cluster-addrs"`
}

// Qiniu 七牛云配置
type Qiniu struct {
	AccessKey  string `mapstructure:"access-key"`
	SecretKey  string `mapstructure:"secret-key"`
	Bucket     string `mapstructure:"bucket"`
	ImgPath    string `mapstructure:"img-path"`
	Zone       string `mapstructure:"zone"`
	UseHTTPS   bool   `mapstructure:"use-https"`
	UseCdnUrls bool   `mapstructure:"use-cdn-urls"`
}

// TencentCOS 腾讯云COS配置
type TencentCOS struct {
	SecretID   string `mapstructure:"secret-id"`
	SecretKey  string `mapstructure:"secret-key"`
	Region     string `mapstructure:"region"`
	Bucket     string `mapstructure:"bucket"`
	PathPrefix string `mapstructure:"path-prefix"`
	Domain     string `mapstructure:"domain"`
	IsPrivate  bool   `mapstructure:"is-private"`
	IsDefault  bool   `mapstructure:"is-default"`
}

// AliyunOSS 阿里云OSS配置
type AliyunOSS struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyId     string `mapstructure:"access-key-id"`
	AccessKeySecret string `mapstructure:"access-key-secret"`
	BucketName      string `mapstructure:"bucket-name"`
	BucketDomain    string `mapstructure:"bucket-domain"`
	IsPrivate       bool   `mapstructure:"is-private"`
	IsDefault       bool   `mapstructure:"is-default"`
}

// LocalOSS 本地存储配置
type LocalOSS struct {
	Path      string `mapstructure:"path"`
	StorePath string `mapstructure:"store-path"`
}

// Zap 日志配置
type Zap struct {
	Level         string `mapstructure:"level"`
	Prefix        string `mapstructure:"prefix"`
	Format        string `mapstructure:"format"`
	Director      string `mapstructure:"director"`
	EncodeLevel   string `mapstructure:"encode-level"`
	StacktraceKey string `mapstructure:"stacktrace-key"`
	ShowLine      bool   `mapstructure:"show-line"`
	LogInConsole  bool   `mapstructure:"log-in-console"`
	RetentionDay  int    `mapstructure:"retention-day"`
	OutputPath    string `mapstructure:"output-path"`
	MaxSize       int    `mapstructure:"max-size"`
	MaxBackups    int    `mapstructure:"max-backups"`
	MaxAge        int    `mapstructure:"max-age"`
	Compress      bool   `mapstructure:"compress"`
	RotateByDate  bool   `mapstructure:"rotate-by-date"`
}

// JWT JWT配置
type JWT struct {
	SecretKey          string `mapstructure:"secret-key"`
	SigningMethod      string `mapstructure:"signing-method"`
	AccessTokenExpire  int    `mapstructure:"access-token-expire"`
	RefreshTokenExpire int    `mapstructure:"refresh-token-expire"`
	Issuer             string `mapstructure:"issuer"`
	PrivateKey         string `mapstructure:"private-key"`
	PublicKey          string `mapstructure:"public-key"`
	PrivateKeyPath     string `mapstructure:"private-key-path"`
	PublicKeyPath      string `mapstructure:"public-key-path"`
}

// GetBuildMode 获取构建模式，优先从环境变量读取，默认为dev模式
func GetBuildMode() string {
	// 默认为dev模式
	mode := "dev"

	// 检查环境变量
	if envMode := os.Getenv("BUILD_MODE"); envMode != "" {
		mode = envMode
	}

	return mode
}
func GetConfig() *Config {
	mode := GetBuildMode()
	log.Printf("当前构建模式: %s", mode)

	// 设置配置文件名
	var configName string
	switch mode {
	case "pro", "prod", "production":
		configName = "config.pro"
	case "test", "testing":
		configName = "config.test"
	case "dev", "development":
		configName = "config.dev"
	default:
		configName = "config"
	}

	viper.SetConfigName(configName) // 配置文件名(不带扩展名)
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath(".")        // 当前目录查找
	viper.AddConfigPath("./config") // config目录查找

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果特定模式的配置文件不存在，尝试读取默认配置文件
		if configName != "config" {
			log.Printf("读取配置文件 %s.yaml 失败: %v，尝试读取默认配置文件", configName, err)
			viper.SetConfigName("config")
			if err := viper.ReadInConfig(); err != nil {
				log.Printf("读取默认配置文件失败: %v", err)
			} else {
				log.Printf("成功读取默认配置文件")
			}
		} else {
			log.Printf("读取配置文件错误: %v", err)
		}
	} else {
		log.Printf("成功读取配置文件: %s.yaml", configName)
	}

	// 将配置映射到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("无法解析配置: %v", err)
	}

	return &config
}

// GetDatabaseConfig 获取当前使用的数据库配置
func (c *Config) GetDatabaseConfig() Database {
	switch c.System.DbType {
	case "mysql":
		return c.MySQL
	case "postgres", "pgsql":
		return c.Postgres
	case "sqlite":
		return c.SQLite
	default:
		return c.MySQL
	}
}

// GetRedisConfig 获取Redis配置
func (c *Config) GetRedisConfig() Redis {
	return c.Redis
}
