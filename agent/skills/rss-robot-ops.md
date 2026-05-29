---
name: rss-robot-ops
description: 机器人运维操作 — 查询状态、下发指令、查看遥测数据
trigger: 机器人|状态|电量|指令|控制|RBT
---

# 机器人运维操作

## 查询机器人列表
调用 GET `/api/v1/robots?page=1&page_size=20`
支持筛选：
- `?status=running` — 只显示运行中的
- `?model=AGV-2000` — 按型号筛选
- `?battery_min=20` — 电量高于 20%

## 查看机器人详情
GET `/api/v1/robots/{id}` — 获取完整信息含电量、位置、灯光

## 下发控制指令
POST `/api/v1/robots/{id}/commands`
```json
{"command": "start", "params": {}}
{"command": "pause", "params": {}}
{"command": "stop", "params": {"emergency": true}}
{"command": "charge", "params": {}}
{"command": "light", "params": {"mode": "breathe", "color": "red", "brightness": 80}}
```

## 常见场景
- "看看有哪些机器人在运行" → GET /robots?status=running
- "RBT-001 电量多少" → GET /robots 筛选 robot_code
- "让巡逻一号回充" → POST /robots/:id/commands {command: "charge"}
