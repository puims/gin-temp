package router

import (
	"gin-temp/controller"
	"gin-temp/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func setupPublicRoutes(app *gin.Engine, authCtrl *controller.AuthController) {
	app.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "home page"})
	})

	app.POST("/register", authCtrl.Register)

	app.POST("/login", authCtrl.Login)
}

func setupAdminRoutes(app *gin.Engine, enforcer *casbin.Enforcer, adminCtrl *controller.AdminController) {
	r0 := app.Group(
		"/admin/",
		middleware.JwtAuthorization(),
		middleware.CasbinAuthorization(enforcer),
	)
	{
		r0.GET("/", adminCtrl.GetAllUsers)
		r0.GET("/:id", adminCtrl.GetUserById)
		r0.PUT("/modify/", adminCtrl.UpdateUser)
		r0.DELETE("/delete/:id", adminCtrl.DeleteUser)
	}
}

func setupUserRoutes(app *gin.Engine, enforcer *casbin.Enforcer, adminCtrl *controller.AdminController) {
	r1 := app.Group(
		"/users/",
		middleware.JwtAuthorization(),
		middleware.CasbinAuthorization(enforcer),
	)
	{
		r1.PUT("/modify/", adminCtrl.UpdateUser)
	}
}
