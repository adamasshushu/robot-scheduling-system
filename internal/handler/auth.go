package handler

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/middleware"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

// ── 速率限制 ──
var (
	loginAttempts = make(map[string]*loginWindow)
	rateMu        sync.Mutex
)

type loginWindow struct {
	count    int
	firstTry time.Time
}

// 同一 IP 每分钟最多 5 次登录尝试
func checkRateLimit(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()
	now := time.Now()
	w, ok := loginAttempts[ip]
	if !ok || now.Sub(w.firstTry) > time.Minute {
		loginAttempts[ip] = &loginWindow{count: 1, firstTry: now}
		return true
	}
	w.count++
	return w.count <= 5
}

// ── AuthHandler ──

type AuthHandler struct {
	db        *gorm.DB
	jwtExpiry time.Duration
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	expiryStr := os.Getenv("JWT_EXPIRY")
	if expiryStr == "" {
		expiryStr = "2h"
	}
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 2 * time.Hour
	}

	// 定期清理过期的登录尝试记录
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rateMu.Lock()
			for ip, entry := range loginAttempts {
				if time.Since(entry.firstTry) > 15*time.Minute {
					delete(loginAttempts, ip)
				}
			}
			rateMu.Unlock()
		}
	}()

	return &AuthHandler{db: db, jwtExpiry: expiry}
}

// Login 用户登录
// POST /api/v1/login
func (h *AuthHandler) Login(c *gin.Context) {
	// 速率限制
	ip := c.ClientIP()
	if !checkRateLimit(ip) {
		c.JSON(http.StatusTooManyRequests, model.Error(429, "too many login attempts, try again later"))
		return
	}

	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Error(400, "invalid input"))
		return
	}

	// 输入长度限制（防超大 payload）
	if len(input.Username) > 64 || len(input.Password) > 128 {
		c.JSON(http.StatusBadRequest, model.Error(400, "username or password too long"))
		return
	}

	var user model.User
	if err := h.db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, model.Error(401, "invalid credentials"))
		return
	}

	if !user.CheckPassword(input.Password) {
		c.JSON(http.StatusUnauthorized, model.Error(401, "invalid credentials"))
		return
	}

	// 生成 JWT — 使用配置的过期时间
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, h.jwtExpiry)
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
