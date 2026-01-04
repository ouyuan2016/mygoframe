## 项目结构

```
starlive/
├── cmd/
│   └── server/           # 应用程序入口
│       └── main.go
├── internal/
│   ├── handlers/         # HTTP处理器
│   │   └── news_handler.go
│   ├── models/           # 数据模型
│   │   └── news.go
│   ├── repositories/     # 数据访问层
│   │   └── news_repository.go
│   └── services/         # 业务逻辑层
│       └── news_service.go
├── pkg/
│   ├── config/           # 配置管理
│   │   └── config.go
│   ├── database/         # 数据库连接
│   │   ├── database.go
│   │   └── grom_logger_writer.go
│   ├── logger/           # 多级别日志系统
│   │   └── logger.go
│   └── utils/            # 工具类
│       ├── jwt.go        # JWT工具
│       └── response.go   # 统一响应格式
├── config/               # 配置文件目录
│   ├── config.yaml       # 主配置文件
│   └── config.pro.yaml   # 生产环境配置
├── docs/                 # 文档
├── routes/               # 路由配置
│   ├── auth_routes.go    # 认证路由
│   ├── middleware/       # 中间件
│   │   ├── cors.go
│   │   ├── jwt_auth.go   # JWT认证中间件
│   │   └── middleware.go
│   ├── news_routes.go    # 新闻路由
│   └── routes.go         # 主路由配置
├── log/                  # 日志文件目录
│   └── 2025-12-11/     # 按日期分类的日志
├── go.mod
├── go.sum
└── README.md
```

## 运行命令

```bash
# 安装依赖
go mod tidy

# 直接运行
go run cmd/server/main.go -mod=prod/dev/test

# 构建并运行
go build -o starlive .
./starlive -mod=prod/dev/test
```