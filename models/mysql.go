package models

import (
	"fmt"
	"gin-temp/config"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	DB  *gorm.DB
	err error
)

func init() {
	user := config.Viper.GetString("mysql.user")
	pwd := config.Viper.GetString("mysql.password")
	host := config.Viper.GetString("mysql.host")
	port := config.Viper.GetString("mysql.port")
	dbname := config.Viper.GetString("mysql.db")

	sqlUri := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pwd, host, port, dbname,
	)

	DB, err = gorm.Open("mysql", sqlUri)
	if err != nil {
		panic("failed to connect mysql...")
	}

	// 数据自动迁移
	DB.AutoMigrate(&UserInfo{})

	DB.DB().SetMaxOpenConns(100)
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetConnMaxLifetime(10 * time.Second)
}
