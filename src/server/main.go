package main

import (
	"INIT-SGGW/hackarena-backend/handler"
	"INIT-SGGW/hackarena-backend/repository"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	TeamHandler := handler.NewTeamHandler(*logger)

	r := gin.Default()
	authGroup := r.Group("/api/v1")
	authGroup.Use(AuthMiddleweare())
	repository.ConnectDataBase()
	//repository.SyncDB() // DBAutoMigration

	// Endpoint create the team and store the team data in database (the password is hashed)
	// Dodac pole Users[] zeby wszystkich wpisywalo do bazy (1-3)
	authGroup.POST("/register", TeamHandler.RegisterTeam)

	//TODO logowanie
	//Dodac logowanie przez email
	// token JWT
	authGroup.POST("/login", TeamHandler.LoginTeam)

	//TODO
	//authGroup.POST("/:team/adduser",TeamHandler.AddUser)

	//TODO
	//authGroup.PUT("/:team/:email:",TeamHandler.UpdateUser)

	r.GET("/hearthbeat", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "I'am alive",
		})
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080

}
