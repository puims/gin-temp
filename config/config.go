package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/viper"
)

var Viper = viper.New()

func init() {
	generateFilesToHome()
	initViper()
}

func initViper() {
	confPath := getConfPath()

	Viper.AddConfigPath(confPath)
	Viper.SetConfigType("yaml")
	Viper.SetConfigName(".gin-temp")

	err := Viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	Viper.WatchConfig()
}

func generateFilesToHome() {
	fileList, err := getFiles()
	if err != nil {
		log.Println(err)
		return
	}

	confPath := getConfPath()
	if _, err := os.Stat(confPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(confPath); err != nil {
			log.Panicln(err)
			return
		}
	}
	if err := os.MkdirAll(confPath, 0755); err != nil {
		log.Println(err)
		return
	}

	for _, fileName := range fileList {
		src, err := os.Open(fmt.Sprintf("./config/%s", fileName))
		if err != nil {
			log.Print(err)
			return
		}
		defer src.Close()

		dst, err := os.Create(fmt.Sprintf("%s/%s", confPath, fileName))
		if err != nil {
			log.Println(err)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			log.Println(err)
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

func getConfPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return ""
	}

	return fmt.Sprintf("%s/.%s", home, "gin-temp")
}
