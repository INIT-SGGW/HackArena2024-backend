package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamHandler struct {
	Handler Handler
}

func NewTeamHandler(logger *zap.Logger) *TeamHandler {
	return &TeamHandler{
		Handler: *NewHandler(logger),
	}
}

func (th TeamHandler) RetreiveTeam(ctx *gin.Context) {
	defer th.Handler.logger.Sync()

	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Dummy endpoint RetreiveTeam",
	})
}
