package routers

import (
	"gin-temp/controllers"

	"github.com/gin-gonic/gin"
)

func App() (app *gin.Engine) {
	app = gin.Default()

	// app.Use(middlewares.MiddleWare())

	app.GET("/", func(c *gin.Context) {
		c.String(200, "home page")
	})

	r1 := app.Group("user")
	{
		r1.GET("/", controllers.GetAllUsers)
		r1.GET("/:id", controllers.GetUserById)
	}

	return
}
