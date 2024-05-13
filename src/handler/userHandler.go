package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	Handler Handler
}

type RegisterInput struct {
	TeamName     string `json:"teamname" binding:"required"`
	TeamPassword string `json:"teampassword" binding:"required"`
}

func NewUserHandler(logger zap.Logger) *UserHandler {
	return &UserHandler{
		Handler: *NewHandler(logger),
	}
}

// TODO create full registration process
func (uh UserHandler) RegisterUser(ctx *gin.Context) {
	var input RegisterInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		uh.Handler.logger.Error("Register team error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uh.Handler.logger.Info("Sucesfully authorize")
	ctx.JSON(http.StatusOK, gin.H{"message": "This is authorize register endpoint"})
}
