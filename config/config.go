package config

import (
	"fmt"
	"os"
	// "github.com/spf13/viper"
)

// var Config = viper.New()

func init() {

	cfgFiles, err := getCfgFiles()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfgFiles)

	/*
		Traverses the configuration file and reads the contents of the file
		遍历配置文件，并读取文件内容
	*/

}

func getCfgFiles() (cfgFiles []string, err error) {
	/*
		Gets the files which format is .yaml from the current directory,and
		generate slices.
		获取当前目录下格式为.yaml的文件，并生成切片
	*/
	f, err := os.Open("./config")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	files, err := f.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Name()[len(file.Name())-5:] == ".yaml" {
			cfgFiles = append(cfgFiles, file.Name())
		}
	}
	return
}
