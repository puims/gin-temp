package middleware

import (
	"gin-temp/utils"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func JwtAuthorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader[:7] == "Bearer " {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "failed to get token from header",
			})
			return
		}
		token := authHeader[7:]

		claims, err := utils.ParseToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		// if !models.VerifyToken(claims.ID, token) {
		// 	ctx.AbortWithStatusJSON(401, gin.H{"error": "Token expired or invalid"})
		// 	return
		// }

		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func CasbinAuthorization(enforer *casbin.Enforcer) gin.HandlerFunc {
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

		roles := claims.(*utils.Claims).Roles
		hasPermission := false
		for _, rl := range roles {
			ok, err := enforer.Enforce(rl.Name, obj, act)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Permission validation failed",
				})
				return
			}
			if ok {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No permission"})
			return
		}

		ctx.Next()
	}
}
