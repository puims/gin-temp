package util

import (
	"gin-temp/model"
	"log"
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const AppName = "gin-temp"

var (
	Viper        *viper.Viper
	Logger       *os.File
	Enforcer     *casbin.Enforcer
	PolicyLoader *CasbinPolicyLoader
	DB           *MysqlDB
	Redis        *redis.Client
	err          error
)

func init() {
	generateFilesToHome()

	Viper = setupViper()

	Logger = createLogFile()

	DB = newMysqlDB()
	DB.mysqlMigrate(&model.User{})
	DB.ping()
	DB.addRoot()

	Enforcer, PolicyLoader, err = setupCasbin(DB.DB)
	if err != nil {
		log.Fatal("failed to init casbin")
	}

	Redis = setupRedis()
}
