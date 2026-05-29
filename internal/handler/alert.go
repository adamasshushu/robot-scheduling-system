package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

type AlertHandler struct {
	db *gorm.DB
}

func NewAlertHandler(db *gorm.DB) *AlertHandler {
	return &AlertHandler{db: db}
}

// List 告警列表
// GET /api/v1/alerts?page=1&page_size=20&status=unack
func (h *AlertHandler) List(c *gin.Context) {
	var pag model.Pagination
	c.ShouldBindQuery(&pag)
	pag.Default()

	query := h.db.Model(&model.Alert{}).Preload("Robot")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if alertType := c.Query("type"); alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}

	query.Count(&pag.Total)

	var alerts []model.Alert
	query.Order("created_at DESC").
		Offset(pag.Offset()).Limit(pag.PageSize).Find(&alerts)

	c.JSON(http.StatusOK, model.Response{Code: 200, Message: "success", Data: gin.H{
		"list":       alerts,
		"pagination": pag,
	}, Timestamp: 0})
}

// Ack 确认告警
// POST /api/v1/alerts/:id/ack
func (h *AlertHandler) Ack(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	now := time.Now()
	userID, _ := c.Get("user_id")

	h.db.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          "acknowledged",
		"acknowledged_by": userID,
		"acknowledged_at": now,
	})

	c.JSON(http.StatusOK, model.Success(nil))
}

// Resolve 解决告警
// POST /api/v1/alerts/:id/resolve
func (h *AlertHandler) Resolve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	now := time.Now()

	h.db.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_at": now,
	})

	c.JSON(http.StatusOK, model.Success(nil))
}

// Stats 告警统计
// GET /api/v1/alerts/stats
func (h *AlertHandler) Stats(c *gin.Context) {
	var total, unack, critical int64

	h.db.Model(&model.Alert{}).Count(&total)
	h.db.Model(&model.Alert{}).Where("status = ?", "unack").Count(&unack)
	h.db.Model(&model.Alert{}).Where("severity = ? AND status = ?", "critical", "unack").Count(&critical)

	c.JSON(http.StatusOK, model.Success(gin.H{
		"total":    total,
		"unack":    unack,
		"critical": critical,
	}))
}
