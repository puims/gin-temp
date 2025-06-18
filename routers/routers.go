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

	app.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "home page\n")
	})
	app.POST("/register", controllers.CreateUser)
	app.POST("/login", controllers.Login)

	r1 := app.Group("users",
		middlewares.Authorization,
		// middlewares.ProHandler,
	)
	{
		r1.GET("/", controllers.GetAllUsers)
		r1.GET("/:id", controllers.GetUserById)

		r1.PUT("/modify/:id", controllers.UpdateUser)
		r1.DELETE("/delete/:id", controllers.DeleteUser)
	}

	r2 := app.Group("articles")
	{
		r2.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"msg": "article page",
			})
		})
	}

	return
}

func initLog() {
	gin.DisableConsoleColor()

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
}
