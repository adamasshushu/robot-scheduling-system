package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

type OTAHandler struct {
	db    *gorm.DB
	minio *minio.Client
}

func NewOTAHandler(db *gorm.DB) *OTAHandler {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" { endpoint = "localhost:9000" }
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" { accessKey = "minioadmin" }
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" { secretKey = "minioadmin" }

	client, _ := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	return &OTAHandler{db: db, minio: client}
}

// ── 固件上传 ──
// POST /api/v1/ota/upload
func (h *OTAHandler) UploadFirmware(c *gin.Context) {
	if h.minio == nil {
		c.JSON(http.StatusServiceUnavailable, model.Error(503, "storage not available"))
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "no file"))
		return
	}
	defer file.Close()

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".bin": true, ".gz": true, ".zip": true, ".tar": true, ".img": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, model.Error(400, "unsupported format, allowed: bin/gz/zip/tar/img"))
		return
	}

	objName := fmt.Sprintf("ota/%d_%s", time.Now().UnixMilli(), header.Filename)
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" { bucket = "robot-assets" }

	_, err = h.minio.PutObject(c, bucket, objName, file, header.Size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Error(500, "upload failed"))
		return
	}

	fwURL := fmt.Sprintf("http://localhost:9000/%s/%s", bucket, objName)
	fw := model.FirmwareVersion{
		Module:      c.PostForm("module"),
		Version:     c.PostForm("version"),
		FileName:    header.Filename,
		FileURL:     fwURL,
		FileSize:    header.Size,
		MD5:         c.PostForm("md5"),
		ChangeLog:   c.PostForm("change_log"),
		AutoUpgrade: c.PostForm("auto_upgrade") == "true",
		CreatedAt:   time.Now(),
	}
	h.db.Create(&fw)
	c.JSON(http.StatusOK, model.Success(fw))
}

// ── 固件列表 ──
// GET /api/v1/ota/firmwares
func (h *OTAHandler) ListFirmwares(c *gin.Context) {
	module := c.DefaultQuery("module", "")
	var fws []model.FirmwareVersion
	query := h.db.Order("created_at DESC")
	if module != "" {
		query = query.Where("module = ?", module)
	}
	query.Find(&fws)
	c.JSON(http.StatusOK, model.Success(fws))
}

// ── 推送升级到指定机器人 ──
// POST /api/v1/ota/upgrade/:robotId
func (h *OTAHandler) UpgradeRobot(c *gin.Context) {
	robotID, _ := strconv.ParseUint(c.Param("robotId"), 10, 64)
	var input struct {
		FirmwareID uint   `json:"firmware_id" binding:"required"`
		Mode       string `json:"mode"` // auto / manual
	}
	c.ShouldBindJSON(&input)

	var fw model.FirmwareVersion
	if h.db.First(&fw, input.FirmwareID).Error != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "firmware not found"))
		return
	}
	var robot model.Robot
	if h.db.First(&robot, robotID).Error != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "robot not found"))
		return
	}

	// 更新机器人对应模块版本
	verUpdate := map[string]interface{}{}
	switch fw.Module {
	case "frontend": verUpdate["sw_frontend_ver"] = fw.Version
	case "backend": verUpdate["sw_backend_ver"] = fw.Version
	case "firmware": verUpdate["sw_firmware_ver"] = fw.Version
	case "nav": verUpdate["sw_nav_ver"] = fw.Version
	}
	h.db.Model(&robot).Updates(verUpdate)

	username, _ := c.Get("username")
	op := "system"
	if u, ok := username.(string); ok { op = u }
	LogOp(h.db, uint(robotID), "manual", op, fmt.Sprintf("OTA %s: %s → %s", fw.Module, fw.Version, input.Mode))

	c.JSON(http.StatusOK, model.Success(gin.H{
		"robot_id":    robotID,
		"module":      fw.Module,
		"version":     fw.Version,
		"mode":        input.Mode,
		"status":      "queued",
	}))
}

// ── 自动推送开关 ──
// PUT /api/v1/ota/firmwares/:id
func (h *OTAHandler) ToggleAutoUpgrade(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		AutoUpgrade bool `json:"auto_upgrade"`
	}
	c.ShouldBindJSON(&input)
	h.db.Model(&model.FirmwareVersion{}).Where("id = ?", id).Update("auto_upgrade", input.AutoUpgrade)
	c.JSON(http.StatusOK, model.Success(nil))
}

// ── 系统配置 (Logo/名称) ──
// GET /api/v1/system/config
func (h *OTAHandler) GetSystemConfig(c *gin.Context) {
	var configs []model.SystemConfig
	h.db.Find(&configs)
	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.ConfigKey] = cfg.ConfigValue
	}
	c.JSON(http.StatusOK, model.Success(result))
}

// PUT /api/v1/system/config
func (h *OTAHandler) UpdateSystemConfig(c *gin.Context) {
	var input map[string]string
	c.ShouldBindJSON(&input)
	for k, v := range input {
		var cfg model.SystemConfig
		h.db.Where("config_key = ?", k).First(&cfg)
		if cfg.ID != 0 {
			h.db.Model(&cfg).Updates(map[string]interface{}{"config_value": v, "updated_at": time.Now()})
		} else {
			h.db.Create(&model.SystemConfig{ConfigKey: k, ConfigValue: v, UpdatedAt: time.Now()})
		}
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// 辅助：记录操作日志
func LogOp(db *gorm.DB, robotID uint, logType, operator, action string) {
	if db != nil {
		db.Create(&model.OperationLog{
			RobotID:   robotID,
			LogType:   logType,
			Operator:  operator,
			Action:    action,
			CreatedAt: time.Now(),
		})
	}
}

// Ensure io is used
var _ = io.Discard
