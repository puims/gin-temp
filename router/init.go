package router

import (
	"gin-temp/controller"
	"gin-temp/utils"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func initApp() *gin.Engine {
	if !utils.Viper.GetBool("app.debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	return gin.Default()
}

func initLog() {
	gin.DisableConsoleColor()

	f, err := os.Create("gin.log")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}

	gin.DefaultWriter = io.MultiWriter(f)
}

func initValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("password",
			controller.PasswordValidator); err != nil {
			panic("failed to register password validator")
		}
		if err := v.RegisterValidation("username",
			controller.UsernameValidator); err != nil {
			panic("failed to register username validator")
		}
	}
}
