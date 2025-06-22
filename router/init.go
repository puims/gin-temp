package router

import (
	"errors"
	"gin-temp/config"
	"gin-temp/models"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	gin.DefaultWriter = io.MultiWriter(f)
}

func initRoot(db *models.MysqlDB) error {
	if err := db.First(&models.User{}, "username = ?", "root").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Transaction(func(tx *gorm.DB) error {
				user := models.User{
					Username: "root",
					Password: "root",
					Role:     "root",
				}
				return tx.Save(&user).Error
			})
		}
	}
	return nil
}
