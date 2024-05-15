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

type TeamHandler struct {
	Handler Handler
}

type TeamCredential struct {
	TeamName     string `json:"teamname" binding:"required"`
	TeamPassword string `json:"password" binding:"required"`
}

func NewTeamHandler(logger zap.Logger) *TeamHandler {
	return &TeamHandler{
		Handler: *NewHandler(logger),
	}
}

func (th TeamHandler) RegisterTeam(ctx *gin.Context) {
	var input TeamCredential

	if err := ctx.ShouldBindJSON(&input); err != nil {
		th.Handler.logger.Error("Register team error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := repository.HashPassword(input.TeamPassword)
	if err != nil {
		th.Handler.logger.Error("Hash password error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	th.Handler.logger.Info("JSON input is valid")
	team := &model.Team{TeamName: input.TeamName, Password: hash}

	result := repository.DB.Create(&team)
	if result.Error != nil {
		th.Handler.logger.Error("Cannot craete new Team")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
		return
	}
	th.Handler.logger.Info("Sucesfully created team")
	ctx.JSON(http.StatusCreated, gin.H{"message": "Sucesfully created team", "TeamName": team.TeamName})
}

func (th TeamHandler) LoginTeam(ctx *gin.Context) {
	var input TeamCredential

	if err := ctx.ShouldBindJSON(&input); err != nil {
		th.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var team model.Team
	repository.DB.First(&team, "team_name = ?", input.TeamName)

	if team.ID == 0 {
		th.Handler.logger.Info("Invalid team name")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or Team Name",
		})
		return
	}
	//Validate provided password
	isValid := repository.CheckPasswordHash(input.TeamPassword, team.Password)
	if !isValid {
		th.Handler.logger.Error("Invalid password")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or Team Name",
		})
		return
	}
	//create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": team.ID,
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

	th.Handler.logger.Info("JWT token created")
	//Add cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", tokenString, 3600*24, "", "", false, true)

	th.Handler.logger.Info("Sucesfully log in")
	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Correct password",
	})

}
