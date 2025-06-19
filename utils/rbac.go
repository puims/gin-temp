package utils

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/storyicon/grbac"
)

func AuthorizationRabc() gin.HandlerFunc {
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
		roles, err := QueryRolesByHeaders(ctx)
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

func QueryRolesByHeaders(ctx *gin.Context) (roles []string, err error) {
	// token, err := getToken(ctx)
	// if err != nil {
	// 	ctx.JSON(401, gin.H{"error": "authorization header required"})
	// 	ctx.Abort()
	// 	return
	// }

	// claims, exists := ctx.Get("claims")
	// if !exists {
	// 	ctx.JSON(401, gin.H{"error": "claims not found"})
	// 	return
	// }

	return
}
