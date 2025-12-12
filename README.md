# Smart Ledger Server

智能账本后端服务，基于 Go 语言开发的记账应用 API 服务器。

## 功能特性

- **用户管理** - 注册、登录、JWT 认证
- **账单管理** - 收入/支出记录的增删改查
- **分类管理** - 自定义收支分类
- **统计报表** - 按日/周/月统计收支情况
- **AI 记账助手** - 支持自然语言记账（集成通义千问/OpenAI）

## 技术栈

- **框架**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **数据库**: MySQL
- **缓存**: Redis
- **认证**: JWT
- **日志**: Zap
- **配置**: Viper

## 快速开始

### 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 6.0+

### 安装

```bash
# 克隆项目
git clone <repository-url>
cd smart-ledger/server

# 安装依赖
go mod download
```

### 配置

```bash
# 复制配置文件模板
cp configs/config.example.yaml configs/config.yaml

# 编辑配置文件，填入数据库、Redis、JWT 密钥等信息
vim configs/config.yaml
```

### 运行

```bash
# 开发模式（热重载）
make dev

# 或直接运行
make run

# 构建
make build
./server
```

### Docker 部署

```bash
# 构建镜像
make docker-build

# 启动服务
make docker-up
```

## 项目结构

```
.
├── cmd/
│   └── server/          # 程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── container/       # 依赖注入
│   ├── handler/         # HTTP 处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   └── pkg/             # 内部工具包
├── pkg/
│   └── errcode/         # 错误码定义
├── migrations/          # 数据库迁移
├── configs/             # 配置文件
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## API 概览

| 模块 | 端点 | 说明 |
|------|------|------|
| 用户 | `POST /api/v1/auth/register` | 用户注册 |
| 用户 | `POST /api/v1/auth/login` | 用户登录 |
| 账单 | `GET /api/v1/bills` | 获取账单列表 |
| 账单 | `POST /api/v1/bills` | 创建账单 |
| 分类 | `GET /api/v1/categories` | 获取分类列表 |
| 统计 | `GET /api/v1/stats` | 获取统计数据 |
| AI | `POST /api/v1/ai/parse` | AI 解析记账内容 |

## 开发命令

```bash
make build          # 编译
make run            # 运行
make dev            # 热重载开发
make test           # 运行测试
make test-coverage  # 测试覆盖率
make clean          # 清理构建产物
make migrate-up     # 执行数据库迁移
make migrate-down   # 回滚数据库迁移
```

## License

[MIT](LICENSE)
