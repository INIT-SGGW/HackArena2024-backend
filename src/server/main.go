package main

import (
	"INIT-SGGW/hackarena-backend/model"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	var user model.Team
	r := gin.Default()
	authGroup := r.Group("/api/v1")
	authGroup.Use(AuthMiddleweare())

	//TODO logowanie
	//authGroup.GET("/login",userHandler.LoginUser)

	//TODO rejestracja
	//authGroup.Get("/register",userHandler.RegisterUser)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	fmt.Println(user)
	r.Run(":8080") // listen and serve on 0.0.0.0:8080

}
