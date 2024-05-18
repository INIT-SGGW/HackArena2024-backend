package repository

import (
	"os"

	"github.com/gin-gonic/gin"
)

func AuthMiddleweare() gin.HandlerFunc {

	return func(c *gin.Context) {
		apiKey := c.GetHeader("Hack-Arena-API-Key")
		keyValue := os.Getenv("HA_API_KEY")
		if apiKey != keyValue {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

// func CORSMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Hack-Arena-API-Key,Connection")
// 		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

// 		if ctx.Request.Method == "OPTIONS" {
// 			ctx.AbortWithStatus(204)
// 			return
// 		}

// 		ctx.Next()
// 	}
// }
