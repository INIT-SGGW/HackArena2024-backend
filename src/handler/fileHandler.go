package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FileHandler struct {
	Handler Handler
}

func NewFileHandler(logger zap.Logger) *FileHandler {
	return &FileHandler{
		Handler: *NewHandler(logger),
	}
}

func (fh FileHandler) UploadFile(ctx *gin.Context) {
	ctx.String(http.StatusAccepted, "Endpoint was reached")
}
