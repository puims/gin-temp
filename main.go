package main

import (
	"gin-temp/models"
	"gin-temp/routers"
)

func main() {
	defer models.DB.Close()

	routers.App().Run(":8080")

}
