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
	optionsGroup := r.Group("/api/v1")
	authGroup.Use(repository.AuthMiddleweare())
	optionsGroup.Use(repository.CORSMiddleware())
	repository.ConnectDataBase()
	repository.SyncDB() // DBAutoMigration

	optionsGroup.OPTIONS("/register", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	optionsGroup.OPTIONS("/login", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})

	authGroup.POST("/register", TeamHandler.RegisterTeam)

	authGroup.POST("/login", TeamHandler.LoginUser)

	//TODO endpoint to add users to the team
	//authGroup.POST("/:team/adduser",TeamHandler.AddUser)

	//TODO endpoint to update user data
	//authGroup.PUT("/:team/:email:",TeamHandler.UpdateUser)

	r.GET("/hearthbeat", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(200, gin.H{
			"message": "I'am alive",
		})
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080

}
