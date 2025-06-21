package router

import (
	"gin-temp/config"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func initApp() *gin.Engine {
	if !config.Viper.GetBool("app.debug") {
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

	gin.DefaultWriter = io.MultiWriter(f, os.Stdout) // 同时输出到文件和控制台
}
