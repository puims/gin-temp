package router

import (
	"gin-temp/utils"

	"github.com/gin-gonic/gin"
)

func SetupApp() (*gin.Engine, *utils.MysqlDB, *utils.CasbinPolicyLoader) {
	initLog()
	initValidator()

	app := initApp()

	setupPublicRoutes(app)
	setupAdminRoutes(app, utils.Enforcer)
	setupUserRoutes(app, utils.Enforcer)

	return app, utils.DB, utils.PolicyLoader
}
