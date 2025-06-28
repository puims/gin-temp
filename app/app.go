package app

import (
	"gin-temp/middleware"
	"gin-temp/util"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupApp() *gin.Engine {
	initValidator()

	app := initApp()

	app.Use(gin.Recovery())
	app.Use(middleware.RateLimiter(
		util.Redis, util.Viper.GetInt("rate.normal"), time.Second))

	setupPublicRoutes(app)
	setupAdminRoutes(app)
	setupUserRoutes(app)

	return app
}
