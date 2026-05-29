package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

type RobotHandler struct {
	db *gorm.DB
}

func NewRobotHandler(db *gorm.DB) *RobotHandler {
	return &RobotHandler{db: db}
}

// List 机器人列表（分页 + 筛选）
// GET /api/v1/robots?page=1&page_size=20&model=AGV&status=running
func (h *RobotHandler) List(c *gin.Context) {
	var pag model.Pagination
	c.ShouldBindQuery(&pag)
	pag.Default()

	var robots []model.Robot
	query := h.db.Model(&model.Robot{})

	if model := c.Query("model"); model != "" {
		query = query.Where("model = ?", model)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if batteryMin := c.Query("battery_min"); batteryMin != "" {
		query = query.Where("battery_pct >= ?", batteryMin)
	}

	query.Count(&pag.Total)
	query.Offset(pag.Offset()).Limit(pag.PageSize).Find(&robots)

	c.JSON(http.StatusOK, model.Response{Code: 200, Message: "success", Data: gin.H{
		"list":       robots,
		"pagination": pag,
	}, Timestamp: 0})
}

// Get 机器人详情
// GET /api/v1/robots/:id
func (h *RobotHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var robot model.Robot
	if err := h.db.First(&robot, id).Error; err != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "robot not found"))
		return
	}

	c.JSON(http.StatusOK, model.Success(robot))
}

// Create 注册机器人
// POST /api/v1/robots
func (h *RobotHandler) Create(c *gin.Context) {
	var robot model.Robot
	if err := c.ShouldBindJSON(&robot); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, err.Error()))
		return
	}

	if err := h.db.Create(&robot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.Error(500, "failed to create robot"))
		return
	}

	c.JSON(http.StatusCreated, model.Success(robot))
}

// Update 更新机器人
// PUT /api/v1/robots/:id
func (h *RobotHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var robot model.Robot
	if err := h.db.First(&robot, id).Error; err != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "robot not found"))
		return
	}

	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)

	// 不允许修改 id, robot_code, created_at
	delete(updates, "id")
	delete(updates, "robot_code")
	delete(updates, "created_at")

	if err := h.db.Model(&robot).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.Error(500, "failed to update robot"))
		return
	}

	c.JSON(http.StatusOK, model.Success(robot))
}

// Delete 注销机器人
// DELETE /api/v1/robots/:id
func (h *RobotHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.Robot{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.Error(500, "failed to delete robot"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// SendCommand 下发控制指令
// POST /api/v1/robots/:id/commands
func (h *RobotHandler) SendCommand(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var cmd struct {
		Command string                 `json:"command"` // start/pause/stop/charge/light
		Params  map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, err.Error()))
		return
	}

	// TODO: 通过 MQTT 下发指令到机器人
	c.JSON(http.StatusOK, model.Success(gin.H{
		"robot_id": id,
		"command":  cmd.Command,
		"status":   "queued",
	}))
}
