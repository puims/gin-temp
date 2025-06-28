package main

import (
	"gin-temp/app"
	"gin-temp/util"
)

func main() {
	engine := app.SetupApp()
	defer app.CleanUp(util.DB, util.Redis, util.PolicyLoader, util.Logger)

	engine.Run(":8080")
}
