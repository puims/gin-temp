package routers

import (
	"gin-temp/controllers"
	"gin-temp/middlewares"
	"gin-temp/models"
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func App() (app *gin.Engine) {
	app = gin.Default()

	initLog()

	authController := controllers.AuthController{DB: models.DB}
	userController := controllers.UserController{DB: models.DB}

	{
		app.GET("/",
			func(ctx *gin.Context) { ctx.JSON(200, "home page") },
		)
		app.POST("/register", authController.Register)
		app.POST("/login", authController.Login)
	}

	r1 := app.Group("/users")
	r1.Use(middlewares.Authorization())
	{
		r1.GET("/", userController.GetAllUsers)
		r1.GET("/:id", userController.GetUserById)

		r1.PUT("/modify/", userController.UpdateUser)
		r1.DELETE("/delete/:id", userController.DeleteUser)
	}

	r2 := app.Group("/admin")
	r2.Use(
		middlewares.Authorization(),
		middlewares.RoleMidware([]string{"admin", "edtior"}),
	)
	{
		r2.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"msg": "admin page",
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
