package router

import (
	"gin-temp/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func setupPublicRoutes(app *gin.Engine) {
	app.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "home page"})
	})

	app.POST("/register", userCtrl.CreateUser)

	app.POST("/login", userCtrl.LoginCheck)
}

func setupAdminRoutes(app *gin.Engine, enforcer *casbin.Enforcer) {
	r1 := app.Group(
		"/admin",
		middleware.JwtAuthorization(),
		middleware.CasbinAuthorization(enforcer),
	)
	{
		r1.GET("", userCtrl.GetAllUsers)
		r1.GET("/search", userCtrl.SearchUserByKeyward)
		r1.GET("/:id", userCtrl.GetUserById)
		r1.PUT("/role/:id", userCtrl.ChangeUserRole)
		r1.DELETE("/delete-user/:id", userCtrl.DeleteUser)
	}
}

func setupUserRoutes(app *gin.Engine, enforcer *casbin.Enforcer) {
	r1 := app.Group(
		"/users",
		middleware.JwtAuthorization(),
		middleware.CasbinAuthorization(enforcer),
	)
	{
		r1.PUT("/modify", userCtrl.ChangeUserinfo)
		r1.PUT("/password", userCtrl.ChangePassword)
	}
}
