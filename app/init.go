package app

import (
	"gin-temp/controller"
	"gin-temp/middleware"
	"gin-temp/util"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	handlerFuncListOfAuthAndRbac = []gin.HandlerFunc{
		middleware.JwtAuthorization(),
		middleware.CasbinAuthorization(util.Enforcer),
	}

	userCtrl = &controller.UserController{DB: util.DB}
)

func initApp() *gin.Engine {
	util.GinLogger(util.Logger)

	return gin.Default()
}

func initValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("password",
			controller.PasswordValidator); err != nil {
			panic("failed to register password validator")
		}
		if err := v.RegisterValidation("username",
			controller.UsernameValidator); err != nil {
			panic("failed to register username validator")
		}
	}
}

func CleanUp(closers ...io.Closer) {
	for _, closer := range closers {
		if closer != nil {
			closer.Close()
		}
	}
}
