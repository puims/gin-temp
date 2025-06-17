package controllers

import (
	"gin-temp/models"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(ctx *gin.Context) {
	var users []models.UserInfo
	models.DB.Find(&users)

	ctx.JSON(200, users)
}

func GetUserById(ctx *gin.Context) {
	id := ctx.Param("id")
	var user models.UserInfo
	models.DB.Where("id = ?", id).First(&user)

	if user.ID == 0 {
		ctx.JSON(200, gin.H{
			"Error": "User not found",
		})
		return
	}
	ctx.JSON(200, user)
}

func CreateUser(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	user := models.UserInfo{
		Name:     username,
		Password: password,
	}
	models.DB.Create(&user)

	ctx.JSON(200, gin.H{
		"user": user,
	})
}

func UpdateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	user := models.UserInfo{}
	models.DB.Where("id = ?", id).First(&user)

	if user.ID == 0 {
		ctx.JSON(200, gin.H{
			"Error": "User not found",
		})
		return
	}

	user.Name = ctx.PostForm("username")
	user.Password = ctx.PostForm("password")

	models.DB.Save(&user)

	ctx.JSON(200, gin.H{
		"user": user,
	})
}

func DeleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	user := models.UserInfo{}
	models.DB.Where("id = ?", id).First(&user)

	if user.ID == 0 {
		ctx.JSON(200, gin.H{
			"Error": "User not found",
		})
		return
	}

	models.DB.Delete(&user)
	ctx.JSON(200, gin.H{
		"msg": "OK",
	})
}
