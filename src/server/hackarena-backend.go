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
	UserAccountHandler := handler.NewUserAccountHandler(logger)
	RegisterHandler := handler.NewRegisterHandler(logger)
	TeamHandler := handler.NewTeamHandler(logger)
	AdminHandler := handler.NewAdminHandler(logger)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20
	authGroup := r.Group("/api/v2")
	authGroup.Use(repository.CORSMiddleware())
	authGroup.Use(repository.AuthMiddleweare())
	adminAuthGroup := r.Group("/api/v2/admin")
	adminAuthGroup.Use(repository.CORSMiddleware(), repository.AuthMiddleweare(), repository.AdminAuthMiddleweare())
	repository.InitializeConfig()
	repository.ConnectDataBase()

	repository.SyncDB() // DBAutoMigration

	authGroup.OPTIONS("/register/team", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/register/member", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})

	authGroup.OPTIONS("/login", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/logout", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})

	authGroup.OPTIONS("/password/reset", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/password/change", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/password/forgot", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/team/:teamname", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/team/confirmation/:teamname", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	authGroup.OPTIONS("/team/approve/:teamname", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	// Admin Endpoints Options
	adminAuthGroup.OPTIONS("/teams", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	adminAuthGroup.OPTIONS("/login", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	adminAuthGroup.OPTIONS("/logout", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	adminAuthGroup.OPTIONS("/register", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})
	adminAuthGroup.OPTIONS("/team/confirmation/:teamname", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "return headers",
		})
	})

	authGroup.POST("/register/team", RegisterHandler.RegisterTeam)

	authGroup.POST("/register/member", RegisterHandler.RegisterMember)

	authGroup.POST("/login", UserAccountHandler.LoginUser)

	authGroup.POST("/logout", func(ctx *gin.Context) {
		ctx.SetCookie("HACK-Arena-Authorization", "", -1, "", "", false, true)
		ctx.JSON(200, gin.H{
			"message": "User logout",
		})
	})

	authGroup.GET("/team/:teamname", repository.CookieAuth, TeamHandler.RetreiveTeam)

	authGroup.POST("/team/confirmation/:teamname", TeamHandler.ConfirmTeam)

	authGroup.POST("/password/forgot", UserAccountHandler.RestartForgotPassword)

	authGroup.POST("/password/change", repository.CookieAuth, UserAccountHandler.ChangePassword)

	authGroup.POST("/password/reset", UserAccountHandler.ResetPassword)

	// Admin endpoints
	adminAuthGroup.POST("/login", AdminHandler.LoginAdmin)

	adminAuthGroup.POST("/logout", func(ctx *gin.Context) {
		ctx.SetCookie("HACK-Arena-Admin-Authorization", "", -1, "", "", false, true)
		ctx.JSON(200, gin.H{
			"message": "User logout",
		})
	})

	adminAuthGroup.POST("/register", repository.AdminCookieAuth, AdminHandler.RegisterAdmin)

	adminAuthGroup.GET("/teams", repository.AdminCookieAuth, TeamHandler.GetAllTeamsAsAdmin)

	adminAuthGroup.POST("/team/approve/:teamname", repository.AdminCookieAuth, AdminHandler.AdminApproveTeam)

	adminAuthGroup.POST("/team/confirmation/:teamname", repository.AdminCookieAuth, AdminHandler.ConfirmTeam)

	// Endpoint for status check
	r.GET("/hearthbeat", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "I'am alive",
		})
	})

	r.Run(":8080") // listen and serve on 0.0.0.0:8080

}
