package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
	"gorm.io/gorm"
)

type MapHandler struct {
	db         *gorm.DB
	minio      *minio.Client
	bucketName string
}

func NewMapHandler(db *gorm.DB) *MapHandler {
	// MinIO 连接
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "rss-maps"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Printf("⚠️  MinIO connect failed: %v — map upload disabled", err)
		return &MapHandler{db: db, bucketName: bucket}
	}

	// 确保 bucket 存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err == nil && !exists {
		client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		log.Printf("📦 MinIO bucket created: %s", bucket)
	}

	log.Println("✅ MinIO connected — map storage ready")
	return &MapHandler{db: db, minio: client, bucketName: bucket}
}

// UploadMap 上传工厂平面图
// POST /api/v1/maps/upload
func (h *MapHandler) UploadMap(c *gin.Context) {
	if h.minio == nil {
		c.JSON(http.StatusServiceUnavailable, model.Error(503, "MinIO not available"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "no file uploaded"))
		return
	}
	defer file.Close()

	// ── 文件类型校验 ──
	ext := filepath.Ext(strings.ToLower(header.Filename))
	allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true, ".bmp": true, ".pdf": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, model.Error(400, "unsupported file type, allowed: png/jpg/gif/webp/svg/bmp/pdf"))
		return
	}

	// 读取前 512 字节检测真实 MIME 类型
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if !strings.HasPrefix(contentType, "image/") && contentType != "application/pdf" {
		c.JSON(http.StatusBadRequest, model.Error(400, "file content is not a valid image or pdf"))
		return
	}

	// 重置读取位置
	file.Seek(0, 0)

	objectName := fmt.Sprintf("maps/%d_%s%s", time.Now().UnixMilli(), sanitizeName(header.Filename), ext)

	ctx := context.Background()
	_, err = h.minio.PutObject(ctx, h.bucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		log.Printf("❌ MinIO upload failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Error(500, "upload failed"))
		return
	}

	// 生成公开 URL
	imageURL := fmt.Sprintf("http://%s/%s/%s", os.Getenv("MINIO_PUBLIC_ENDPOINT"), h.bucketName, objectName)
	if os.Getenv("MINIO_PUBLIC_ENDPOINT") == "" {
		imageURL = fmt.Sprintf("http://localhost:9000/%s/%s", h.bucketName, objectName)
	}

	mapCfg := model.MapConfig{
		Name:   c.PostForm("name"),
		ImageURL: imageURL,
	}
	if mapCfg.Name == "" {
		mapCfg.Name = header.Filename
	}
	h.db.Create(&mapCfg)

	c.JSON(http.StatusOK, model.Success(mapCfg))
}

// ListMaps 地图列表
func (h *MapHandler) ListMaps(c *gin.Context) {
	var maps []model.MapConfig
	h.db.Order("id DESC").Find(&maps)
	c.JSON(http.StatusOK, model.Success(maps))
}

// CalibrateMap 坐标校准
// PUT /api/v1/maps/:id/calibrate
func (h *MapHandler) CalibrateMap(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var input struct {
		ImageWidth   float64 `json:"image_width"`
		ImageHeight  float64 `json:"image_height"`
		RealWidth    float64 `json:"real_width"`
		RealHeight   float64 `json:"real_height"`
		CalibPoint1X float64 `json:"calib_point1_x"`
		CalibPoint1Y float64 `json:"calib_point1_y"`
		CalibPoint1RX float64 `json:"calib_point1_rx"`
		CalibPoint1RY float64 `json:"calib_point1_ry"`
		CalibPoint2X float64 `json:"calib_point2_x"`
		CalibPoint2Y float64 `json:"calib_point2_y"`
		CalibPoint2RX float64 `json:"calib_point2_rx"`
		CalibPoint2RY float64 `json:"calib_point2_ry"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "invalid data"))
		return
	}

	var m model.MapConfig
	if h.db.First(&m, id).Error != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "map not found"))
		return
	}

	h.db.Model(&m).Updates(map[string]interface{}{
		"image_width":    input.ImageWidth,
		"image_height":   input.ImageHeight,
		"real_width":     input.RealWidth,
		"real_height":    input.RealHeight,
		"calib_point1_x":  input.CalibPoint1X,
		"calib_point1_y":  input.CalibPoint1Y,
		"calib_point1_rx": input.CalibPoint1RX,
		"calib_point1_ry": input.CalibPoint1RY,
		"calib_point2_x":  input.CalibPoint2X,
		"calib_point2_y":  input.CalibPoint2Y,
		"calib_point2_rx": input.CalibPoint2RX,
		"calib_point2_ry": input.CalibPoint2RY,
	})

	c.JSON(http.StatusOK, model.Success(m))
}

// SetActiveMap 激活地图
func (h *MapHandler) SetActiveMap(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Model(&model.MapConfig{}).Where("active = ?", true).Update("active", false)
	h.db.Model(&model.MapConfig{}).Where("id = ?", id).Update("active", true)
	c.JSON(http.StatusOK, model.Success(nil))
}

// GetActiveMap 获取当前激活地图
func (h *MapHandler) GetActiveMap(c *gin.Context) {
	var m model.MapConfig
	h.db.Where("active = ?", true).First(&m)
	c.JSON(http.StatusOK, model.Success(m))
}

// DeleteMap 删除地图
func (h *MapHandler) DeleteMap(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m model.MapConfig
	if h.db.First(&m, id).Error != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "not found"))
		return
	}

	// 从 MinIO 删除
	if h.minio != nil && m.ImageURL != "" {
		objName := extractObjectName(m.ImageURL)
		if objName != "" {
			h.minio.RemoveObject(context.Background(), h.bucketName, objName, minio.RemoveObjectOptions{})
		}
	}

	h.db.Delete(&m)
	c.JSON(http.StatusOK, model.Success(nil))
}

func sanitizeName(name string) string {
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			result += string(c)
		} else {
			result += "_"
		}
	}
	if len(result) > 64 {
		result = result[:64]
	}
	return result
}

func extractObjectName(url string) string {
	for i := len(url)-1; i >= 0; i-- {
		if url[i] == '/' {
			return url[i+1:]
		}
	}
	return ""
}

// Ensure io is used (for interface compliance)
var _ = io.Discard
