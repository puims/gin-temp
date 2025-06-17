package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/storyicon/grbac"
)

func Authorization() gin.HandlerFunc {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	rbac, err := grbac.New(
		grbac.WithYAML(fmt.Sprintf("%s/.gin-rbac", home), 10*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	return func(ctx *gin.Context) {
		roles, err := QueryRolesByHeaders(ctx.Request.Header)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		state, _ := rbac.IsRequestGranted(ctx.Request, roles)
		if !state.IsGranted() {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func QueryRolesByHeaders(header http.Header) (roles []string, err error) {
	return
}
