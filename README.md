## 项目结构

```
mygoframe/
├── cmd/
│   └── server/
│       └── main.go          # 程序入口
├── config/
│   ├── config.yaml         # 开发环境配置
│   └── config.pro.yaml     # 生产环境配置
├── internal/
│   ├── handlers/           # HTTP处理器
│   ├── models/             # 数据模型
│   ├── repositories/       # 数据访问层
│   └── services/           # 业务逻辑层
├── pkg/
│   ├── config/             # 配置管理
│   ├── database/           # 数据库连接
│   ├── logger/             # 日志系统
│   └── utils/              # 工具函数
└── routes/                 # 路由定义
    └── middleware/         # 中间件
```

## 使用说明

1. 确保已安装Go环境
2. 根据需要修改配置文件
3. 安装依赖：
   ```bash
   go mod tidy
   ```
4. 运行项目：
   ```bash
   go run cmd/server/main.go
   ```
