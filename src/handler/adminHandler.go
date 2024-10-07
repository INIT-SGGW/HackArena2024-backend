package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AdminHandler struct {
	Handler Handler
}

func NewAdminHandler(logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		Handler: *NewHandler(logger),
	}
}

func (ah AdminHandler) RegisterAdmin(ctx *gin.Context) {
	defer ah.Handler.logger.Sync()

	var adminRequest model.RegisterAdminRequest

	if err := ctx.ShouldBindJSON(&adminRequest); err != nil {
		ah.Handler.logger.Error("Register admin error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ah.Handler.logger.Info("JSON input is valid")

	hash, err := repository.HashPassword(adminRequest.Password)
	if err != nil {
		ah.Handler.logger.Error("Error when hashing password",
			zap.String("email", adminRequest.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Hashing password error",
			"email": adminRequest.Email})
	}

	admin := &model.HackArenaAdmin{
		Name:     adminRequest.Name,
		Email:    adminRequest.Email,
		User:     adminRequest.UserName,
		Password: hash,
	}
	result := repository.DB.Create(&admin)
	if result.Error != nil {
		ah.Handler.logger.Error("Cannot craete new admin")
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot create new admin, duplicate"})
		return
	}
	ah.Handler.logger.Info("Admin sucesfully created",
		zap.String("admin", admin.User))

	ctx.AbortWithStatus(200)
}

func (ah AdminHandler) LoginAdmin(ctx *gin.Context) {
	defer ah.Handler.logger.Sync()

	var loginRequest model.LoginAdminRequest

	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		ah.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//retreive password from database
	var dbObject model.HackArenaAdmin
	row := repository.DB.Table("hack_arena_admins").Where("hack_arena_admins.user = ?", loginRequest.User).
		Select([]string{"id", "privilage", "password"}).Find(&dbObject)

	if row.Error != nil {
		ah.Handler.logger.Info("Invalid user")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or user",
		})
		return
	}
	ah.Handler.logger.Info("Find User",
		zap.String("Hash", dbObject.Password))

	//Validate provided password
	isValid := repository.CheckPasswordHash(loginRequest.Password, dbObject.Password)
	if !isValid {
		ah.Handler.logger.Error("Invalid password")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or user",
		})
		return
	}
	ah.Handler.logger.Info("The user is sucesfully authenticated")

	//create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": dbObject.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	//sign token
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_JWT")))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	ah.Handler.logger.Info("JWT token created")
	//Add cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("HACK-Arena-Admin-Authorization", tokenString, 3600*24, "", "", false, true)

	ah.Handler.logger.Info("Sucesfully log in")
	ctx.String(http.StatusAccepted, "Sucesfully log in user")

}
