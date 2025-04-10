package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var Viper = viper.New()

func init() {
	cfgFiles, err := getCfgFiles()
	if err != nil {
		fmt.Println(err)
		return
	}

	Viper.AddConfigPath("./config")
	Viper.SetConfigType("yaml")
	/*
		Traverses the configuration file and reads the contents of the file
		遍历配置文件，并读取文件内容
	*/
	for _, file := range cfgFiles {
		Viper.SetConfigName(file)
		if err := Viper.MergeInConfig(); err != nil {
			fmt.Println(err)
			return
		}
	}

}

func getCfgFiles() (cfgFiles []string, err error) {
	/*
		Gets the files which format is .yaml from the current directory,and
		generate slices.
		获取当前目录下格式为.yaml的文件，并生成切片
	*/
	f, err := os.Open("./config")
	if err != nil {
		return
	}
	defer f.Close()

	files, err := f.ReadDir(-1)
	if err != nil {
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			cfgFiles = append(cfgFiles, file.Name())
		}
	}
	return
}
