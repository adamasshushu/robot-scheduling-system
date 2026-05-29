package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
	"github.com/adamasshushu/robot-scheduling-system/internal/service"
)

type TaskHandler struct {
	db        *gorm.DB
	scheduler *service.Scheduler
}

func NewTaskHandler(db *gorm.DB, scheduler *service.Scheduler) *TaskHandler {
	return &TaskHandler{db: db, scheduler: scheduler}
}

// List 任务列表
// GET /api/v1/tasks?page=1&page_size=20&status=pending&robot_id=1
func (h *TaskHandler) List(c *gin.Context) {
	var pag model.Pagination
	c.ShouldBindQuery(&pag)
	pag.Default()

	query := h.db.Model(&model.Task{}).Preload("Robot")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if robotID := c.Query("robot_id"); robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}

	query.Count(&pag.Total)

	var tasks []model.Task
	query.Order("priority ASC, created_at DESC").
		Offset(pag.Offset()).Limit(pag.PageSize).Find(&tasks)

	c.JSON(http.StatusOK, model.Response{Code: 200, Message: "success", Data: gin.H{
		"list":       tasks,
		"pagination": pag,
	}, Timestamp: 0})
}

// Get 任务详情
func (h *TaskHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var task model.Task
	if err := h.db.Preload("Robot").First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "task not found"))
		return
	}

	c.JSON(http.StatusOK, model.Success(task))
}

// Create 创建任务
func (h *TaskHandler) Create(c *gin.Context) {
	var input struct {
		TaskType       string  `json:"task_type" binding:"required"`
		Priority       int     `json:"priority"`
		TargetLocation string  `json:"target_location" binding:"required"`
		SourceLocation string  `json:"source_location"`
		TargetX        float64 `json:"target_x"`
		TargetY        float64 `json:"target_y"`
		Description    string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, err.Error()))
		return
	}

	if input.Priority < 1 || input.Priority > 10 {
		input.Priority = 5
	}

	now := time.Now()
	task := model.Task{
		TaskCode:       fmt.Sprintf("TSK-%s-%04d", now.Format("20060102"), now.UnixMilli()%10000),
		TaskType:       input.TaskType,
		Priority:       input.Priority,
		Status:         "pending",
		TargetLocation: input.TargetLocation,
		SourceLocation: input.SourceLocation,
		Description:    input.Description,
		CreatedAt:      now,
	}

	if input.TargetX != 0 || input.TargetY != 0 {
		task.TargetX = &input.TargetX
		task.TargetY = &input.TargetY
	}

	if err := h.db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.Error(500, "failed to create task"))
		return
	}

	c.JSON(http.StatusCreated, model.Success(task))
}

// Assign 指派机器人
func (h *TaskHandler) Assign(c *gin.Context) {
	taskID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var input struct {
		RobotID uint `json:"robot_id"`
		Auto    bool `json:"auto"`
	}
	c.ShouldBindJSON(&input)

	// 先获取 task
	var task model.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "task not found"))
		return
	}

	if task.Status != "pending" {
		c.JSON(http.StatusBadRequest, model.Error(400, "task is not in pending status"))
		return
	}

	// 智能指派：使用多因子评分算法
	if input.Auto {
		robot, err := h.scheduler.AutoAssign(&task)
		if err != nil || robot == nil {
			c.JSON(http.StatusNotFound, model.Error(404, "no available robot found (need standby + battery>20%)"))
			return
		}
		input.RobotID = robot.ID
	}

	// 检查机器人是否存在
	var robot model.Robot
	if err := h.db.First(&robot, input.RobotID).Error; err != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "robot not found"))
		return
	}

	h.db.Model(&task).Updates(map[string]interface{}{
		"robot_id": input.RobotID,
		"status":   "assigned",
	})

	c.JSON(http.StatusOK, model.Success(task))
}

// Cancel 取消任务
func (h *TaskHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var task model.Task
	if err := h.db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "task not found"))
		return
	}

	if task.Status == "completed" || task.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, model.Error(400, "task already finished"))
		return
	}

	h.db.Model(&task).Update("status", "cancelled")
	c.JSON(http.StatusOK, model.Success(task))
}
