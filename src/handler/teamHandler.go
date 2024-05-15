package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamHandler struct {
	Handler Handler
}

type RegisterInput struct {
	TeamName     string `json:"teamname" binding:"required"`
	TeamPassword string `json:"teampassword" binding:"required"`
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

	th.Handler.logger.Info("Sucesfully authorize")
	ctx.JSON(http.StatusOK, gin.H{"message": "This is authorize register endpoint"})
}
