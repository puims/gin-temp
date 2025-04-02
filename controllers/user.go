package controllers

import "github.com/gin-gonic/gin"

func GetAllUsers(c *gin.Context) {
	c.String(200, "user page: get all users")
}

func GetUserById(c *gin.Context) {
	id := c.Param("id")
	c.String(200, "get user by id:"+id)
}
