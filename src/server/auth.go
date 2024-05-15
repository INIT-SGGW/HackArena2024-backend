package main

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
