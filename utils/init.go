package utils

import (
	"gin-temp/models"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const AppName = "gin-temp"

var (
	Viper        *viper.Viper
	Enforcer     *casbin.Enforcer
	PolicyLoader *CasbinPolicyLoader
	DB           *MysqlDB
	Redis        *redis.Client
	err          error
)

func init() {
	generateFilesToHome()

	Viper = setupViper()

	DB = newMysqlDB()
	DB.mysqlMigrate(&models.User{})
	DB.ping()
	DB.addRoot()

	Enforcer, PolicyLoader, err = setupCasbin(DB.DB)
	if err != nil {
		panic("failed to init casbin")
	}

	Redis = setupRedis()
}
