# Smart Ledger Server

智能账本后端服务，基于 Go 语言开发的记账应用 API 服务器。

## 功能特性

- **用户管理** - 注册、登录、JWT 认证、个人资料管理
- **账单管理** - 收入/支出记录的增删改查
- **分类管理** - 自定义收支分类，支持系统预设模板
- **统计报表** - 收支汇总统计、分类统计分析
- **AI 截图识别** - 上传支付截图自动识别并创建账单（支持通义千问/OpenAI）

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
│   └── server/          # 程序入口、路由注册
├── internal/
│   ├── config/          # 配置管理 (Viper)
│   ├── container/       # 依赖注入容器
│   ├── handler/         # HTTP 处理器
│   ├── middleware/      # 中间件 (认证、日志、限流、CORS)
│   ├── model/           # 数据模型
│   │   └── dto/         # 请求/响应 DTO
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   └── pkg/             # 内部工具包
│       ├── ai/          # AI 客户端 (OpenAI 兼容接口)
│       ├── database/    # 数据库连接 (MySQL、Redis)
│       ├── logger/      # 日志工具 (Zap)
│       └── response/    # 统一响应封装
├── pkg/
│   └── errcode/         # 错误码定义
├── migrations/          # 数据库迁移 (Goose)
├── configs/             # 配置文件
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## API 概览

| 模块 | 端点 | 说明 |
|------|------|------|
| 健康检查 | `GET /health` | 服务健康检查 |
| 用户 | `POST /v1/user/register` | 用户注册 |
| 用户 | `POST /v1/user/login` | 用户登录 |
| 用户 | `GET /v1/user/profile` | 获取个人资料 |
| 用户 | `PUT /v1/user/profile` | 更新个人资料 |
| 分类 | `GET /v1/categories` | 获取分类列表 |
| 分类 | `POST /v1/categories` | 创建分类 |
| 分类 | `PUT /v1/categories/:id` | 更新分类 |
| 分类 | `DELETE /v1/categories/:id` | 删除分类 |
| 账单 | `GET /v1/bills` | 获取账单列表 |
| 账单 | `GET /v1/bills/:id` | 获取账单详情 |
| 账单 | `POST /v1/bills` | 创建账单 |
| 账单 | `PUT /v1/bills/:id` | 更新账单 |
| 账单 | `DELETE /v1/bills/:id` | 删除账单 |
| 统计 | `GET /v1/stats/summary` | 获取收支汇总 |
| 统计 | `GET /v1/stats/category` | 获取分类统计 |
| AI | `POST /v1/ai/recognize` | 识别支付截图 |
| AI | `POST /v1/ai/recognize-and-save` | 识别截图并创建账单 |

## 开发命令

```bash
make build          # 编译
make run            # 运行
make dev            # 热重载开发 (需要 air)
make test           # 运行测试
make test-coverage  # 测试覆盖率
make clean          # 清理构建产物
make tidy           # 整理 go.mod
make swagger        # 生成 Swagger 文档 (需要 swag)
make docker-build   # 构建 Docker 镜像
make docker-up      # 启动 Docker 容器
make docker-down    # 停止 Docker 容器
make docker-logs    # 查看 Docker 日志
make migrate-up     # 执行数据库迁移
make migrate-down   # 回滚数据库迁移
```

## License

[MIT](LICENSE)
