package controller

import (
	"errors"
	"gin-temp/models"
	"gin-temp/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *models.MysqlDB
}

func (ac *AuthController) TokenRefresh(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	startTime := time.Now()
	var user models.User
	if err := ac.DB.Select("id", "username", "email").
		First(&user, claims.(*utils.Claims).ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
		"response_time": time.Since(startTime).Milliseconds(),
	})
}
