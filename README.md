# 🤖 机器人调度系统 — Robot Scheduling System (RSS)

集中式多机器人任务调度与状态监控平台。

## 快速开始

```bash
# 1. 启动基础设施
docker compose up -d

# 2. 复制环境变量
cp .env.example .env

# 3. 启动 API 服务
cd cmd/api && go run .

# 4. 启动 MQTT 桥接
cd cmd/mqtt-bridge && go run .
```

## 架构

```
机器人 ─MQTT─▶ MQTT Bridge ─▶ PostgreSQL/TimescaleDB
                    │
前端 ◀─REST/WS── API Server ─▶ Redis (缓存)
                    │
            ┌───────┼───────┐
            ▼       ▼       ▼
        告警引擎  调度引擎  报表引擎
```

## 技术栈

| 层 | 技术 |
|------|---------|
| 后端 | Go 1.25 + Gin + GORM |
| 数据库 | PostgreSQL 16 + TimescaleDB |
| 缓存 | Redis 7 |
| 消息 | EMQX (MQTT) + RabbitMQ |
| 存储 | MinIO |
| 前端 | Vue 3 + Element Plus + Leaflet |
| AI | Adamas Agent (本地集成) |

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| API Server | 8000 | RESTful API |
| MQTT | 1883 | 机器人通信 |
| MQTT WS | 8083 | WebSocket MQTT |
| EMQX Dashboard | 18083 | MQTT 管理 |
| RabbitMQ | 5672/15672 | 消息队列/管理UI |
| MinIO | 9000/9001 | 对象存储/控制台 |

## API 概览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/robots` | 机器人列表 |
| POST | `/api/v1/robots` | 注册机器人 |
| POST | `/api/v1/robots/:id/commands` | 下发控制指令 |
| GET | `/api/v1/tasks` | 任务列表 |
| POST | `/api/v1/tasks` | 创建任务 |
| POST | `/api/v1/tasks/:id/assign` | 指派机器人 |

## 项目状态

- [x] 项目骨架搭建
- [x] Docker 开发环境
- [x] 机器人管理 API
- [x] 任务调度 API
- [x] MQTT 桥接服务
- [x] 数据库迁移
- [x] Adamas Agent 集成
- [ ] 前端开发
- [ ] 告警引擎
- [ ] 调度算法优化
- [ ] 地图可视化
- [ ] 报表系统
