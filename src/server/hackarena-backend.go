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
	authGroup.Use(repository.CORSMiddleware())
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

	authGroup.OPTIONS("/:teamname/users", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/logout", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})

	authGroup.OPTIONS("/:teamname/update", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/:teamname/changepassword", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})

	authGroup.POST("/logout", func(ctx *gin.Context) {
		ctx.SetCookie("HACK-Arena-Authorization", "", -1, "", "", false, true)
		ctx.JSON(200, gin.H{
			"message": "user logout",
		})
	})
	authGroup.POST("/register", TeamHandler.RegisterTeam)

	authGroup.POST("/login", TeamHandler.LoginUser)

	authGroup.GET("/:teamname/users", repository.CookieAuth, TeamHandler.ReteiveUsers)

	authGroup.POST("/:teamname/update", repository.CookieAuth, TeamHandler.UpdeteTeam)

	authGroup.POST("/:teamname/changepassword", repository.CookieAuth, TeamHandler.ChangePassword)

	//TODO endpoint to update user data
	//authGroup.PUT("/:team/:email:",TeamHandler.UpdateUsers)

	//TODO endpoint to add users to the team
	//authGroup.POST("/:team/adduser",TeamHandler.AddUser)

	r.GET("/hearthbeat", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "I'am alive",
		})
	})

	r.Run(":8080") // listen and serve on 0.0.0.0:8080

}
