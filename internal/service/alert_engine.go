package service

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

// AlertEngine 告警规则引擎
type AlertEngine struct {
	db      *gorm.DB
	running bool
}

func NewAlertEngine(db *gorm.DB) *AlertEngine {
	return &AlertEngine{db: db}
}

// Start 启动告警检查循环（每 10 秒）
func (e *AlertEngine) Start() {
	e.running = true
	go e.loop()
	log.Println("🚨 Alert engine started (10s interval)")
}

func (e *AlertEngine) loop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for e.running {
		<-ticker.C
		e.checkAll()
	}
}

func (e *AlertEngine) Stop() {
	e.running = false
}

// checkAll 执行所有规则检查
func (e *AlertEngine) checkAll() {
	e.checkBatteryLow()
	e.checkTaskTimeout()
	e.checkCommLost()
	e.checkFaults()
}

// checkBatteryLow 电量低告警
func (e *AlertEngine) checkBatteryLow() {
	var robots []model.Robot
	e.db.Where("battery_pct < ? AND battery_status != ?", 20, "critical").
		Find(&robots)

	for _, r := range robots {
		// 检查是否已有未解决的同类告警
		var count int64
		e.db.Model(&model.Alert{}).
			Where("robot_id = ? AND alert_type = ? AND status = ?", r.ID, "battery_low", "unack").
			Count(&count)
		if count > 0 {
			continue
		}

		severity := "warning"
		if r.BatteryPct <= 5 {
			severity = "critical"
		}
		alert := model.Alert{
			AlertType: "battery_low",
			Severity:  severity,
			RobotID:   &r.ID,
			Title:     fmt.Sprintf("%s 电量过低", r.Name),
			Content:   fmt.Sprintf("%s (%.1f%%) 电量低于20%%阈值，当前状态: %s", r.Name, r.BatteryPct, r.BatteryStatus),
			Status:    "unack",
			CreatedAt: time.Now(),
		}
		e.db.Create(&alert)
		log.Printf("⚠️  [ALERT] battery_low | %s | %.1f%%", r.RobotCode, r.BatteryPct)
	}
}

// checkTaskTimeout 任务超时告警
func (e *AlertEngine) checkTaskTimeout() {
	var tasks []model.Task
	e.db.Where("status IN ? AND expected_end < ?", []string{"running", "assigned"}, time.Now()).
		Preload("Robot").Find(&tasks)

	for _, t := range tasks {
		var count int64
		e.db.Model(&model.Alert{}).
			Where("task_id = ? AND alert_type = ? AND status = ?", t.ID, "task_timeout", "unack").
			Count(&count)
		if count > 0 {
			continue
		}

		robotName := "未指派"
		if t.Robot != nil {
			robotName = t.Robot.Name
		}
		alert := model.Alert{
			AlertType: "task_timeout",
			Severity:  "warning",
			RobotID:   t.RobotID,
			TaskID:    &t.ID,
			Title:     fmt.Sprintf("任务 %s 超时", t.TaskCode),
			Content:   fmt.Sprintf("任务 %s (%s) 已超过预期完成时间，机器人: %s", t.TaskCode, t.TaskType, robotName),
			Status:    "unack",
			CreatedAt: time.Now(),
		}
		e.db.Create(&alert)
		log.Printf("⚠️  [ALERT] task_timeout | %s | robot=%s", t.TaskCode, robotName)
	}
}

// checkCommLost 通信中断告警
func (e *AlertEngine) checkCommLost() {
	threshold := time.Now().Add(-30 * time.Second)

	var robots []model.Robot
	e.db.Where("comm_status = ?", "online").
		Where("last_heartbeat < ? OR last_heartbeat IS NULL", threshold).
		Find(&robots)

	for _, r := range robots {
		// 标记为离线
		e.db.Model(&r).Update("comm_status", "offline")

		alert := model.Alert{
			AlertType: "comm_lost",
			Severity:  "critical",
			RobotID:   &r.ID,
			Title:     fmt.Sprintf("%s 通信中断", r.Name),
			Content:   fmt.Sprintf("%s 超过30秒未收到心跳，最后心跳: %v", r.Name, r.LastHeartbeat),
			Status:    "unack",
			CreatedAt: time.Now(),
		}
		e.db.Create(&alert)
		log.Printf("🚨 [ALERT] comm_lost | %s", r.RobotCode)
	}
}

// checkFaults 故障告警
func (e *AlertEngine) checkFaults() {
	var robots []model.Robot
	e.db.Where("status = ?", "fault").Find(&robots)

	for _, r := range robots {
		var count int64
		e.db.Model(&model.Alert{}).
			Where("robot_id = ? AND alert_type = ? AND status = ?", r.ID, "fault", "unack").
			Count(&count)
		if count > 0 {
			continue
		}

		alert := model.Alert{
			AlertType: "fault",
			Severity:  "critical",
			RobotID:   &r.ID,
			Title:     fmt.Sprintf("%s 发生故障", r.Name),
			Content:   fmt.Sprintf("%s 状态为 fault，需人工检查", r.Name),
			Status:    "unack",
			CreatedAt: time.Now(),
		}
		e.db.Create(&alert)
		log.Printf("🚨 [ALERT] fault | %s", r.RobotCode)
	}
}
