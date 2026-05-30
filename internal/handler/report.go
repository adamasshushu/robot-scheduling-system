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
func (h *ReportHandler) RobotTelemetry(c *gin.Context) {
	id := c.Param("id")
	hours := c.DefaultQuery("hours", "24")
	since := time.Now().Add(-parseDuration(hours))
	var telemetry []model.RobotTelemetry
	h.db.Where("robot_id = ? AND time > ?", id, since).
		Order("time ASC").Limit(1000).Find(&telemetry)
	c.JSON(http.StatusOK, model.Success(telemetry))
}

// OnlineRobots 在线机器人
func (h *ReportHandler) OnlineRobots(c *gin.Context) {
	var robots []model.Robot
	h.db.Where("comm_status = ?", "online").Find(&robots)
	c.JSON(http.StatusOK, model.Success(robots))
}

// MapRobots 地图机器人位置
func (h *ReportHandler) MapRobots(c *gin.Context) {
	var robots []model.Robot
	h.db.Select("id, robot_code, name, location_x, location_y, location_theta, map_zone, status, battery_pct, light_color").
		Where("comm_status = ?", "online").Find(&robots)
	c.JSON(http.StatusOK, model.Success(robots))
}

// TaskStats 任务统计趋势
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
		FROM tasks WHERE created_at > ? GROUP BY DATE(created_at) ORDER BY date
	`, since).Scan(&stats)
	c.JSON(http.StatusOK, model.Success(stats))
}

// BatteryTrend 电量历史趋势 (所有机器人)
func (h *ReportHandler) BatteryTrend(c *gin.Context) {
	hours := c.DefaultQuery("hours", "24")
	since := time.Now().Add(-parseDuration(hours))

	type BatteryPoint struct {
		Time       string  `json:"time"`
		RobotID    uint    `json:"robot_id"`
		RobotName  string  `json:"robot_name"`
		BatteryPct float64 `json:"battery_pct"`
	}
	var points []BatteryPoint
	h.db.Raw(`
		SELECT t.time, t.robot_id, r.name as robot_name, t.battery_pct
		FROM robot_telemetries t
		JOIN robots r ON r.id = t.robot_id
		WHERE t.time > ?
		ORDER BY t.time ASC LIMIT 2000
	`, since).Scan(&points)
	c.JSON(http.StatusOK, model.Success(points))
}

// TaskTypeDistribution 任务类型分布
func (h *ReportHandler) TaskTypeDistribution(c *gin.Context) {
	type TypeCount struct {
		TaskType string `json:"task_type"`
		Count    int64  `json:"count"`
	}
	var dist []TypeCount
	h.db.Raw(`
		SELECT task_type, COUNT(*) as count FROM tasks
		WHERE created_at > ? GROUP BY task_type ORDER BY count DESC
	`, time.Now().AddDate(0, 0, -30)).Scan(&dist)
	c.JSON(http.StatusOK, model.Success(dist))
}

// RobotUtilization 机器人利用率 (任务数排名)
func (h *ReportHandler) RobotUtilization(c *gin.Context) {
	type RobotUsage struct {
		RobotID   uint   `json:"robot_id"`
		RobotName string `json:"robot_name"`
		TaskCount int64  `json:"task_count"`
	}
	var usage []RobotUsage
	h.db.Raw(`
		SELECT t.robot_id, r.name as robot_name, COUNT(*) as task_count
		FROM tasks t JOIN robots r ON r.id = t.robot_id
		WHERE t.created_at > ? AND t.robot_id IS NOT NULL
		GROUP BY t.robot_id, r.name ORDER BY task_count DESC LIMIT 10
	`, time.Now().AddDate(0, 0, -30)).Scan(&usage)
	c.JSON(http.StatusOK, model.Success(usage))
}

func completionRate(completed, total int64) float64 {
	if total == 0 { return 0 }
	return float64(completed) / float64(total) * 100
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	if d == 0 { d, _ = time.ParseDuration(s + "h") }
	return d
}
