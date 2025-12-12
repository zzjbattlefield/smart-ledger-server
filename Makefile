.PHONY: build run test clean docker-build docker-up docker-down tidy

# 变量
APP_NAME := smart-ledger-server
BUILD_DIR := ./bin
MAIN_FILE := ./cmd/server/main.go
CONFIG_FILE := ./configs/config.yaml

# 构建
build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run:
	@go run $(MAIN_FILE) -config $(CONFIG_FILE)

# 热重载运行（需要安装 air: go install github.com/cosmtrek/air@latest）
dev:
	@air -c .air.toml

# 测试
test:
	@go test -v ./...

# 测试覆盖率
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 清理
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# 整理依赖
tidy:
	@go mod tidy

# Docker构建
docker-build:
	@docker build -t $(APP_NAME):latest .

# Docker启动
docker-up:
	@docker-compose up -d

# Docker停止
docker-down:
	@docker-compose down

# Docker日志
docker-logs:
	@docker-compose logs -f app

# 生成Swagger文档（需要安装 swag: go install github.com/swaggo/swag/cmd/swag@latest）
swagger:
	@swag init -g $(MAIN_FILE) -o ./docs

# 数据库迁移
migrate-up:
	@go run ./cmd/migrate/main.go up

migrate-down:
	@go run ./cmd/migrate/main.go down

# 帮助
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make tidy           - Tidy go modules"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make docker-logs    - Show Docker logs"
	@echo "  make swagger        - Generate Swagger docs"
