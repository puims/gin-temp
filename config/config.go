package config

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"
)

var Viper = viper.New()

func init() {
	generateFilesToHome()
	initViper()
}

func initViper() {
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

	Viper.WatchConfig()
}

func generateFilesToHome() {
	fileList, err := getFiles()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, fileName := range fileList {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			return
		}

		src, err := os.Open(fmt.Sprintf("./config/%s", fileName))
		if err != nil {
			fmt.Print(err)
			return
		}
		defer src.Close()

		dst, err := os.Create(fmt.Sprintf("%s/%s", home, fileName))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func getFiles() (files []string, err error) {
	dir, err := os.ReadDir("./config/")
	if err != nil {
		return nil, err
	}

	for _, file := range dir {
		if file.Name()[:4] == ".gin" {
			files = append(files, file.Name())
		}
	}

	return
}
