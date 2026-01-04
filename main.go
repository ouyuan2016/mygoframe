package main

import (
	"flag"
	"log"
	"os"

	"github.com/ouyuan2016/mygoframe/cmd/server"
)

func main() {
	// 定义命令行参数
	var buildMode string
	flag.StringVar(&buildMode, "mod", "dev", "构建模式 (dev/pro)")
	flag.Parse()

	// 如果指定了构建模式，设置环境变量
	if buildMode != "" {
		os.Setenv("BUILD_MODE", buildMode)
	}

	if err := server.Run(); err != nil {
		log.Printf("Error running server: %v", err)
		os.Exit(1)
	}
}
