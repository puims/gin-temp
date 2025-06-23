package main

import (
	"gin-temp/router"
)

func main() {
	app, db, policyLoader := router.SetupApp()
	defer db.Close()
	defer policyLoader.Close()

	app.Run(":8080")
}
