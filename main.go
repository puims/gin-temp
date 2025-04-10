package main

import (
	"gin-temp/models"
	"gin-temp/routers"
)

func main() {
	defer models.DB.Close()

	app := routers.App()
	app.Run()

}
