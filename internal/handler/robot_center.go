package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

// RobotCenterHandler 机器人管理中心 (模块1-7)
type RobotCenterHandler struct {
	db *gorm.DB
}

func NewRobotCenterHandler(db *gorm.DB) *RobotCenterHandler {
	return &RobotCenterHandler{db: db}
}

// ── 模块1: 机器人信息 (读取/更新扩展字段) ──

// GetRobotFull 获取机器人完整信息
func (h *RobotCenterHandler) GetRobotFull(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var robot model.Robot
	if h.db.First(&robot, id).Error != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "robot not found"))
		return
	}
	c.JSON(http.StatusOK, model.Success(robot))
}

// UpdateRobotInfo 更新机器人软硬件信息
func (h *RobotCenterHandler) UpdateRobotInfo(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var robot model.Robot
	if h.db.First(&robot, id).Error != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "robot not found"))
		return
	}
	var input map[string]interface{}
	c.ShouldBindJSON(&input)
	// 白名单：只允许更新信息字段
	allowed := map[string]bool{
		"name": true, "serial_number": true, "image_url": true,
		"sw_frontend_ver": true, "sw_backend_ver": true, "sw_firmware_ver": true, "sw_nav_ver": true,
		"hw_model": true, "hw_dimensions": true, "hw_weight": true, "hw_payload": true,
		"hw_temp_range": true, "hw_charge_method": true,
	}
	updates := make(map[string]interface{})
	for k, v := range input {
		if allowed[k] {
			updates[k] = v
		}
	}
	h.db.Model(&robot).Updates(updates)
	c.JSON(http.StatusOK, model.Success(robot))
}

// ── 模块2: 机器人-地图绑定 ──

// BindRobotMap 绑定/更新机器人地图
func (h *RobotCenterHandler) BindRobotMap(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input model.RobotMapBinding
	c.ShouldBindJSON(&input)
	input.RobotID = uint(robotID)

	// upsert: 一个机器人一张地图
	var existing model.RobotMapBinding
	h.db.Where("robot_id = ?", robotID).First(&existing)
	if existing.ID != 0 {
		h.db.Model(&existing).Updates(map[string]interface{}{
			"map_id": input.MapID, "map_name": input.MapName,
			"is_auto": input.IsAuto, "pos_x": input.PosX, "pos_y": input.PosY, "pos_theta": input.PosTheta,
			"updated_at": time.Now(),
		})
	} else {
		h.db.Create(&input)
	}
	c.JSON(http.StatusOK, model.Success(input))
}

// GetRobotMap 获取机器人当前绑定地图
func (h *RobotCenterHandler) GetRobotMap(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var binding model.RobotMapBinding
	h.db.Where("robot_id = ?", robotID).First(&binding)
	c.JSON(http.StatusOK, model.Success(binding))
}

// ── 模块3: 定时任务 ──

// ListSchedules 机器人定时任务列表
func (h *RobotCenterHandler) ListSchedules(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var schedules []model.RobotSchedule
	h.db.Where("robot_id = ?", robotID).Order("id DESC").Find(&schedules)
	c.JSON(http.StatusOK, model.Success(schedules))
}

// CreateSchedule 创建定时任务
func (h *RobotCenterHandler) CreateSchedule(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var s model.RobotSchedule
	c.ShouldBindJSON(&s)
	s.RobotID = uint(robotID)
	h.db.Create(&s)
	c.JSON(http.StatusOK, model.Success(s))
}

// UpdateSchedule 更新定时任务 (含开关)
func (h *RobotCenterHandler) UpdateSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("scheduleId"), 10, 64)
	var input map[string]interface{}
	c.ShouldBindJSON(&input)
	h.db.Model(&model.RobotSchedule{}).Where("id = ?", id).Updates(input)
	c.JSON(http.StatusOK, model.Success(nil))
}

// DeleteSchedule 删除定时任务
func (h *RobotCenterHandler) DeleteSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("scheduleId"), 10, 64)
	h.db.Delete(&model.RobotSchedule{}, id)
	c.JSON(http.StatusOK, model.Success(nil))
}

// ── 模块4: 工作记录 ──

// ListWorkRecords 工作记录列表
func (h *RobotCenterHandler) ListWorkRecords(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var records []model.WorkRecord
	h.db.Where("robot_id = ?", robotID).Order("start_time DESC").Limit(200).Find(&records)
	c.JSON(http.StatusOK, model.Success(records))
}

// ── 模块5: 操作日志 ──

// ListOperationLogs 操作日志
func (h *RobotCenterHandler) ListOperationLogs(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	logType := c.DefaultQuery("type", "") // manual/remote/alert
	query := h.db.Where("robot_id = ?", robotID)
	if logType != "" {
		query = query.Where("log_type = ?", logType)
	}
	var logs []model.OperationLog
	query.Order("created_at DESC").Limit(500).Find(&logs)
	c.JSON(http.StatusOK, model.Success(logs))
}

// LogOperation 记录操作 (内部调用)
func (h *RobotCenterHandler) LogOperation(robotID uint, logType, operator, action, detail string) {
	h.db.Create(&model.OperationLog{
		RobotID: robotID, LogType: logType, Operator: operator, Action: action, Detail: detail,
		CreatedAt: time.Now(),
	})
}

// ── 模块6: 重定位 ──

// Relocate 手动/自动重定位
func (h *RobotCenterHandler) Relocate(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		Mode  string  `json:"mode"`  // auto / manual
		PosX  float64 `json:"pos_x"`
		PosY  float64 `json:"pos_y"`
		Theta float64 `json:"theta"`
	}
	c.ShouldBindJSON(&input)

	username, _ := c.Get("username")
	op := "system"
	if u, ok := username.(string); ok { op = u }

	if input.Mode == "manual" {
		h.db.Model(&model.Robot{}).Where("id = ?", robotID).Updates(map[string]interface{}{
			"location_x": input.PosX, "location_y": input.PosY, "location_theta": input.Theta,
		})
		h.LogOperation(uint(robotID), "manual", op, "手动重定位", "设置位置")
		c.JSON(http.StatusOK, model.Success(gin.H{"status": "relocated", "x": input.PosX, "y": input.PosY}))
	} else {
		// 自动重定位 — 通知机器人执行
		h.LogOperation(uint(robotID), "manual", op, "自动重定位", "下发定位指令")
		c.JSON(http.StatusOK, model.Success(gin.H{"status": "relocating_auto"}))
	}
}

// ── 模块7: 机器人设置 ──

// UpdateRobotSettings 更新机器人设置 (音量/速度/电量阈值/锁屏)
func (h *RobotCenterHandler) UpdateRobotSettings(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		VoiceVolume     *int     `json:"voice_volume"`
		MaxSpeedSetting *float64 `json:"max_speed_setting"`
		BatteryAlertPct *float64 `json:"battery_alert_pct"`
		AutoChargePct   *float64 `json:"auto_charge_pct"`
		ScreenLockPwd   string   `json:"screen_lock_pwd"`
		Locked          *bool    `json:"locked"`
	}
	c.ShouldBindJSON(&input)

	updates := make(map[string]interface{})
	if input.VoiceVolume != nil { updates["voice_volume"] = *input.VoiceVolume }
	if input.MaxSpeedSetting != nil { updates["max_speed_setting"] = *input.MaxSpeedSetting }
	if input.BatteryAlertPct != nil { updates["battery_alert_pct"] = *input.BatteryAlertPct }
	if input.AutoChargePct != nil { updates["auto_charge_pct"] = *input.AutoChargePct }
	if input.Locked != nil { updates["locked"] = *input.Locked }
	if input.ScreenLockPwd != "" {
		hash, _ := model.HashPassword(input.ScreenLockPwd)
		updates["screen_lock_pwd"] = hash
	}

	h.db.Model(&model.Robot{}).Where("id = ?", robotID).Updates(updates)

	username, _ := c.Get("username")
	op := "system"
	if u, ok := username.(string); ok { op = u }
	h.LogOperation(uint(robotID), "manual", op, "更新设置", "机器人参数调整")

	c.JSON(http.StatusOK, model.Success(nil))
}
