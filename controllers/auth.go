package controllers

import (
	"errors"
	"gin-temp/models"
	"gin-temp/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func (ac *AuthController) Register(ctx *gin.Context) {
	userIn := models.UserRegister{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	user := models.User{}
	if err := ac.DB.Where("username = ? OR email = ?", userIn.Username, userIn.Email).
		First(&user).Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	hashPwd, err := utils.HashPassword(userIn.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Could not hash password"})
		return
	}

	role := models.Role{}
	if err := ac.DB.First(&role, "name = ?", "user").Error; err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to find role.name:user"})
		return
	}

	user.Username = userIn.Username
	user.Password = hashPwd
	user.Email = userIn.Email
	user.Roles = []models.Role{role}

	if err := ac.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to create user"})
		return
	}
	ctx.JSON(http.StatusCreated, user)
}

func (ac *AuthController) Login(ctx *gin.Context) {
	userIn := models.UserLogin{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	user := models.User{}
	if err := ac.DB.Preload("Roles").First(&user, "username = ?", userIn.Username).
		Error; err != nil {
		ctx.JSON(http.StatusUnauthorized,
			gin.H{"error": "Invalid credentials"})
		return
	}
	if err := utils.VerifyPassword(userIn.Password, user.Password); err != nil {
		ctx.JSON(http.StatusUnauthorized,
			gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to generate token"})
		return
	}

	err = models.StoreToken(user.ID, token["token"].(string))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{
		"token": token,
		"user":  user,
	})
}

func (ac *AuthController) TokenRefresh(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	user := models.User{}
	if err := ac.DB.First(&user, claims.(*utils.Claims).ID).Error; err != nil {
		ctx.JSON(401, gin.H{"error": "User not found"})
		return
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Could not generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
