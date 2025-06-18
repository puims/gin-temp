package controllers

import (
	"gin-temp/middlewares"
	"gin-temp/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

var userFoundError = gin.H{"Error": "User not found"}

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
		ctx.JSON(200, userFoundError)
		return
	}
	ctx.JSON(200, user)
}

func CreateUser(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password, _ = HashPassword(password)

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
		ctx.JSON(200, userFoundError)
		return
	}

	password := ctx.PostForm("password")
	password, _ = HashPassword(password)

	user.Name = ctx.PostForm("username")
	user.Password = password

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
		ctx.JSON(200, userFoundError)
		return
	}

	models.DB.Delete(&user)
	ctx.JSON(200, gin.H{
		"status": "OK",
	})
}

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	user := models.UserInfo{}
	models.DB.Where("name = ?", username).First(&user)
	if user.ID == 0 {
		ctx.JSON(200, userFoundError)
		return
	}

	state, err := VerifyPassword(password, user.Password)
	if err != nil {
		ctx.JSON(200, gin.H{
			"state": state,
			"error": err,
		})
		return
	}

	token, err := middlewares.NewToken(user.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, token)
	} else {
		ctx.JSON(200, token)
	}
}
