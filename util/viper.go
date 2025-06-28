package util

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func setupViper() *viper.Viper {
	confPath := getConfPath()

	vp := viper.New()
	vp.AddConfigPath(confPath)
	vp.SetConfigType("yaml")
	vp.SetConfigName(".config")

	err := vp.ReadInConfig()
	if err != nil {
		panic(err)
	}

	vp.WatchConfig()

	vp.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
		if err := vp.ReadInConfig(); err != nil {
			log.Println("Error reloading config:", err)
		}
		// 重新初始化依赖配置的组件
		initJWTKeys()
	})
	return vp
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

		dst, err := os.Create(fmt.Sprintf("%s/%s", confPath, fileName))
		if err != nil {
			log.Println(err)
			return
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			log.Println(err)
			return
		}

		src.Close()
		dst.Close()
	}
}

func getFiles() (files []string, err error) {
	dir, err := os.ReadDir("./config/")
	if err != nil {
		return nil, err
	}

	for _, file := range dir {
		name := file.Name()
		if len(name) > 0 && name[0] == '.' {
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

	return fmt.Sprintf("%s/.%s", home, AppName)
}
