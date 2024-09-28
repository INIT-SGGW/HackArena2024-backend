package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type UserAccountHandler struct {
	Handler Handler
}

func NewUserAccountHandler(logger *zap.Logger) *UserAccountHandler {
	return &UserAccountHandler{
		Handler: *NewHandler(logger),
	}
}

func (uh UserAccountHandler) LoginUser(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	var input model.LoginRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		uh.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//retreive password from database
	var dbObject model.Member
	row := repository.DB.Table("members").Where("email = ?", input.Email).
		Select([]string{"id", "email", "password"}).Find(&dbObject)

	if row.Error != nil {
		uh.Handler.logger.Info("Invalid email")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or email",
		})
		return
	}
	//Validate provided password
	isValid := repository.CheckPasswordHash(input.Password, dbObject.Password)
	if !isValid {
		uh.Handler.logger.Error("Invalid password")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or email",
		})
		return
	}
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

	uh.Handler.logger.Info("JWT token created")
	//Add cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("HACK-Arena-Authorization", tokenString, 3600*24, "", "", false, true)

	var team model.Team
	result := repository.DB.Model(&model.Team{}).Select("teams.team_name").Joins("inner join members on members.team_id = teams.id").Where("members.email = ? ", dbObject.Email).First(&team)

	if result.Error != nil {
		uh.Handler.logger.Error("Error marshaling response",
			zap.String("email", dbObject.Email))

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to retreive the team",
		})
		return
	}
	jsonBody, err := json.Marshal(model.LoginResponse{
		TeamName: team.TeamName,
		Email:    dbObject.Email,
	})
	if err != nil {
		uh.Handler.logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	uh.Handler.logger.Info("Sucesfully log in")
	ctx.Data(http.StatusAccepted, "application/json", jsonBody)

}

//Chanhe to member not team pasword

func (uh UserAccountHandler) ChangePassword(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	teamName := ctx.Param("teamname")
	var newCredentials model.ChangePasswordRequest
	var team model.Team

	//Check if session have access to the resource
	cookieTeam, _ := ctx.Get("team")
	hasAccessTo := strings.ToLower(cookieTeam.(model.Team).TeamName)
	if hasAccessTo != strings.ToLower(teamName) {
		uh.Handler.logger.Error("User have no access to this team")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "This user have no acces to this team",
			"teamName": teamName})
		return
	}
	uh.Handler.logger.Info("User have acces to the resource")

	if err := ctx.ShouldBindJSON(&newCredentials); err != nil {
		uh.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uh.Handler.logger.Info("Body was sucesfully binded")

	row := repository.DB.Select("team_name", "id").Where("team_name = ?", teamName).Find(&team)
	uh.Handler.logger.Info("Retreive following team from DB",
		zap.String("teamName", team.TeamName),
		zap.Uint("team_id", team.ID))

	if team.ID == 0 || row.Error != nil {
		uh.Handler.logger.Error("Invalid team name")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find team for the teamname",
			"teamName": teamName,
		})
		return
	}

	hash, err := repository.HashPassword(newCredentials.NewPassword)
	if err != nil {
		uh.Handler.logger.Error("Hash password error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uh.Handler.logger.Info("Password was Hashed")
	team.TeamName = hash
	repository.DB.Model(&team).Updates(team)

	ctx.JSON(http.StatusAccepted, gin.H{
		"message":  "Sucesfully updated password",
		"teamName": team.TeamName,
	})
}

func (uh UserAccountHandler) RestartForgotPassword(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Dummy endpoint RestartForgotPassword",
	})
}

func (uh UserAccountHandler) ResetPassword(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Dummy endpoint ResetPassword",
	})
}
