package middlewares

import (
	"gin-temp/models"
	"gin-temp/utils"

	"github.com/gin-gonic/gin"
)

func Authorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := utils.GetClaimsWithToken(ctx)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		token, err := utils.GetToken(ctx)
		if err != nil {
			ctx.JSON(401, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		if !models.VerifyToken(claims.ID, token) {
			ctx.JSON(401, gin.H{"error": "Token expired or invalid"})
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func RoleMidware(roles []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := utils.GetClaimsWithToken(ctx)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		allowed := false
		for _, role := range roles {
			if role == claims.Role {
				allowed = true
				break
			}
		}

		if !allowed {
			ctx.JSON(403,
				gin.H{"error": "Insufficient permissions"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
