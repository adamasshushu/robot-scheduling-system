package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

type SettingsHandler struct {
	db *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

// ── 用户管理 ──

// ListUsers 用户列表
func (h *SettingsHandler) ListUsers(c *gin.Context) {
	var users []model.User
	var p model.Pagination
	c.ShouldBindQuery(&p)
	p.Default()

	var total int64
	h.db.Model(&model.User{}).Count(&total)
	h.db.Offset(p.Offset()).Limit(p.PageSize).Find(&users)

	c.JSON(http.StatusOK, model.Success(gin.H{
		"list":       users,
		"pagination": gin.H{"page": p.Page, "page_size": p.PageSize, "total": total},
	}))
}

// CreateUser 创建用户
func (h *SettingsHandler) CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "invalid user data"))
		return
	}
	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, model.Error(400, "username and password required"))
		return
	}
	// 密码强度：至少 8 位，含字母和数字
	if len(input.Password) < 8 {
		c.JSON(http.StatusBadRequest, model.Error(400, "password must be at least 8 characters"))
		return
	}
	hasLetter, hasDigit := false, false
	for _, c := range input.Password {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' { hasLetter = true }
		if c >= '0' && c <= '9' { hasDigit = true }
	}
	if !hasLetter || !hasDigit {
		c.JSON(http.StatusBadRequest, model.Error(400, "password must contain both letters and numbers"))
		return
	}
	if len(input.Username) < 3 || len(input.Username) > 32 {
		c.JSON(http.StatusBadRequest, model.Error(400, "username must be 3-32 characters"))
		return
	}
	role := input.Role
	if role == "" {
		role = "operator"
	}

	user := model.User{
		Username: input.Username,
		Role:     role,
	}
	user.SetPassword(input.Password) // bcrypt hash
	result := h.db.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusConflict, model.Error(409, "username already exists"))
		return
	}
	c.JSON(http.StatusOK, model.Success(user))
}

// UpdateUser 更新用户
func (h *SettingsHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var input model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "invalid data"))
		return
	}
	var user model.User
	if h.db.First(&user, id).Error != nil {
		c.JSON(http.StatusNotFound, model.Error(404, "user not found"))
		return
	}
	updates := map[string]interface{}{}
	if input.Username != "" { updates["username"] = input.Username }
	if input.Password != "" {
		if len(input.Password) < 8 {
			c.JSON(http.StatusBadRequest, model.Error(400, "password must be at least 8 characters"))
			return
		}
		hash, _ := model.HashPassword(input.Password)
		updates["password"] = hash
	}
	if input.Role != "" { updates["role"] = input.Role }
	h.db.Model(&user).Updates(updates)
	c.JSON(http.StatusOK, model.Success(user))
}

// DeleteUser 删除用户
func (h *SettingsHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 1 {
		c.JSON(http.StatusForbidden, model.Error(403, "cannot delete admin"))
		return
	}
	if h.db.Delete(&model.User{}, id).RowsAffected == 0 {
		c.JSON(http.StatusNotFound, model.Error(404, "user not found"))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// ── 机器人型号管理 ──

// ListModels 型号列表
func (h *SettingsHandler) ListModels(c *gin.Context) {
	var models []model.RobotModel
	h.db.Find(&models)
	c.JSON(http.StatusOK, model.Success(models))
}

// CreateModel 创建型号
func (h *SettingsHandler) CreateModel(c *gin.Context) {
	var m model.RobotModel
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "invalid data"))
		return
	}
	if m.Name == "" {
		c.JSON(http.StatusBadRequest, model.Error(400, "name required"))
		return
	}
	if h.db.Create(&m).Error != nil {
		c.JSON(http.StatusConflict, model.Error(409, "model exists"))
		return
	}
	c.JSON(http.StatusOK, model.Success(m))
}

// DeleteModel 删除型号
func (h *SettingsHandler) DeleteModel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if h.db.Delete(&model.RobotModel{}, id).RowsAffected == 0 {
		c.JSON(http.StatusNotFound, model.Error(404, "not found"))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}
