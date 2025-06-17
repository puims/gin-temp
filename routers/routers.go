package routers

import (
	"gin-temp/controllers"
	"gin-temp/middlewares"
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func App() (app *gin.Engine) {
	app = gin.Default()

	initLog()
	app.Use(middlewares.MiddleWare())

	app.GET("/", func(c *gin.Context) {
		c.String(200, "home page\n")
	})

	r1 := app.Group("users", middlewares.Authorization())
	{
		r1.GET("/", controllers.GetAllUsers)
		r1.GET("/:id", controllers.GetUserById)
		r1.POST("/", controllers.CreateUser)
		r1.PUT("/:id", controllers.UpdateUser)
		r1.DELETE("/:id", controllers.DeleteUser)
	}

	return
}

func initLog() {
	gin.DisableConsoleColor()

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
}
