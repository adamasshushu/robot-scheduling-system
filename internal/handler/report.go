package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

type ReportHandler struct {
	db *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// Dashboard 仪表盘数据
// GET /api/v1/monitor/dashboard
func (h *ReportHandler) Dashboard(c *gin.Context) {
	var onlineRobots, totalRobots, todayTasks, completedTasks, unackAlerts int64
	today := time.Now().Format("2006-01-02")

	h.db.Model(&model.Robot{}).Where("comm_status = ?", "online").Count(&onlineRobots)
	h.db.Model(&model.Robot{}).Count(&totalRobots)
	h.db.Model(&model.Task{}).Where("DATE(created_at) = ?", today).Count(&todayTasks)
	h.db.Model(&model.Task{}).Where("DATE(created_at) = ? AND status = ?", today, "completed").Count(&completedTasks)
	h.db.Model(&model.Alert{}).Where("status = ?", "unack").Count(&unackAlerts)

	c.JSON(http.StatusOK, model.Success(gin.H{
		"online_robots":   onlineRobots,
		"total_robots":    totalRobots,
		"today_tasks":     todayTasks,
		"completed_tasks": completedTasks,
		"unack_alerts":    unackAlerts,
		"completion_rate": completionRate(completedTasks, todayTasks),
	}))
}

// RobotTelemetry 机器人遥测历史
// GET /api/v1/robots/:id/telemetry?hours=24
func (h *ReportHandler) RobotTelemetry(c *gin.Context) {
	id := c.Param("id")
	hours := c.DefaultQuery("hours", "24")

	since := time.Now().Add(-parseDuration(hours))

	var telemetry []model.RobotTelemetry
	h.db.Where("robot_id = ? AND time > ?", id, since).
		Order("time ASC").Limit(1000).Find(&telemetry)

	c.JSON(http.StatusOK, model.Success(telemetry))
}

// OnlineRobots 在线机器人实时状态
// GET /api/v1/monitor/robots/online
func (h *ReportHandler) OnlineRobots(c *gin.Context) {
	var robots []model.Robot
	h.db.Where("comm_status = ?", "online").Find(&robots)
	c.JSON(http.StatusOK, model.Success(robots))
}

// MapRobots 地图上的机器人位置
// GET /api/v1/monitor/map/robots
func (h *ReportHandler) MapRobots(c *gin.Context) {
	var robots []model.Robot
	h.db.Select("id, robot_code, name, location_x, location_y, location_theta, map_zone, status, battery_pct, light_color").
		Where("comm_status = ?", "online").
		Find(&robots)
	c.JSON(http.StatusOK, model.Success(robots))
}

// TaskStats 任务统计
// GET /api/v1/reports/tasks?days=7
func (h *ReportHandler) TaskStats(c *gin.Context) {
	days := c.DefaultQuery("days", "7")
	since := time.Now().Add(-parseDuration(days + "d"))

	type DailyStat struct {
		Date      string `json:"date"`
		Total     int64  `json:"total"`
		Completed int64  `json:"completed"`
	}

	var stats []DailyStat
	h.db.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as total,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM tasks
		WHERE created_at > ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, since).Scan(&stats)

	c.JSON(http.StatusOK, model.Success(stats))
}

func completionRate(completed, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total) * 100
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	// 如果没有单位，按小时算
	if d == 0 {
		d, _ = time.ParseDuration(s + "h")
	}
	return d
}
