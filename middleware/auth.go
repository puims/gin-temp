package middleware

import (
	"gin-temp/utils"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func JwtAuthorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")

		isblack, err := utils.InBlackList(token, utils.Redis)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "server errors"})
			return
		}
		if isblack {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "user has logout"})
			return
		}

		if token[:7] == "Bearer " {
			token = token[7:]
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "failed to get token from header",
			})
			return
		}

		claims, err := utils.ParseToken(token, utils.JwtKey)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func CasbinAuthorization(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取请求的URI
		obj := ctx.Request.URL.RequestURI()
		// 获取请求方法
		act := ctx.Request.Method
		// 获取用户的角色
		claims, exists := ctx.Get("claims")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		sub := claims.(*utils.Claims).Role

		ok, err := enforcer.Enforce(sub, obj, act)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Permission validation failed",
			})
			return
		}
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No permission"})
			return
		}

		ctx.Next()
	}
}
