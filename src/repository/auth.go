package repository

import (
	"INIT-SGGW/hackarena-backend/model"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func CookieAuth(ctx *gin.Context) {
	tokenString, err := ctx.Cookie("HACK-Arena-Authorization")
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SECRET_JWT")), nil
	})
	if token == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "The token has expired"})
		}

		var member model.Member
		DB.First(&member, claims["sub"])
		if member.ID == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "There is no such user"})
		}

		ctx.Set("user", member)

		ctx.Next()
	} else {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://test.hackarena.pl/")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With,Hack-Arena-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
