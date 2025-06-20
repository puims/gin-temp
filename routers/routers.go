package routers

import (
	"fmt"
	"gin-temp/config"
	"gin-temp/controllers"
	"gin-temp/middleware"
	"gin-temp/models"
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

var App *gin.Engine

func init() {
	initLog()
	initEnv()

	App = gin.Default()

	authController := controllers.AuthController{DB: models.DB}
	userController := controllers.UserController{DB: models.DB}

	{
		App.GET("/",
			func(ctx *gin.Context) {
				ctx.JSON(200,
					fmt.Sprintf("%s home page", config.Viper.GetString("app.name")))
			},
		)
		App.POST("/register", authController.Register)
		App.POST("/login", authController.Login)
	}

	r1 := App.Group(
		"/users",
		middleware.Authorization(),
	)
	{
		r1.PUT("/modify/", userController.UpdateUser)
	}

	r2 := App.Group(
		"/admin",
		middleware.Authorization(),
		middleware.RoleMidware([]string{"admin"}),
	)
	{
		r2.GET("/", userController.GetAllUsers)
		r2.GET("/:id", userController.GetUserById)
		r2.PUT("/modify/", userController.UpdateUser)
		r2.DELETE("/delete/:id", userController.DeleteUser)
	}
}

func initLog() {
	gin.DisableConsoleColor()

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
}

func initEnv() {
	if appMode := config.Viper.GetBool("app.debug"); !appMode {
		gin.SetMode(gin.ReleaseMode)
	}
}
