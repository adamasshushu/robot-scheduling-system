package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/config"
	"github.com/adamasshushu/robot-scheduling-system/internal/handler"
	"github.com/adamasshushu/robot-scheduling-system/internal/middleware"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
	"github.com/adamasshushu/robot-scheduling-system/internal/service"
)

func main() {
	cfg := config.Load()

	// ── 生产模式 ──
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	middleware.InitJWT(cfg.JWT.Secret)

	// ── 数据库 ──
	var db *gorm.DB
	var err error

	if cfg.DevMode {
		log.Println("🔧 DEV mode — using SQLite")
		db, err = gorm.Open(sqlite.Open("rss_dev.db"), &gorm.Config{})
	} else {
		log.Println("🚀 PROD mode — using PostgreSQL")
		db, err = gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	log.Println("✅ Database connected")

	// Auto-migrate
	db.AutoMigrate(
		&model.User{},
		&model.Robot{},
		&model.Task{},
		&model.Alert{},
		&model.RobotTelemetry{},
		&model.MapConfig{},
		&model.RobotModel{},
		&model.OperationLog{},
		&model.WorkRecord{},
		&model.RobotSchedule{},
		&model.RobotMapBinding{},
		&model.FirmwareVersion{},
		&model.SystemConfig{},
	)
	log.Println("✅ Tables migrated")

	// 种子数据
	seedAdmin(db)
	seedRobots(db)

	// ── 初始化服务 ──
	alertEngine := service.NewAlertEngine(db)
	scheduler := service.NewScheduler(db)

	// 启动告警引擎（后台 10s 检查）
	alertEngine.Start()

	// ── MQTT 桥接 (机器人底盘接入) ──
	mqttBridge := service.NewMQTTBridge(db, service.Hub)

	// ── Router ──
	r := gin.New()
	r.Use(gin.Recovery())

	// 请求体大小限制 10MB
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
		c.Next()
	})

	// CORS — 生产模式限制来源
	allowedOrigins := parseOrigins(os.Getenv("CORS_ORIGINS"))
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:8080", "http://127.0.0.1:8080"}
	}
	if cfg.DevMode {
		allowedOrigins = []string{"*"}
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "robot-scheduling-system", "dev_mode": cfg.DevMode})
	})

	// WebSocket 实时推送
	r.GET("/api/v1/ws", handler.WSHandler)

	// ── API v1 ──
	v1 := r.Group("/api/v1")

	// 公开路由
	authH := handler.NewAuthHandler(db)
	v1.POST("/login", authH.Login)

	robotH := handler.NewRobotHandler(db)
	taskH := handler.NewTaskHandler(db, scheduler)
	alertH := handler.NewAlertHandler(db)
	reportH := handler.NewReportHandler(db)
	settingsH := handler.NewSettingsHandler(db)
	mapH := handler.NewMapHandler(db)
	centerH := handler.NewRobotCenterHandler(db)
	otaH := handler.NewOTAHandler(db)

	if cfg.DevMode {
		log.Println("⚠️  DEV mode — auth disabled")
		registerDevRoutes(v1, robotH, taskH, alertH, reportH, settingsH, mapH, centerH, otaH)
	} else {
		auth := v1.Group("")
		auth.Use(middleware.AuthRequired())
		registerRoutes(auth, robotH, taskH, alertH, reportH, settingsH, mapH, centerH, otaH)
	}

	// ── MQTT 机器人指令 ──
	v1.POST("/mqtt/command", func(c *gin.Context) {
		var cmd service.RobotCommand
		if c.ShouldBindJSON(&cmd) != nil || cmd.Command == "" {
			c.JSON(http.StatusBadRequest, model.Error(400, "invalid command"))
			return
		}
		robotCode := c.Query("robot_code")
		if robotCode == "" {
			c.JSON(http.StatusBadRequest, model.Error(400, "robot_code required"))
			return
		}
		mqttBridge.PublishCommand(robotCode, cmd)
		c.JSON(http.StatusOK, model.Success(gin.H{"robot_code": robotCode, "command": cmd.Command}))
	})

	// ── 启动 ──
	addr := fmt.Sprintf(":%s", cfg.Server.APIPort)
	log.Printf("🚀 API Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func parseOrigins(s string) []string {
	if s == "" || s == "*" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func registerRoutes(rg *gin.RouterGroup, robotH *handler.RobotHandler, taskH *handler.TaskHandler, alertH *handler.AlertHandler, reportH *handler.ReportHandler, settingsH *handler.SettingsHandler, mapH *handler.MapHandler, centerH *handler.RobotCenterHandler, otaH *handler.OTAHandler) {
	rg.GET("/robots", robotH.List)
	rg.GET("/robots/:id", robotH.Get)
	rg.POST("/robots", robotH.Create)
	rg.PUT("/robots/:id", robotH.Update)
	rg.DELETE("/robots/:id", robotH.Delete)
	rg.PUT("/robots/sort", robotH.SortRobots)
	rg.POST("/robots/:id/commands", robotH.SendCommand)
	rg.GET("/robots/:id/telemetry", reportH.RobotTelemetry)

	// 任务调度
	rg.GET("/tasks", taskH.List)
	rg.GET("/tasks/:id", taskH.Get)
	rg.POST("/tasks", taskH.Create)
	rg.PUT("/tasks/:id", taskH.Update)
	rg.DELETE("/tasks/:id", taskH.Delete)
	rg.POST("/tasks/:id/assign", taskH.Assign)
	rg.POST("/tasks/:id/cancel", taskH.Cancel)

	// 告警管理
	rg.GET("/alerts", alertH.List)
	rg.POST("/alerts/:id/ack", alertH.Ack)
	rg.POST("/alerts/:id/resolve", alertH.Resolve)
	rg.GET("/alerts/stats", alertH.Stats)

	// 监控 & 报表
	rg.GET("/monitor/dashboard", reportH.Dashboard)
	rg.GET("/monitor/robots/online", reportH.OnlineRobots)
	rg.GET("/monitor/map/robots", reportH.MapRobots)
	rg.GET("/reports/tasks", reportH.TaskStats)
	rg.GET("/reports/battery", reportH.BatteryTrend)
	rg.GET("/reports/task-types", reportH.TaskTypeDistribution)
	rg.GET("/reports/robot-usage", reportH.RobotUtilization)

	// 系统设置
	rg.GET("/settings/users", settingsH.ListUsers)
	rg.POST("/settings/users", settingsH.CreateUser)
	rg.PUT("/settings/users/:id", settingsH.UpdateUser)
	rg.DELETE("/settings/users/:id", settingsH.DeleteUser)
	rg.GET("/settings/models", settingsH.ListModels)
	rg.POST("/settings/models", settingsH.CreateModel)
	rg.DELETE("/settings/models/:id", settingsH.DeleteModel)

	// 地图管理
	rg.POST("/maps/upload", mapH.UploadMap)
	rg.GET("/maps", mapH.ListMaps)
	rg.GET("/maps/active", mapH.GetActiveMap)
	rg.PUT("/maps/:id/calibrate", mapH.CalibrateMap)
	rg.POST("/maps/:id/active", mapH.SetActiveMap)
	rg.DELETE("/maps/:id", mapH.DeleteMap)

	// ── 机器人管理中心 (模块1-7) ──
	rg.GET("/robots/:id/full", centerH.GetRobotFull)
	rg.PUT("/robots/:id/info", centerH.UpdateRobotInfo)
	rg.GET("/robots/:id/map-binding", centerH.GetRobotMap)
	rg.POST("/robots/:id/map-binding", centerH.BindRobotMap)
	rg.GET("/robots/:id/schedules", centerH.ListSchedules)
	rg.POST("/robots/:id/schedules", centerH.CreateSchedule)
	rg.PUT("/robots/:id/schedules/:scheduleId", centerH.UpdateSchedule)
	rg.DELETE("/robots/:id/schedules/:scheduleId", centerH.DeleteSchedule)
	rg.GET("/robots/:id/work-records", centerH.ListWorkRecords)
	rg.GET("/robots/:id/op-logs", centerH.ListOperationLogs)
	rg.POST("/robots/:id/relocate", centerH.Relocate)
	rg.PUT("/robots/:id/settings", centerH.UpdateRobotSettings)

	// ── OTA 远程升级 ──
	rg.POST("/ota/upload", otaH.UploadFirmware)
	rg.GET("/ota/firmwares", otaH.ListFirmwares)
	rg.PUT("/ota/firmwares/:id", otaH.ToggleAutoUpgrade)
	rg.POST("/ota/upgrade/:robotId", otaH.UpgradeRobot)

	// ── 系统配置 ──
	rg.GET("/system/config", otaH.GetSystemConfig)
	rg.PUT("/system/config", otaH.UpdateSystemConfig)
}

func registerDevRoutes(rg *gin.RouterGroup, robotH *handler.RobotHandler, taskH *handler.TaskHandler, alertH *handler.AlertHandler, reportH *handler.ReportHandler, settingsH *handler.SettingsHandler, mapH *handler.MapHandler, centerH *handler.RobotCenterHandler, otaH *handler.OTAHandler) {
	registerRoutes(rg, robotH, taskH, alertH, reportH, settingsH, mapH, centerH, otaH)
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		admin := model.User{
			Username: "admin",
			Role:     "admin",
		}
		admin.SetPassword("admin123")
		db.Create(&admin)
		log.Println("✅ Default admin created (admin / admin123)")
	}
}

func seedRobots(db *gorm.DB) {
	var count int64
	db.Model(&model.Robot{}).Count(&count)
	if count == 0 {
		robots := []model.Robot{
			{RobotCode: "RBT-001", Model: "AGV-2000", Name: "搬运一号", LocationX: 10.0, LocationY: 20.0, MapZone: "A区-仓库", Status: "standby"},
			{RobotCode: "RBT-002", Model: "AGV-2000", Name: "搬运二号", LocationX: 50.0, LocationY: 30.0, MapZone: "B区-装配线", Status: "standby"},
			{RobotCode: "RBT-003", Model: "AMR-100", Name: "巡逻一号", LocationX: 100.0, LocationY: 80.0, MapZone: "C区-办公区", Status: "standby"},
		}
		for _, r := range robots {
			db.Create(&r)
		}
		log.Printf("✅ %d demo robots created", len(robots))
	}
}
