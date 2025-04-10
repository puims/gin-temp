package controllers

import (
	"gin-temp/models"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(c *gin.Context) {
	var users []models.UserInfo
	models.DB.Find(&users)

	c.JSON(200, users)
}

func GetUserById(c *gin.Context) {
	id := c.Param("id")
	var user models.UserInfo
	models.DB.Where("id = ?", id).First(&user)

	c.JSON(200, user)
}
