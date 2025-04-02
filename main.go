package main

import (
	"gin-temp/routers"
)

func main() {

	app := routers.App()
	app.Run()

}
