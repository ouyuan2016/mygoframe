package database

import (
	"fmt"

	"github.com/ouyuan2016/mygoframe/pkg/config"
	"gorm.io/gorm/logger"

	pkglogger "github.com/ouyuan2016/mygoframe/pkg/logger"
)

type Writer struct {
	config config.Database
	writer logger.Writer
}

func NewWriter(config config.Database) *Writer {
	return &Writer{config: config}
}

// Printf 格式化打印日志
func (c *Writer) Printf(message string, data ...any) {

	// 当有日志时候均需要输出到控制台
	fmt.Printf(message, data...)

	// 当开启了zap的情况，会打印到日志记录
	if c.config.LogZap {
		switch c.config.LevelLog() {
		case logger.Silent:
			pkglogger.Debug(fmt.Sprintf(message, data...))
		case logger.Error:
			pkglogger.Error(fmt.Sprintf(message, data...))
		case logger.Warn:
			pkglogger.Warn(fmt.Sprintf(message, data...))
		case logger.Info:
			pkglogger.Info(fmt.Sprintf(message, data...))
		default:
			pkglogger.Info(fmt.Sprintf(message, data...))
		}
		return
	}
}
