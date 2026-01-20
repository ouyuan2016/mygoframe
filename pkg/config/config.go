package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

type Config struct {
	System     System     `mapstructure:"system"`
	MySQL      Database   `mapstructure:"mysql"`
	Postgres   Database   `mapstructure:"pgsql"`
	SQLite     Database   `mapstructure:"sqlite"`
	Zap        Zap        `mapstructure:"zap"`
	JWT        JWT        `mapstructure:"jwt"`
	Redis      Redis      `mapstructure:"redis"`       // 新增 Redis 配置
	LocalCache LocalCache `mapstructure:"local-cache"` // 本地缓存配置
	Queue      Queue      `mapstructure:"queue"`       // 队列配置
}

type System struct {
	DbType             string `mapstructure:"db-type"`
	Addr               int    `mapstructure:"addr"`
	DisableAutoMigrate bool   `mapstructure:"disable-auto-migrate"`
}

type Database struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	Config       string `mapstructure:"config"`
	DbName       string `mapstructure:"db-name"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	LogMode      string `mapstructure:"log-mode"`
	MaxIdleConns int    `mapstructure:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns"`
	LogZap       bool   `mapstructure:"log-zap"`
}

func (d *Database) LevelLog() logger.LogLevel {
	switch strings.ToLower(d.LogMode) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	default:
		return logger.Info
	}
}

type Zap struct {
	Level           string `mapstructure:"level"`
	Format          string `mapstructure:"format"`
	OutputPath      string `mapstructure:"output-path"`
	MaxSize         int    `mapstructure:"max-size"`
	MaxBackups      int    `mapstructure:"max-backups"`
	MaxAge          int    `mapstructure:"max-age"`
	Compress        bool   `mapstructure:"compress"`
	RotateByDate    bool   `mapstructure:"rotate-by-date"`
	LogInConsole    bool   `mapstructure:"log-in-console"`
	LevelSeparation bool   `mapstructure:"level-separation"` // 按级别分离日志文件
}

type JWT struct {
	SecretKey          string `mapstructure:"secret-key"`
	SigningMethod      string `mapstructure:"signing-method"`
	AccessTokenExpire  int    `mapstructure:"access-token-expire"`
	RefreshTokenExpire int    `mapstructure:"refresh-token-expire"`
	Issuer             string `mapstructure:"issuer"`
	PrivateKeyPath     string `mapstructure:"private-key-path"`
	PublicKeyPath      string `mapstructure:"public-key-path"`
}

// Redis Redis配置
type Redis struct {
	Enabled      bool   `mapstructure:"enabled"` // 是否启用Redis
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool-size"`
	MinIdleConns int    `mapstructure:"min-idle-conns"`
	MaxRetries   int    `mapstructure:"max-retries"`
	DialTimeout  int    `mapstructure:"dial-timeout"`
	ReadTimeout  int    `mapstructure:"read-timeout"`
	WriteTimeout int    `mapstructure:"write-timeout"`
	Prefix       string `mapstructure:"prefix"` // Redis缓存前缀
}

// GetAddr 获取Redis地址
func (r *Redis) GetAddr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// LocalCache 本地缓存配置
type LocalCache struct {
	MaxCost int64 `mapstructure:"max-cost"` // 本地缓存最大容量(字节)
	MaxKeys int64 `mapstructure:"max-keys"` // 本地缓存最大key数量
}

// Queue 队列配置
type Queue struct {
	Enabled     bool           `mapstructure:"enabled"`     // 是否启用队列服务
	Concurrency int            `mapstructure:"concurrency"` // 并发工作线程数
	Queues      map[string]int `mapstructure:"queues"`      // 队列优先级配置
	MaxRetry    int            `mapstructure:"max-retry"`   // 最大重试次数
	Timeout     int            `mapstructure:"timeout"`     // 任务超时时间（秒）
	Retention   int            `mapstructure:"retention"`   // 任务保留时间（秒）
}

func GetBuildMode() string {
	mode := "dev"
	if envMode := os.Getenv("BUILD_MODE"); envMode != "" {
		mode = envMode
	}
	return mode
}

func GetConfig() *Config {
	mode := GetBuildMode()
	log.Printf("当前构建模式: %s", mode)

	var configName string
	switch mode {
	case "pro", "prod", "production":
		configName = "config.pro"
	case "test", "testing":
		configName = "config.test"
	default:
		configName = "config"
	}

	viper.SetConfigName(configName) // 配置文件名(不带扩展名)
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath(".")        // 当前目录查找
	viper.AddConfigPath("./config") // config目录查找

	if err := viper.ReadInConfig(); err != nil {
		if configName != "config" {
			log.Printf("读取配置文件 %s.yaml 失败: %v，尝试读取默认配置文件", configName, err)
			viper.SetConfigName("config")
			if err := viper.ReadInConfig(); err != nil {
				log.Printf("读取默认配置文件失败: %v", err)
			}
		} else {
			log.Printf("读取配置文件错误: %v", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("无法解析配置: %v", err)
	}

	return &config
}

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
