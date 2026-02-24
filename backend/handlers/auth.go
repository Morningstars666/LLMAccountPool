package handlers

import (
	"llmaccountpool/config"
	"llmaccountpool/models"
	"llmaccountpool/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type ChangeUsernameRequest struct {
	NewUsername string `json:"new_username" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

func Login(c *gin.Context, cfg *config.Config) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 检查账户是否被锁定
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		remaining := int(user.LockedUntil.Sub(time.Now()).Minutes())
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":             "Account is temporarily locked",
			"remaining_minutes": remaining,
		})
		return
	}

	// 重置失败的尝试（如果锁定时间已过）
	if user.LockedUntil != nil && user.LockedUntil.Before(time.Now()) {
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		models.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("failed_login_attempts", gorm.Expr("failed_login_attempts + 1"))

		var updatedUser models.User
		models.DB.First(&updatedUser, user.ID)

		if updatedUser.FailedLoginAttempts >= cfg.MaxLoginAttempts {
			lockUntil := time.Now().Add(time.Duration(cfg.LockoutDuration) * time.Minute)
			models.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
				"locked_until": lockUntil,
			})
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many failed attempts. Account locked for 15 minutes",
			})
		} else {
			remaining := cfg.MaxLoginAttempts - updatedUser.FailedLoginAttempts
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":              "Invalid credentials",
				"remaining_attempts": remaining,
			})
		}
		return
	}

	now := time.Now()
	models.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"failed_login_attempts": 0,
		"locked_until":          nil,
		"last_login_at":         now,
	})

	token, err := utils.GenerateJWT(cfg.JWTSecret, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "username": user.Username})
}

func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": user.Username})
}

func RefreshToken(c *gin.Context, cfg *config.Config) {
	userID := c.GetUint("user_id")

	token, err := utils.GenerateJWT(cfg.JWTSecret, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 验证新密码强度
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	isValid, message := utils.ValidatePasswordStrength(req.NewPassword)
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Old password is incorrect"})
		return
	}

	hash, _ := utils.HashPassword(req.NewPassword)
	user.Password = hash
	models.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func ChangeUsername(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req ChangeUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is incorrect"})
		return
	}

	if req.NewUsername == user.Username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New username must be different from current username"})
		return
	}

	var existingUser models.User
	if err := models.DB.Where("username = ? AND id != ?", req.NewUsername, user.ID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	user.Username = req.NewUsername
	models.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Username changed successfully", "username": user.Username})
}
