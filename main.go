package main

import (
	"gin-temp/router"
)

func main() {
	app, db, loader := router.SetupApp()
	defer db.Close()
	defer loader.Close()

	app.Run(":8080")
}
