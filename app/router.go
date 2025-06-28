package app

import (
	"gin-temp/middleware"
	"gin-temp/util"
	"time"

	"github.com/gin-gonic/gin"
)

func setupPublicRoutes(app *gin.Engine) {
	app.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "home page"})
	})

	app.POST("/register", userCtrl.CreateUser)

	app.POST(
		"/login",
		userCtrl.Login,
		middleware.RateLimiter(util.Redis, util.Viper.GetInt("rate.login"), time.Minute),
	)
}

func setupAdminRoutes(app *gin.Engine) {
	r1 := app.Group("/admin")
	r1.Use(handlerFuncListOfAuthAndRbac...)
	{
		r1.GET("", userCtrl.GetAllUsers)
		r1.GET("/search", userCtrl.SearchUserByKeyward)
		r1.GET("/:id", userCtrl.GetUserById)
		r1.PUT("/role/:id", userCtrl.ChangeUserRole)
		r1.DELETE("/delete-user/:id", userCtrl.DeleteUser)
	}
}

func setupUserRoutes(app *gin.Engine) {
	r1 := app.Group("/users")
	r1.Use(handlerFuncListOfAuthAndRbac...)
	{
		r1.PUT("/logout", userCtrl.Logout)
		r1.PUT("/modify", userCtrl.ChangeUserinfo)
		r1.PUT("/password", userCtrl.ChangePassword)
		r1.POST("/refresh", userCtrl.TokenRefresh)
	}
}
