package main

import (
	"github.com/gin-gonic/gin"
)

func AuthMiddleweare() gin.HandlerFunc {

	//TODO: JWT token session authentication

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}
