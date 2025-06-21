package router

import (
	"gin-temp/controller"
	"gin-temp/models"
	"gin-temp/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func SetupApp() (*gin.Engine, *models.MysqlDB, *utils.CasbinPolicyLoader) {
	// 初始化日志
	initLog()

	// 初始化Gin引擎
	app := initApp()

	// 初始化数据库
	db, err := models.NewMysqlDB(&models.User{}, &models.Role{})
	if err != nil {
		panic(err)
	}

	enforcer, loader, err := utils.SetupCasbin(db.DB)
	if err != nil {
		log.Fatal("Failed to init casbin", err)
	}

	// 初始化控制器
	authController := &controller.AuthController{DB: db}
	adminController := &controller.AdminController{DB: db}

	// 设置路由
	setupPublicRoutes(app, authController)
	setupAdminRoutes(app, enforcer, adminController)
	setupUserRoutes(app, enforcer, adminController)

	return app, db, loader
}
