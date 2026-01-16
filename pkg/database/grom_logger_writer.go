package database

import (
	"fmt"
	"os"

	"mygoframe/pkg/config"
	"mygoframe/pkg/logger"
)

type Writer struct {
	cfg *config.Config
}

func (w *Writer) Printf(format string, args ...interface{}) {
	if w.cfg.MySQL.LogZap {
		logger.Info(fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func NewWriter(cfg *config.Config) *Writer {
	return &Writer{cfg: cfg}
}
