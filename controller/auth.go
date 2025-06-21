package controller

import (
	"errors"
	"gin-temp/models"
	"gin-temp/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *models.MysqlDB
}

type UserRegister struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserProfile struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	AdminKey string `json:"adminkey"`
}

func (ac *AuthController) Register(ctx *gin.Context) {
	userIn := UserRegister{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	user := models.User{}
	if err := ac.DB.First(&user, "username = ? OR email = ?", userIn.Username, userIn.Email).
		Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	err := ac.DB.Transaction(func(tx *gorm.DB) error {
		user := models.User{
			Username: userIn.Username,
			Password: userIn.Password,
			Email:    userIn.Email,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (ac *AuthController) Login(ctx *gin.Context) {
	userIn := UserLogin{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()
	user := models.User{}
	if err := ac.DB.Select("id", "username", "password", "email").
		First(&user, "username = ?", userIn.Username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userIn.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized,
			gin.H{"error": "Invalid credentials"})
		return
	}

	if err := ac.DB.Model(&user).Association("Roles").Find(&user.Roles); err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to load user roles"})
		return
	}

	userRoles := []string{}
	for _, rl := range user.Roles {
		userRoles = append(userRoles, rl.Name)
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.ID,
			"email":    user.Email,
			"roles":    userRoles,
		},
		"response_time": time.Since(startTime).Milliseconds(),
	})
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
