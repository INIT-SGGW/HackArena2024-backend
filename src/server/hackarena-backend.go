package main

import (
	"INIT-SGGW/hackarena-backend/handler"
	"INIT-SGGW/hackarena-backend/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	TeamHandler := handler.NewTeamHandler(*logger)

	r := gin.Default()
	authGroup := r.Group("/api/v1")
	// CORS middleware config
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With", "Hack-Arena-API-Key", "Connection"}

	authGroup.Use(cors.New(corsConfig))
	authGroup.Use(repository.AuthMiddleweare())
	repository.ConnectDataBase()

	repository.SyncDB() // DBAutoMigration

	authGroup.OPTIONS("/register", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/login", func(ctx *gin.Context) {
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

	//TODO endpoint for returning users
	//authGroup.GET("/api/v1/:teamName", TeamHandler.ReteiveUsers)

	r.GET("/hearthbeat", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "I'am alive",
		})
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080

}
