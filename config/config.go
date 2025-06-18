package config

import (
	"os"

	"github.com/spf13/viper"
)

var Viper = viper.New()

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	Viper.AddConfigPath(home)
	Viper.SetConfigType("yaml")
	Viper.SetConfigName(".gin-temp")

	err = Viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
