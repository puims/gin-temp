package middleware

import (
	"gin-temp/models"
	"gin-temp/utils"

	"github.com/gin-gonic/gin"
)

func Authorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := utils.GetClaimsWithToken(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		token, err := utils.GetToken(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
			return
		}

		if !models.VerifyToken(claims.ID, token) {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Token expired or invalid"})
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
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		allowed := false
		for _, role := range roles {
			for _, userRole := range claims.Roles {
				if role == userRole.Name {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			ctx.AbortWithStatusJSON(403,
				gin.H{"error": "Insufficient permissions"})
			return
		}

		ctx.Next()
	}
}
