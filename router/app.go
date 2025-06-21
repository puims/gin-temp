package router

import (
	"gin-temp/controller"
	"gin-temp/models"
	"gin-temp/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func SetupApp() (*gin.Engine, *models.MysqlDB, *utils.CasbinPolicyLoader) {
	initLog()

	app := initApp()

	db, err := models.NewMysqlDB(&models.User{}, &models.Role{}, &models.UserRole{})
	if err != nil {
		panic(err)
	}

	enforcer, loader, err := utils.SetupCasbin(db.DB)
	if err != nil {
		log.Fatal("Failed to init casbin", err)
	}

	authCtrl := &controller.AuthController{DB: db}
	adminCtrl := &controller.AdminController{DB: db}

	setupPublicRoutes(app, authCtrl)
	setupAdminRoutes(app, enforcer, adminCtrl)
	setupUserRoutes(app, enforcer, adminCtrl)

	return app, db, loader
}
