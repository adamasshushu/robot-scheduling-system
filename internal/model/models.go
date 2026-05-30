package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// bcrypt helpers
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func bcryptHash(password string) (string, error) {
	return HashPassword(password)
}

func bcryptCheck(hash, password string) bool {
	return CheckPasswordHash(hash, password)
}

// ── 机器人 ──
type Robot struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	RobotCode      string         `gorm:"uniqueIndex;size:32" json:"robot_code"`
	SerialNumber   string         `gorm:"size:64" json:"serial_number"`     // SN 码
	Model          string         `gorm:"size:64" json:"model"`
	Name           string         `gorm:"size:128" json:"name"`
	SortOrder      int            `json:"sort_order" gorm:"default:0"`
	ImageURL       string         `gorm:"size:512" json:"image_url"`        // 设备图片

	// 软件版本
	SwFrontendVer string `gorm:"size:32" json:"sw_frontend_ver"`
	SwBackendVer  string `gorm:"size:32" json:"sw_backend_ver"`
	SwFirmwareVer string `gorm:"size:32" json:"sw_firmware_ver"` // 下位机
	SwNavVer      string `gorm:"size:32" json:"sw_nav_ver"`      // 导航系统

	// 硬件信息
	HwModel        string  `gorm:"size:64" json:"hw_model"`
	HwDimensions   string  `gorm:"size:64" json:"hw_dimensions"`   // 产品尺寸
	HwWeight       float64 `gorm:"type:decimal(8,2)" json:"hw_weight"`     // 整机重量 kg
	HwPayload      float64 `gorm:"type:decimal(8,2)" json:"hw_payload"`    // 最大载重 kg
	HwTempRange    string  `gorm:"size:32" json:"hw_temp_range"`           // 工作温度
	HwChargeMethod string  `gorm:"size:64" json:"hw_charge_method"`        // 充电方式

	// 运行时状态
	BatteryPct     float64   `gorm:"type:decimal(5,2);default:100" json:"battery_pct"`
	BatteryStatus  string    `gorm:"size:16;default:normal" json:"battery_status"`
	LocationX      float64   `gorm:"type:decimal(10,3)" json:"location_x"`
	LocationY      float64   `gorm:"type:decimal(10,3)" json:"location_y"`
	LocationTheta  float64   `gorm:"type:decimal(8,3)" json:"location_theta"`
	MapZone        string    `gorm:"size:64" json:"map_zone"`
	Speed          float64   `gorm:"type:decimal(6,2);default:0" json:"speed"`
	LightMode      string    `gorm:"size:16;default:off" json:"light_mode"`
	LightColor     string    `gorm:"size:16;default:white" json:"light_color"`
	LightBrightness int      `gorm:"default:100" json:"light_brightness"`
	Status         string    `gorm:"size:16;default:standby" json:"status"`
	CommStatus     string    `gorm:"size:16;default:online" json:"comm_status"`
	LastHeartbeat  *time.Time `json:"last_heartbeat"`
	IPAddress      *string   `gorm:"type:inet" json:"ip_address"`
	FirmwareVer    string    `gorm:"size:32" json:"firmware_ver"`
	MaxPayload     float64   `gorm:"type:decimal(8,2)" json:"max_payload"`
	MaxSpeed       float64   `gorm:"type:decimal(6,2)" json:"max_speed"`

	// 机器人设置
	VoiceVolume     int     `gorm:"default:80" json:"voice_volume"`
	MaxSpeedSetting float64 `gorm:"type:decimal(6,2)" json:"max_speed_setting"`
	BatteryAlertPct float64 `gorm:"type:decimal(5,2);default:20" json:"battery_alert_pct"`
	AutoChargePct   float64 `gorm:"type:decimal(5,2);default:15" json:"auto_charge_pct"`
	ScreenLockPwd   string  `gorm:"size:256" json:"-"` // bcrypt hash
	Locked          bool    `gorm:"default:false" json:"locked"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ── 操作日志 ──
type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RobotID   uint      `gorm:"index" json:"robot_id"`
	Robot     *Robot    `gorm:"foreignKey:RobotID" json:"robot,omitempty"`
	LogType   string    `gorm:"size:32" json:"log_type"`   // manual / remote / alert
	Operator  string    `gorm:"size:64" json:"operator"`    // 操作用户名
	Action    string    `gorm:"size:256" json:"action"`     // 操作描述
	Detail    string    `gorm:"type:text" json:"detail"`    // 详情
	CreatedAt time.Time `json:"created_at"`
}

// ── 工作记录 ──
type WorkRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RobotID   uint      `gorm:"index" json:"robot_id"`
	Robot     *Robot    `gorm:"foreignKey:RobotID" json:"robot,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	TaskCode  string    `gorm:"size:32" json:"task_code"`
	TaskDesc  string    `gorm:"size:256" json:"task_desc"`
	PathData  string    `gorm:"type:text" json:"path_data"`  // 行走线路 JSON
	CreatedAt time.Time `json:"created_at"`
}

// ── 机器人定时任务 ──
type RobotSchedule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RobotID   uint      `gorm:"index" json:"robot_id"`
	Robot     *Robot    `gorm:"foreignKey:RobotID" json:"robot,omitempty"`
	Name      string    `gorm:"size:128" json:"name"`
	CronExpr  string    `gorm:"size:64" json:"cron_expr"`   // cron 表达式
	TaskType  string    `gorm:"size:32" json:"task_type"`
	Params    string    `gorm:"type:text" json:"params"`     // 任务参数 JSON
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── 机器人-地图绑定 ──
type RobotMapBinding struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RobotID   uint      `gorm:"index" json:"robot_id"`
	MapID     uint      `json:"map_id"`
	MapName   string    `gorm:"size:128" json:"map_name"`
	IsAuto    bool      `gorm:"default:true" json:"is_auto"` // 自动定位
	PosX      float64   `json:"pos_x"`
	PosY      float64   `json:"pos_y"`
	PosTheta  float64   `json:"pos_theta"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── OTA 固件版本 ──
type FirmwareVersion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Module      string    `gorm:"size:32" json:"module"`       // frontend/backend/firmware/nav
	Version     string    `gorm:"size:32" json:"version"`
	FileName    string    `gorm:"size:256" json:"file_name"`
	FileURL     string    `gorm:"size:512" json:"file_url"`    // MinIO 下载链接
	FileSize    int64     `json:"file_size"`
	MD5         string    `gorm:"size:64" json:"md5"`
	ChangeLog   string    `gorm:"type:text" json:"change_log"`
	AutoUpgrade bool      `gorm:"default:false" json:"auto_upgrade"` // 是否自动推送
	CreatedAt   time.Time `json:"created_at"`
}

// ── 系统配置 ──
type SystemConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ConfigKey    string    `gorm:"uniqueIndex;size:64" json:"config_key"`
	ConfigValue  string    `gorm:"type:text" json:"config_value"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	RobotCode      string    `gorm:"size:64" json:"robot_code"`
	Robot          *Robot    `gorm:"foreignKey:RobotID" json:"robot,omitempty"`
	TaskID         *uint     `json:"task_id"`
	Title          string    `gorm:"size:256" json:"title"`
	Content        string    `gorm:"type:text" json:"content"`
	Level          string    `gorm:"size:16" json:"level"`     // info/warning/error/critical
	Type           string    `gorm:"size:32" json:"type"`      // battery/collision/sensor/motor/system
	Message        string    `gorm:"size:512" json:"message"`
	Data           string    `gorm:"type:text" json:"data"`
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
	Password  string         `gorm:"size:256" json:"-"` // bcrypt hash, never expose
	Role      string         `gorm:"size:32;default:operator" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// SetPassword 设置 bcrypt 哈希密码
func (u *User) SetPassword(raw string) error {
	hash, err := bcryptHash(raw)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(raw string) bool {
	return bcryptCheck(u.Password, raw)
}

// ── 机器人型号 ──
type RobotModel struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Name       string  `gorm:"uniqueIndex;size:64" json:"name"`
	Maker      string  `gorm:"size:64" json:"maker"`
	MaxPayload float64 `json:"max_payload"`
	MaxSpeed   float64 `json:"max_speed"`
	ImageURL   string  `gorm:"size:512" json:"image_url"`
	Specs      string  `gorm:"type:text" json:"specs"`
}

// ── 地图配置 ──
type MapConfig struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Name         string  `gorm:"size:128" json:"name"`           // 地图名称
	ImageURL     string  `gorm:"size:512" json:"image_url"`      // MinIO 图片地址
	ImageWidth   float64 `json:"image_width"`                    // 图片宽度(px)
	ImageHeight  float64 `json:"image_height"`                   // 图片高度(px)
	RealWidth    float64 `json:"real_width"`                     // 对应现实宽度(m)
	RealHeight   float64 `json:"real_height"`                    // 对应现实高度(m)
	OriginX      float64 `gorm:"default:0" json:"origin_x"`      // 图原点对应现实X
	OriginY      float64 `gorm:"default:0" json:"origin_y"`      // 图原点对应现实Y
	CalibPoint1X float64 `json:"calib_point1_x"`                 // 标定点1 图片X
	CalibPoint1Y float64 `json:"calib_point1_y"`                 // 标定点1 图片Y
	CalibPoint1RX float64 `json:"calib_point1_rx"`               // 标定点1 现实X
	CalibPoint1RY float64 `json:"calib_point1_ry"`               // 标定点1 现实Y
	CalibPoint2X float64 `json:"calib_point2_x"`                 // 标定点2 图片X
	CalibPoint2Y float64 `json:"calib_point2_y"`                 // 标定点2 图片Y
	CalibPoint2RX float64 `json:"calib_point2_rx"`               // 标定点2 现实X
	CalibPoint2RY float64 `json:"calib_point2_ry"`               // 标定点2 现实Y
	Active       bool    `gorm:"default:true" json:"active"`     // 当前使用
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
