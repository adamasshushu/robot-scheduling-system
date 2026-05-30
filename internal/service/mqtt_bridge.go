// ── MQTT 桥接层：连接机器人底盘 ──
package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

// ── 协议数据格式 ──

// RobotStatusReport 机器人上报的状态数据
type RobotStatusReport struct {
	RobotCode   string  `json:"robot_code"`
	BatteryPct  int     `json:"battery_pct"`
	LocationX   float64 `json:"location_x"`
	LocationY   float64 `json:"location_y"`
	MapZone     string  `json:"map_zone"`
	Speed       float64 `json:"speed"`
	Status      string  `json:"status"`
	TaskID      uint    `json:"task_id"`
	ProgressPct int     `json:"progress_pct"`
	Temperature float64 `json:"temperature"`
	LightMode   string  `json:"light_mode"`
	Timestamp   int64   `json:"timestamp"`
}

// RobotCommand 系统下发的指令
type RobotCommand struct {
	Command   string                 `json:"command"`
	Params    map[string]interface{} `json:"params,omitempty"`
	TaskID    uint                   `json:"task_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// RobotResponse 机器人对指令的回应
type RobotResponse struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"` // ok / error
	Message   string `json:"message"`
	Data      string `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// RobotAlert 机器人告警
type RobotAlert struct {
	RobotCode string `json:"robot_code"`
	Level     string `json:"level"`  // info / warning / error / critical
	Type      string `json:"type"`   // battery / collision / sensor / motor / system
	Message   string `json:"message"`
	Data      string `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// ── MQTT 主题 ──
func robotTopic(robotCode, suffix string) string {
	return fmt.Sprintf("robot/%s/%s", robotCode, suffix)
}

// heartbeatTTL 心跳超时阈值
const heartbeatTTL = 30 * time.Second

// MQTTBridge MQTT 桥接器
type MQTTBridge struct {
	client  mqtt.Client
	db      *gorm.DB
	hub     *WSHub

	// 心跳追踪
	mu       sync.RWMutex
	recentAlive map[string]time.Time    // robotCode → 最后心跳时间
}

// NewMQTTBridge 创建 MQTT 桥接器
func NewMQTTBridge(db *gorm.DB, hub *WSHub) *MQTTBridge {
	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		broker = "localhost"
	}
	port := os.Getenv("MQTT_PORT")
	if port == "" {
		port = "1883"
	}

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port)).
		SetClientID("rss-backend").
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(10 * time.Second).
		SetConnectRetry(true).
		SetKeepAlive(10 * time.Second)

	client := mqtt.NewClient(opts)

	bridge := &MQTTBridge{
		client:      client,
		db:          db,
		hub:         hub,
		recentAlive: make(map[string]time.Time),
	}

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("⚠️  MQTT 连接失败: %v (机器人底盘接入不可用)", token.Error())
		return bridge
	}

	// 订阅所有机器人主题
	bridge.subscribe("robot/+/status",   1, bridge.handleStatus)
	bridge.subscribe("robot/+/heartbeat", 1, bridge.handleHeartbeat)
	bridge.subscribe("robot/+/alert",     1, bridge.handleAlert)
	bridge.subscribe("robot/+/response", 1, bridge.handleResponse)

	// 心跳超时检测
	go bridge.heartbeatWatcher()

	log.Println("✅ MQTT 桥接器已启动 — 机器人底盘可接入")
	return bridge
}

func (b *MQTTBridge) subscribe(topic string, qos byte, handler func(mqtt.Client, mqtt.Message)) {
	if b.client == nil || !b.client.IsConnected() {
		return
	}
	token := b.client.Subscribe(topic, qos, handler)
	token.Wait()
	if token.Error() != nil {
		log.Printf("⚠️  MQTT 订阅 %s 失败: %v", topic, token.Error())
	} else {
		log.Printf("  ✓ 订阅: %s", topic)
	}
}

// PublishCommand 向指定机器人发送指令
func (b *MQTTBridge) PublishCommand(robotCode string, cmd RobotCommand) {
	if b.client == nil || !b.client.IsConnected() {
		log.Println("⚠️  MQTT 未连接，无法发送指令")
		return
	}
	cmd.Timestamp = time.Now().UnixMilli()
	payload, _ := json.Marshal(cmd)
	topic := robotTopic(robotCode, "command")
	b.client.Publish(topic, 1, false, payload)
	log.Printf("📤 %s → %s %s", topic, cmd.Command, shortJSON(payload))
}

// IsOnline 判断机器人是否在线
func (b *MQTTBridge) IsOnline(robotCode string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	last, ok := b.recentAlive[robotCode]
	return ok && time.Since(last) < heartbeatTTL
}

// ── 消息处理器 ──

func (b *MQTTBridge) handleStatus(c mqtt.Client, msg mqtt.Message) {
	var report RobotStatusReport
	if err := json.Unmarshal(msg.Payload(), &report); err != nil {
		log.Printf("⚠️  status 解析失败: %v", err)
		return
	}

	// 更新数据库机器人状态
	updates := map[string]interface{}{
		"battery_pct":  report.BatteryPct,
		"location_x":   report.LocationX,
		"location_y":   report.LocationY,
		"map_zone":     report.MapZone,
		"speed":        report.Speed,
		"status":       report.Status,
		"progress_pct": report.ProgressPct,
	}
	if report.Temperature > 0 {
		updates["temperature"] = report.Temperature
	}
	if report.LightMode != "" {
		updates["light_mode"] = report.LightMode
	}

	result := b.db.Model(&model.Robot{}).Where("robot_code = ?", report.RobotCode).Updates(updates)
	if result.RowsAffected == 0 {
		log.Printf("⚠️  未找到机器人: %s (请先在系统中注册)", report.RobotCode)
		return
	}

	// 通过 WebSocket 推送实时状态
	if b.hub != nil {
		b.hub.Broadcast(WSMessage{
			Type: "robot_status",
			Data: map[string]interface{}{
				"robot_code": report.RobotCode,
				"battery_pct": report.BatteryPct,
				"location_x":  report.LocationX,
				"location_y":  report.LocationY,
				"map_zone":    report.MapZone,
				"speed":       report.Speed,
				"status":      report.Status,
			},
		})
	}

	log.Printf("📡 %s: 电量=%d%% 位置=(%.1f,%.1f) 状态=%s",
		report.RobotCode, report.BatteryPct, report.LocationX, report.LocationY, report.Status)
}

func (b *MQTTBridge) handleHeartbeat(c mqtt.Client, msg mqtt.Message) {
	var hb struct {
		RobotCode string `json:"robot_code"`
		Version   string `json:"version"`
		Uptime    int64  `json:"uptime"`
	}
	json.Unmarshal(msg.Payload(), &hb)

	robotCode := hb.RobotCode
	if robotCode == "" {
		// 从 topic 提取: robot/ROBOT_CODE/heartbeat
		parts := strings.Split(msg.Topic(), "/")
		if len(parts) >= 2 { robotCode = parts[1] }
	}
	if robotCode == "" { return }

	b.mu.Lock()
	b.recentAlive[robotCode] = time.Now()
	b.mu.Unlock()

	// 标记在线状态
	b.db.Model(&model.Robot{}).Where("robot_code = ?", robotCode).
		Updates(map[string]interface{}{"status": "standby", "updated_at": time.Now()})
}

func (b *MQTTBridge) handleAlert(c mqtt.Client, msg mqtt.Message) {
	var alert RobotAlert
	if err := json.Unmarshal(msg.Payload(), &alert); err != nil {
		log.Printf("⚠️  alert 解析失败: %v", err)
		return
	}

	// 查找机器人ID
	var robot model.Robot
	if err := b.db.Where("robot_code = ?", alert.RobotCode).First(&robot).Error; err != nil {
		log.Printf("⚠️  未找到告警机器人: %s", alert.RobotCode)
		return
	}

	// 写入告警表
	rid := robot.ID
	alarm := model.Alert{
		RobotID:   &rid,
		RobotCode: alert.RobotCode,
		Level:     alert.Level,
		Type:      alert.Type,
		Message:   alert.Message,
		Data:      alert.Data,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	b.db.Create(&alarm)

	// WebSocket 推送告警
	if b.hub != nil {
		b.hub.Broadcast(WSMessage{
			Type: "robot_alert",
			Data: alarm,
		})
	}

	log.Printf("🚨 告警 %s/%s: %s", alert.RobotCode, alert.Level, alert.Message)
}

func (b *MQTTBridge) handleResponse(c mqtt.Client, msg mqtt.Message) {
	var resp RobotResponse
	if err := json.Unmarshal(msg.Payload(), &resp); err != nil {
		log.Printf("⚠️  response 解析失败: %v", err)
		return
	}
	log.Printf("📥 响应: cmd=%s status=%s msg=%s", resp.CommandID, resp.Status, resp.Message)
}

// ── 心跳超时检测 ──

func (b *MQTTBridge) heartbeatWatcher() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.RLock()
		var offline []string
		for code, last := range b.recentAlive {
			if time.Since(last) > heartbeatTTL {
				offline = append(offline, code)
			}
		}
		b.mu.RUnlock()

		for _, code := range offline {
			b.mu.Lock()
			// 二次确认
			if time.Since(b.recentAlive[code]) > heartbeatTTL {
				b.db.Model(&model.Robot{}).Where("robot_code = ?", code).
					Update("status", "offline")
				if b.hub != nil {
					b.hub.Broadcast(WSMessage{
						Type: "robot_offline",
						Data: map[string]interface{}{"robot_code": code},
					})
				}
				delete(b.recentAlive, code)
				log.Printf("⚠️  机器人 %s 心跳超时，标记离线", code)
			}
			b.mu.Unlock()
		}
	}
}

// 辅助：安全打印 JSON
func shortJSON(b []byte) string {
	if len(b) > 120 {
		return string(b[:120]) + "..."
	}
	return string(b)
}
