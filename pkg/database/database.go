package database

import (
	"fmt"
	"log"
	"time"

	"github.com/ouyuan2016/mygoframe/internal/models"
	"github.com/ouyuan2016/mygoframe/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	dbConfig := cfg.GetDatabaseConfig()
	var dsn string
	var err error

	switch cfg.System.DbType {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			dbConfig.Username, dbConfig.Password, dbConfig.Path, dbConfig.Port, dbConfig.DbName, dbConfig.Config)
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.New(NewWriter(dbConfig), logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      dbConfig.LevelLog(),
				Colorful:      true,
			}),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   dbConfig.Prefix,
				SingularTable: dbConfig.Singular,
			},
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	case "postgres", "pgsql":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			dbConfig.Path, dbConfig.Username, dbConfig.Password, dbConfig.DbName, dbConfig.Port)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.New(NewWriter(dbConfig), logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      dbConfig.LevelLog(),
				Colorful:      true,
			}),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   dbConfig.Prefix,
				SingularTable: dbConfig.Singular,
			},
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	case "sqlite":
		dsn = dbConfig.Path
		if dsn == "" {
			dsn = dbConfig.DbName
		}
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.New(NewWriter(dbConfig), logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      dbConfig.LevelLog(),
				Colorful:      true,
			}),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   dbConfig.Prefix,
				SingularTable: dbConfig.Singular,
			},
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.System.DbType)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 设置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	// 自动迁移数据库
	if !cfg.System.DisableAutoMigrate {
		log.Println("开始自动迁移数据库...")
		if err := DB.AutoMigrate(
			&models.News{},
		); err != nil {
			return fmt.Errorf("failed to auto migrate database: %w", err)
		}
		log.Println("数据库自动迁移完成")
	}

	log.Printf("数据库连接成功, 类型: %s", cfg.System.DbType)
	return nil
}
