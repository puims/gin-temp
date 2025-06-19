package controllers

import (
	"fmt"
	"gin-temp/models"
	"gin-temp/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func (uc *UserController) GetAllUsers(ctx *gin.Context) {
	var users []models.User
	uc.DB.Find(&users)

	ctx.JSON(200, users)
}

func (uc *UserController) GetUserById(ctx *gin.Context) {
	id := ctx.Param("id")
	var user models.User
	uc.DB.Where("id = ?", id).First(&user)

	if user.ID == 0 {
		ctx.JSON(200, gin.H{"error": "User not found"})
		return
	}
	ctx.JSON(200, user)
}

func (uc *UserController) UpdateUser(ctx *gin.Context) {
	userIn := models.UserLogin{}
	if err := ctx.ShouldBindJSON(&userIn); err != nil {
		ctx.JSON(403, gin.H{"error": err.Error()})
		return
	}

	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	user := models.User{}
	if err := uc.DB.Where("id = ?", claims.(*utils.Claims).ID).First(&user).Error; err != nil {
		ctx.JSON(200, gin.H{"error": "Failed to search user from database"})
		return
	}

	hashPwd, err := utils.HashPassword(userIn.Password)
	if err != nil {
		ctx.JSON(403, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Username = userIn.Username
	user.Password = hashPwd

	if err := models.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to save data to database"})
		return
	}

	ctx.JSON(200, user)
}

func (uc *UserController) DeleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	user := models.User{}

	err := models.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		ctx.JSON(403, gin.H{"error": "User not found"})
		return
	}

	err = models.DB.Delete(&user).Error
	if err != nil {
		ctx.JSON(403, gin.H{"error": "Failed to delete user"})
		return
	}

	ctx.JSON(200, gin.H{
		"state": true,
		"msg":   fmt.Sprintf("User %s has deleted", user.Username),
	})
}
