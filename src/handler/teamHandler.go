package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamHandler struct {
	Handler Handler
}

type RegisterInput struct {
	TeamName     string `json:"teamname" binding:"required"`
	TeamPassword string `json:"password" binding:"required"`
}

func NewTeamHandler(logger zap.Logger) *TeamHandler {
	return &TeamHandler{
		Handler: *NewHandler(logger),
	}
}

// TODO create full registration process
func (th TeamHandler) RegisterTeam(ctx *gin.Context) {
	var input RegisterInput

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
