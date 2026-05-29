# ── 机器人调度系统 Makefile ──

.PHONY: help dev up down build test

help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: up ## 启动开发环境（Docker 依赖 + API 服务）
	@echo "🚀 Starting API server..."
	cd cmd/api && go run .

up: ## 启动所有 Docker 依赖服务
	@echo "📦 Starting Docker services..."
	docker compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 5
	@docker compose ps

down: ## 停止所有 Docker 服务
	docker compose down

build: ## 编译所有服务
	go build -o bin/api ./cmd/api
	go build -o bin/mqtt-bridge ./cmd/mqtt-bridge
	@echo "✅ Build complete: bin/api bin/mqtt-bridge"

test: ## 运行测试
	go test ./... -v -count=1

lint: ## 代码检查
	go vet ./...

migrate: ## 运行数据库迁移
	@echo "📊 Running migrations..."
	docker compose exec -T postgres psql -U rss -d robot_scheduling -f /docker-entrypoint-initdb.d/001_init.sql

clean: ## 清理编译产物
	rm -rf bin/
	docker compose down -v

# ── Docker Compose 管理 ──
db-shell: ## 进入 PostgreSQL shell
	docker compose exec postgres psql -U rss -d robot_scheduling

redis-cli: ## 进入 Redis CLI
	docker compose exec redis redis-cli

emqx-dashboard: ## 打开 EMQX Dashboard
	@echo "EMQX Dashboard: http://localhost:18083 (admin/admin_dev_2026)"

rabbitmq-dashboard: ## 打开 RabbitMQ Dashboard
	@echo "RabbitMQ Dashboard: http://localhost:15672 (rss/rss_dev_2026)"

minio-dashboard: ## 打开 MinIO Console
	@echo "MinIO Console: http://localhost:9001 (minioadmin/minio_dev_2026)"
