# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

智能账本后端服务 - 基于 Go 语言的记账 API 服务器，支持 AI 识别支付截图自动记账。

## 常用命令

```bash
# 开发
make run              # 运行服务 (go run)
make dev              # 热重载开发 (需要 air)
make build            # 编译到 bin/smart-ledger-server

# 测试
make test             # 运行所有测试
make test-coverage    # 生成覆盖率报告
go test -v ./internal/service/...  # 运行单个包测试

# 依赖
make tidy             # 整理 go.mod

# Docker
make docker-build     # 构建镜像
make docker-up        # 启动容器
make docker-down      # 停止容器

# 数据库迁移 (使用 goose)
goose -dir migrations mysql "user:pass@tcp(localhost:3306)/dbname?parseTime=true" up
goose -dir migrations mysql "..." down
goose -dir migrations create <name> sql   # 创建 SQL 迁移
goose -dir migrations create <name> go    # 创建 Go 迁移
```

## 架构设计

### 三层架构 + 依赖注入

```
Handler → Service → Repository
    ↓         ↓          ↓
  (接口)    (接口)     (GORM)
```

- **Handler**: HTTP 请求处理，参数校验，调用 Service
- **Service**: 业务逻辑，依赖 Repository 接口（依赖反转）
- **Repository**: 数据访问层，封装 GORM 操作

### 依赖容器 (`internal/container/container.go`)

所有依赖通过 `Container` 统一管理和注入：
- 初始化顺序：Repositories → Services → Handlers
- 通过访问器方法获取实例（如 `ctn.UserService()`）

### 接口定义

- `internal/service/service_interfaces.go` - Handler 依赖的 Service 接口
- `internal/service/repo_interfaces.go` - Service 依赖的 Repository 接口

### 数据流

1. 请求 → Router → Middleware → Handler
2. Handler 解析参数 → 调用 Service
3. Service 执行业务逻辑 → 调用 Repository
4. Repository 操作数据库 → 返回 Model
5. Service 转换为 DTO → 返回给 Handler

## 目录结构要点

- `cmd/server/` - 程序入口、启动引导、路由注册
- `internal/config/` - Viper 配置管理
- `internal/model/` - GORM 数据模型
- `internal/model/dto/` - 请求/响应 DTO
- `internal/pkg/ai/` - AI 客户端（OpenAI 兼容接口）
- `internal/pkg/response/` - 统一响应封装
- `pkg/errcode/` - 错误码定义（分段管理）
- `migrations/` - Goose 数据库迁移

## 错误码规范

错误码定义在 `pkg/errcode/errcode.go`，按模块分段：
- 10000-19999: 通用错误
- 20000-29999: 认证错误
- 30000-39999: 用户错误
- 40000-49999: 账单错误
- 50000-59999: AI 错误
- 60000-69999: 分类错误

## AI 服务

支持 OpenAI 兼容 API（可配置通义千问等）：
- 单张/批量图片识别
- 自动创建账单
- Worker Pool 并发处理
- Rate Limiter 限流

## 配置

配置文件：`configs/config.yaml`（从 `config.example.yaml` 复制）

关键配置项：
- `server.mode`: debug/release
- `ai.provider`: openai/qwen
- `jwt.secret`: JWT 密钥（生产环境必须修改）

git commit时 永远不要提交下面这段话
```
🤖 Generated with [Claude Code](https://claude.com/claude-code)
Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```