package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func GinLogger(file *os.File) {
	if !Viper.GetBool("app.debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(file)
}

func createLogFile() *os.File {
	now := time.Now().Format("01-02-2006")
	fileName := fmt.Sprintf("%s-%s.log", Viper.GetString("app.name"), now)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("failed to create log file")
		return nil
	}

	return file
}
