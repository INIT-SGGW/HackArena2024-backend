package main

import (
	"INIT-SGGW/hackarena-backend/handler"
	"INIT-SGGW/hackarena-backend/repository"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	TeamHandler := handler.NewTeamHandler(*logger)

	r := gin.Default()
	authGroup := r.Group("/api/v1")
	authGroup.Use(AuthMiddleweare())
	repository.ConnectDataBase()
	repository.SyncDB()

	//TODO logowanie
	//authGroup.GET("/login",userHandler.LoginUser)

	//TODO rejestracja
	authGroup.POST("/register", TeamHandler.RegisterTeam)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080

}
