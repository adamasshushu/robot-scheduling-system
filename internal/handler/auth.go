package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/middleware"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Login 用户登录
// POST /api/v1/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "invalid input"))
		return
	}

	var user model.User
	if err := h.db.Where("username = ? AND password = ?", input.Username, input.Password).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, model.Error(401, "invalid credentials"))
		return
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, 2*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Error(500, "failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
	}))
}
