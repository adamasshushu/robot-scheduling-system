package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/config"
	"github.com/adamasshushu/robot-scheduling-system/internal/handler"
	"github.com/adamasshushu/robot-scheduling-system/internal/middleware"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

func main() {
	cfg := config.Load()
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
	)
	log.Println("✅ Tables migrated")

	// 种子数据
	seedAdmin(db)
	seedRobots(db)

	// ── Router ──
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "robot-scheduling-system", "dev_mode": cfg.DevMode})
	})

	// ── API v1 ──
	v1 := r.Group("/api/v1")

	// 公开路由（开发模式免认证）
	robotH := handler.NewRobotHandler(db)
	taskH := handler.NewTaskHandler(db)

	if cfg.DevMode {
		// 开发模式：免 JWT 认证
		log.Println("⚠️  DEV mode — auth disabled")
		registerDevRoutes(v1, robotH, taskH)
	} else {
		// 生产模式：需要 JWT 认证
		auth := v1.Group("")
		auth.Use(middleware.AuthRequired())
		registerRoutes(auth, robotH, taskH)
	}

	// ── 启动 ──
	addr := fmt.Sprintf(":%s", cfg.Server.APIPort)
	log.Printf("🚀 API Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func registerRoutes(rg *gin.RouterGroup, robotH *handler.RobotHandler, taskH *handler.TaskHandler) {
	// 机器人管理
	rg.GET("/robots", robotH.List)
	rg.GET("/robots/:id", robotH.Get)
	rg.POST("/robots", robotH.Create)
	rg.PUT("/robots/:id", robotH.Update)
	rg.DELETE("/robots/:id", robotH.Delete)
	rg.POST("/robots/:id/commands", robotH.SendCommand)

	// 任务调度
	rg.GET("/tasks", taskH.List)
	rg.GET("/tasks/:id", taskH.Get)
	rg.POST("/tasks", taskH.Create)
	rg.POST("/tasks/:id/assign", taskH.Assign)
	rg.POST("/tasks/:id/cancel", taskH.Cancel)
}

func registerDevRoutes(rg *gin.RouterGroup, robotH *handler.RobotHandler, taskH *handler.TaskHandler) {
	registerRoutes(rg, robotH, taskH)
}

// seedAdmin 创建默认管理员
func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		db.Create(&model.User{
			Username: "admin",
			Password: "admin123",
			Role:     "admin",
		})
		log.Println("✅ Default admin created (admin / admin123)")
	}
}

// seedRobots 创建示例机器人
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
