package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/config"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

// TelemetryPayload 机器人遥测 JSON 载荷
type TelemetryPayload struct {
	RobotCode string `json:"robot_code"`
	Timestamp string `json:"timestamp"`
	Battery   struct {
		Percentage  float64 `json:"percentage"`
		Voltage     float64 `json:"voltage"`
		Current     float64 `json:"current"`
		Temperature float64 `json:"temperature"`
	} `json:"battery"`
	Location struct {
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
		Theta float64 `json:"theta"`
		Zone  string  `json:"zone"`
	} `json:"location"`
	Motion struct {
		Speed        float64 `json:"speed"`
		Acceleration float64 `json:"acceleration"`
		Mode         string  `json:"mode"`
	} `json:"motion"`
	Light struct {
		Mode       string `json:"mode"`
		Color      string `json:"color"`
		Brightness int    `json:"brightness"`
	} `json:"light"`
	Status string `json:"status"`
	TaskID *uint  `json:"task_id"`
}

var db *gorm.DB

func main() {
	cfg := config.Load()

	var err error
	db, err = gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	log.Println("✅ MQTT Bridge — PostgreSQL connected")

	// ── MQTT 连接 ──
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.URI()).
		SetClientID("rss-mqtt-bridge").
		SetCleanSession(true).
		SetKeepAlive(30 * time.Second).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(10 * time.Second).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("⚠️  MQTT connection lost: %v", err)
		}).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Println("✅ MQTT connected — subscribing to robot topics...")

			// 订阅所有机器人遥测数据
			c.Subscribe("robots/+/telemetry", 1, onTelemetry)
			// 订阅所有心跳
			c.Subscribe("robots/+/heartbeat", 1, onHeartbeat)
			// 订阅状态变更
			c.Subscribe("robots/+/status", 1, onStatus)
		})

	if cfg.MQTT.Username != "" {
		opts.SetUsername(cfg.MQTT.Username)
		opts.SetPassword(cfg.MQTT.Password)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ MQTT connection failed: %v", token.Error())
	}

	// ── 优雅退出 ──
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Println("🚀 MQTT Bridge running...")
	<-quit

	log.Println("Shutting down...")
	client.Disconnect(250)
}

// onTelemetry 处理遥测数据
func onTelemetry(client mqtt.Client, msg mqtt.Message) {
	var payload TelemetryPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		log.Printf("⚠️  Invalid telemetry: %v", err)
		return
	}

	// 查找机器人
	var robot model.Robot
	if err := db.Where("robot_code = ?", payload.RobotCode).First(&robot).Error; err != nil {
		log.Printf("⚠️  Unknown robot: %s", payload.RobotCode)
		return
	}

	now := time.Now()

	// 更新机器人实时状态
	updates := map[string]interface{}{
		"battery_pct":    payload.Battery.Percentage,
		"battery_status": batteryStatus(payload.Battery.Percentage),
		"location_x":     payload.Location.X,
		"location_y":     payload.Location.Y,
		"location_theta": payload.Location.Theta,
		"map_zone":       payload.Location.Zone,
		"speed":          payload.Motion.Speed,
		"light_mode":     payload.Light.Mode,
		"light_color":    payload.Light.Color,
		"light_brightness": payload.Light.Brightness,
		"status":         payload.Status,
		"comm_status":    "online",
		"last_heartbeat": now,
		"updated_at":     now,
	}
	db.Model(&robot).Updates(updates)

	// 写入时序数据
	telemetry := model.RobotTelemetry{
		Time:           now,
		RobotID:        robot.ID,
		BatteryPct:     payload.Battery.Percentage,
		LocationX:      payload.Location.X,
		LocationY:      payload.Location.Y,
		LocationTheta:  payload.Location.Theta,
		Speed:          payload.Motion.Speed,
		LightMode:      payload.Light.Mode,
		LightColor:     payload.Light.Color,
		LightBrightness: payload.Light.Brightness,
		Status:         payload.Status,
		CommStatus:     "online",
	}
	db.Create(&telemetry)
}

// onHeartbeat 处理心跳
func onHeartbeat(client mqtt.Client, msg mqtt.Message) {
	robotCode := extractRobotCode(msg.Topic())
	if robotCode == "" {
		return
	}

	now := time.Now()
	db.Model(&model.Robot{}).
		Where("robot_code = ?", robotCode).
		Updates(map[string]interface{}{
			"comm_status":    "online",
			"last_heartbeat": now,
		})
}

// onStatus 处理状态变更
func onStatus(client mqtt.Client, msg mqtt.Message) {
	robotCode := extractRobotCode(msg.Topic())
	if robotCode == "" {
		return
	}

	var statusUpdate struct {
		Status string `json:"status"`
	}
	json.Unmarshal(msg.Payload(), &statusUpdate)

	if statusUpdate.Status != "" {
		db.Model(&model.Robot{}).
			Where("robot_code = ?", robotCode).
			Update("status", statusUpdate.Status)
	}
}

func batteryStatus(pct float64) string {
	switch {
	case pct <= 5:
		return "critical"
	case pct <= 20:
		return "low"
	case pct >= 95:
		return "full"
	default:
		return "normal"
	}
}

func extractRobotCode(topic string) string {
	// robots/{robot_code}/telemetry → robot_code
	parts := make([]string, 0)
	start := 0
	for i, c := range topic {
		if c == '/' {
			parts = append(parts, topic[start:i])
			start = i + 1
		}
	}
	parts = append(parts, topic[start:])

	if len(parts) >= 2 && parts[0] == "robots" {
		return parts[1]
	}
	return ""
}

func init() {
	// 确保日志带时间戳
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// 占位（避免 unused import）
var _ = fmt.Sprintf
