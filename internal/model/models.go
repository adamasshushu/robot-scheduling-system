package model

import (
	"time"

	"gorm.io/gorm"
)

// ── 机器人 ──
type Robot struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	RobotCode      string         `gorm:"uniqueIndex;size:32" json:"robot_code"`
	Model          string         `gorm:"size:64" json:"model"`
	Name           string         `gorm:"size:128" json:"name"`
	BatteryPct     float64        `gorm:"type:decimal(5,2);default:100" json:"battery_pct"`
	BatteryStatus  string         `gorm:"size:16;default:normal" json:"battery_status"`
	LocationX      float64        `gorm:"type:decimal(10,3)" json:"location_x"`
	LocationY      float64        `gorm:"type:decimal(10,3)" json:"location_y"`
	LocationTheta  float64        `gorm:"type:decimal(8,3)" json:"location_theta"`
	MapZone        string         `gorm:"size:64" json:"map_zone"`
	Speed          float64        `gorm:"type:decimal(6,2);default:0" json:"speed"`
	LightMode      string         `gorm:"size:16;default:off" json:"light_mode"`
	LightColor     string         `gorm:"size:16;default:white" json:"light_color"`
	LightBrightness int           `gorm:"default:100" json:"light_brightness"`
	Status         string         `gorm:"size:16;default:standby" json:"status"`
	CommStatus     string         `gorm:"size:16;default:online" json:"comm_status"`
	LastHeartbeat  *time.Time     `json:"last_heartbeat"`
	IPAddress      string         `gorm:"size:45" json:"ip_address"`
	FirmwareVer    string         `gorm:"size:32" json:"firmware_ver"`
	MaxPayload     float64        `gorm:"type:decimal(8,2)" json:"max_payload"`
	MaxSpeed       float64        `gorm:"type:decimal(6,2)" json:"max_speed"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// ── 任务 ──
type Task struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	TaskCode       string         `gorm:"uniqueIndex;size:32" json:"task_code"`
	TaskType       string         `gorm:"size:32" json:"task_type"`
	RobotID        *uint          `json:"robot_id"`
	Robot          *Robot         `gorm:"foreignKey:RobotID" json:"robot,omitempty"`
	Priority       int            `gorm:"default:5" json:"priority"`
	Status         string         `gorm:"size:16;default:pending" json:"status"`
	SourceLocation string         `gorm:"size:128" json:"source_location"`
	TargetLocation string         `gorm:"size:128" json:"target_location"`
	SourceX        *float64       `json:"source_x"`
	SourceY        *float64       `json:"source_y"`
	TargetX        *float64       `json:"target_x"`
	TargetY        *float64       `json:"target_y"`
	Description    string         `gorm:"type:text" json:"description"`
	ExpectedStart  *time.Time     `json:"expected_start"`
	ExpectedEnd    *time.Time     `json:"expected_end"`
	ActualStart    *time.Time     `json:"actual_start"`
	ActualEnd      *time.Time     `json:"actual_end"`
	ProgressPct    int            `gorm:"default:0" json:"progress_pct"`
	CreatedBy      *uint          `json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// ── 告警 ──
type Alert struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	AlertType      string    `gorm:"size:32" json:"alert_type"`
	Severity       string    `gorm:"size:16" json:"severity"`
	RobotID        *uint     `json:"robot_id"`
	Robot          *Robot    `gorm:"foreignKey:RobotID" json:"robot,omitempty"`
	TaskID         *uint     `json:"task_id"`
	Title          string    `gorm:"size:256" json:"title"`
	Content        string    `gorm:"type:text" json:"content"`
	Status         string    `gorm:"size:16;default:unack" json:"status"`
	AcknowledgedBy *uint     `json:"acknowledged_by"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// ── 遥测数据 (TimescaleDB hypertable) ──
type RobotTelemetry struct {
	Time           time.Time `gorm:"primaryKey" json:"time"`
	RobotID        uint      `gorm:"primaryKey" json:"robot_id"`
	BatteryPct     float64   `json:"battery_pct"`
	LocationX      float64   `json:"location_x"`
	LocationY      float64   `json:"location_y"`
	LocationTheta  float64   `json:"location_theta"`
	Speed          float64   `json:"speed"`
	LightMode      string    `json:"light_mode"`
	LightColor     string    `json:"light_color"`
	LightBrightness int      `json:"light_brightness"`
	Status         string    `json:"status"`
	CommStatus     string    `json:"comm_status"`
}

// ── 用户 ──
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:64" json:"username"`
	Password  string         `gorm:"size:256" json:"-"` // never expose
	Role      string         `gorm:"size:32;default:operator" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ── API 响应 ──
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func Success(data interface{}) Response {
	return Response{Code: 200, Message: "success", Data: data, Timestamp: time.Now().UnixMilli()}
}

func Error(code int, msg string) Response {
	return Response{Code: code, Message: msg, Timestamp: time.Now().UnixMilli()}
}

// ── 分页 ──
type Pagination struct {
	Page     int   `json:"page" form:"page"`
	PageSize int   `json:"page_size" form:"page_size"`
	Total    int64 `json:"total"`
}

func (p *Pagination) Default() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}
