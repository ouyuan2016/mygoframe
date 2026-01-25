# MyGoFrame API Server

这是一个基于 Go 语言构建的、结构清晰、功能完善的现代化 Web API 服务器框架。它遵循分层架构和依赖注入的设计原则，集成了日志、配置、数据库、缓存和任务队列等常用组件，旨在为快速开发高质量的后端服务提供一个坚实的基础。

## 1. 功能特性

- **配置管理**: 使用 `Viper` 库加载 YAML 配置文件 (`config.yaml`)，支持多环境配置和热加载。
- **日志系统**: 集成 `Zap` 和 `Lumberjack`，提供高性能的结构化日志。支持：
  - 日志级别分离（如 `info`, `error`）。
  - 同时输出到控制台和文件。
  - 日志文件按日期自动切割和归档。
- **数据库 ORM**: 使用 `GORM` 作为数据库ORM，支持平滑的数据库连接和关闭。业务相关的数据库迁移（AutoMigrate）逻辑已从基础设施层解耦。
- **缓存管理**: 提供了统一的缓存管理器，支持 `Redis` 和进程内缓存（`in-memory`）两种模式，可根据配置灵活切换。
- **任务队列**: 集成 `Asynq` 实现强大的异步任务处理能力。
  - 支持普通任务、延迟任务和周期性定时任务（Cron Jobs）。
  - 对任务的创建和入队逻辑进行了封装，简化了业务层的调用。
  - 提供了可扩展的定时任务管理方案。
- **认证与授权**: 基于 `JWT` (JSON Web Tokens) 实现无状态的用户认证。
  - 使用策略模式（Strategy Pattern）优雅地处理不同签名算法（HMAC, RSA）。
  - 支持访问令牌（Access Token）和刷新令牌（Refresh Token）机制。
- **分层架构**: 清晰的 `handlers` -> `services` -> `repositories` 分层设计，职责分明，易于维护。
- **依赖注入**: 通过构造函数注入依赖（如数据库连接），实现了模块间的松耦合。
- **优雅停机**: 实现了 HTTP 服务器的优雅启动与关闭，确保在服务停止时能处理完所有进行中的请求。
- **RESTful API**: 提供了用户（注册、登录、获取个人资料）和新闻相关的 API 接口。

## 2. 项目结构

```
/
├── cmd/
│   └── server/
│       └── main.go              # 应用程序主入口，负责初始化和启动服务
├── config/
│   ├── config.yaml              # 默认配置文件
│   └── config.test.yaml         # 测试环境配置文件
├── internal/
│   ├── dto/                     # 数据传输对象 (Data Transfer Objects)
│   ├── handlers/                # HTTP 处理器，负责解析请求和返回响应
│   ├── models/                  # 数据库模型 (GORM models)
│   ├── repositories/            # 数据仓库层，负责与数据库交互
│   ├── services/                # 业务逻辑层
│   └── task/                    # 异步任务定义和处理器
│       ├── email_tasks.go       # 邮件相关任务
│       ├── general_tasks.go     # 通用任务
│       ├── setup.go             # 任务和定时任务的注册
│       └── types.go             # 任务类型常量
├── pkg/
│   ├── cache/                   # 缓存包，支持 Redis 和内存缓存
│   ├── config/                  # 配置加载
│   ├── database/                # 数据库初始化
│   ├── logger/                  # 日志系统
│   ├── queue/                   # 任务队列客户端和管理器
│   └── utils/                   # 通用工具（如响应格式化、JWT）
├── routes/
│   ├── middleware/              # 中间件（如 JWT 认证、CORS）
│   └── routes.go                # 路由定义
├── go.mod
├── go.sum
└── README.md
```

## 3. 启动方式

1.  **克隆项目**
    ```bash
    git clone <your-repository-url>
    cd mygoframe
    ```

2.  **安装依赖**
    ```bash
    go mod tidy
    ```

3.  **配置环境**
    复制 `config/config.yaml` 并根据您的本地环境修改其中的配置，特别是 `database` 和 `redis` 部分。
    ```yaml
    system:
      env: "development"
      addr: 8080
      # ...

    database:
      driver: "mysql"
      source: "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
      # ...

    redis:
      addr: "127.0.0.1:6379"
      password: ""
      db: 0
      # ...
    ```

4.  **运行项目**
    ```bash
    go run ./cmd/server/main.go
    ```

5.  **访问服务**
    服务启动后，您可以通过 `http://localhost:8080` (或您在配置中指定的端口) 访问 API。