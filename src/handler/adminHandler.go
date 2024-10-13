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
	defer ah.Handler.Logger.Sync()

	var adminRequest model.RegisterAdminRequest

	if err := ctx.ShouldBindJSON(&adminRequest); err != nil {
		ah.Handler.Logger.Error("Register admin error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ah.Handler.Logger.Info("JSON input is valid")

	hash, err := repository.HashPassword(adminRequest.Password)
	if err != nil {
		ah.Handler.Logger.Error("Error when hashing password",
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
		ah.Handler.Logger.Error("Cannot craete new admin")
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot create new admin, duplicate"})
		return
	}
	ah.Handler.Logger.Info("Admin sucesfully created",
		zap.String("admin", admin.User))

	ctx.AbortWithStatus(200)
}

func (ah AdminHandler) LoginAdmin(ctx *gin.Context) {
	defer ah.Handler.Logger.Sync()

	var loginRequest model.LoginAdminRequest

	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		ah.Handler.Logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//retreive password from database
	var dbObject model.HackArenaAdmin
	row := repository.DB.Table("hack_arena_admins").Where("hack_arena_admins.email = ?", loginRequest.Email).
		Select([]string{"id", "privilage", "password"}).Find(&dbObject)

	if row.Error != nil {
		ah.Handler.Logger.Info("Invalid user")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or user",
		})
		return
	}
	ah.Handler.Logger.Info("Find User")

	//Validate provided password
	isValid := repository.CheckPasswordHash(loginRequest.Password, dbObject.Password)
	if !isValid {
		ah.Handler.Logger.Error("Invalid password")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or user",
		})
		return
	}
	ah.Handler.Logger.Info("The user is sucesfully authenticated")

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

	ah.Handler.Logger.Info("JWT token created")
	//Add cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("HACK-Arena-Admin-Authorization", tokenString, 3600*24, "", "", false, true)

	ah.Handler.Logger.Info("Sucesfully log in")
	ctx.String(http.StatusAccepted, "Sucesfully log in user")

}
func (ah AdminHandler) AdminApproveTeam(ctx *gin.Context) {
	defer ah.Handler.Logger.Sync()

	teamName := ctx.Param("teamname")
	var teamApproveRequest model.UpdateTeamRequest

	if teamName == "" {
		ah.Handler.Logger.Error("There is no team name in the request")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Empty team name param"})
		return
	}

	if err := ctx.ShouldBindJSON(&teamApproveRequest); err != nil {
		ah.Handler.Logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ah.Handler.Logger.Info("The input is valid query DB for team",
		zap.String("teamName", teamName))

	team := model.Team{}
	err := repository.DB.Model(&model.Team{}).Select("id,team_name,approve_status").Where("team_name = ?", teamName).First(&team).Error
	if err != nil {
		ah.Handler.Logger.Error("There is no such team name in database")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No team in database"})
		return
	}

	err = repository.DB.Model(&model.Team{}).Where("id = ? ", team.ID).Update("approve_status", teamApproveRequest.Status).Error
	if err != nil {
		ah.Handler.Logger.Error("Error inserting status to database")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Issue in inserting record in database"})
		return
	}

	ah.Handler.Logger.Info("Team approve status sucesfully updated")

	ctx.AbortWithStatus(200)

}

func (ah AdminHandler) ConfirmTeam(ctx *gin.Context) {
	defer ah.Handler.Logger.Sync()

	teamName := ctx.Param("teamname")
	if teamName == "" {
		ah.Handler.Logger.Error("Missing teamName parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Mising teamName parameter"})
		return
	}

	ah.Handler.Logger.Info("The input is valid",
		zap.String("teamName", teamName))

	team := &model.Team{}
	result := repository.DB.Select("team_name,id,is_verified").Where("team_name = ?", teamName).First(&team)
	if result.Error != nil {
		ah.Handler.Logger.Error("There is no such team in database")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "There is no such team in database"})
		return
	}
	ah.Handler.Logger.Info("Sucesfully retreive team data from database")

	err := repository.DB.Model(&model.Team{}).Where("id = ?", team.ID).Update("is_confirmed", true).Error
	if err != nil {
		ah.Handler.Logger.Error("Error inserting to database")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database insert failed",
		})
		return
	}

	ctx.AbortWithStatus(200)

}
