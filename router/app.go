package router

import (
	"gin-temp/controller"
	"gin-temp/models"
	"gin-temp/utils"
	"log"

	"github.com/gin-gonic/gin"
)

var (
	db = models.NewMysqlDB(&models.User{})

	userCtrl = &controller.UserController{DB: db}
)

func SetupApp() (*gin.Engine, *models.MysqlDB, *utils.CasbinPolicyLoader) {
	initLog()

	if err := initRoot(db); err != nil {
		log.Println(err)
	}

	app := initApp()

	enforcer, loader, err := utils.SetupCasbin(db.DB)
	if err != nil {
		log.Fatal("Failed to init casbin", err)
	}

	setupPublicRoutes(app)
	setupAdminRoutes(app, enforcer)
	setupUserRoutes(app, enforcer)

	return app, db, loader
}
