package database

import (
	"fmt"
	"time"

	"mygoframe/internal/models"
	"mygoframe/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.System.DbType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			cfg.MySQL.Username,
			cfg.MySQL.Password,
			cfg.MySQL.Host,
			cfg.MySQL.Port,
			cfg.MySQL.DbName,
			cfg.MySQL.Config,
		)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.Postgres.Host,
			cfg.Postgres.Username,
			cfg.Postgres.Password,
			cfg.Postgres.DbName,
			cfg.Postgres.Port,
		)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.SQLite.Host)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.System.DbType)
	}

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	if cfg.System.DbType != "sqlite" {
		gormConfig.Logger = gormlogger.New(
			&Writer{cfg: cfg},
			gormlogger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  getLogLevel(cfg),
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)

	if !cfg.System.DisableAutoMigrate {
		if err := db.AutoMigrate(&models.News{}); err != nil {
			return nil, fmt.Errorf("数据库迁移失败: %v", err)
		}
	}

	return db, nil
}

func getLogLevel(cfg *config.Config) gormlogger.LogLevel {
	switch cfg.MySQL.LogMode {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn":
		return gormlogger.Warn
	default:
		return gormlogger.Info
	}
}
