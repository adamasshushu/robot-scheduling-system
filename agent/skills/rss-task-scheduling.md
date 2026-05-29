---
name: rss-task-scheduling
description: 任务调度操作 — 创建、指派、追踪任务
trigger: 任务|调度|搬运|巡逻|指派|TSK
---

# 任务调度操作

## 创建任务
POST `/api/v1/tasks`
```json
{
  "task_type": "transport",
  "target_location": "B区-装配线",
  "source_location": "A区-仓库",
  "target_x": 50.0,
  "target_y": 30.0,
  "priority": 3,
  "description": "搬运原材料到装配线"
}
```

任务类型：transport（搬运）、patrol（巡逻）、charge（充电）

## 智能指派
POST `/api/v1/tasks/{id}/assign`
```json
{"auto": true}
```
系统自动选择：电量 > 20%、standby 状态、距离最近

## 手动指派
POST `/api/v1/tasks/{id}/assign`
```json
{"robot_id": 1}
```

## 任务追踪
GET `/api/v1/tasks?status=assigned` — 查看已指派任务
GET `/api/v1/tasks/{id}` — 查看任务详情含进度

## 取消任务
POST `/api/v1/tasks/{id}/cancel`

## 常见场景
- "派一个机器人去 B区搬运" → 创建 transport 任务 → auto assign
- "当前有哪些任务在跑" → GET /tasks?status=running
- "取消 TSK-001 这个任务" → POST /tasks/:id/cancel
