package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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

	teamName := ctx.Param("teamname")
	//var team model.Team

	//Check if session have access to the resource
	cookieTeam, _ := ctx.Get("team")
	hasAccessTo := strings.ToLower(cookieTeam.(model.Team).TeamName)
	if hasAccessTo != strings.ToLower(teamName) {
		fh.Handler.logger.Error("User have no access to this team")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "This user have no acces to this team",
			"teamName": teamName})
		return
	}
	fh.Handler.logger.Info("User have acces to the resource")

	file, err := ctx.FormFile("file")

	if err != nil {
		fh.Handler.logger.Error("Input file error")
		ctx.JSON(http.StatusNoContent, gin.H{"error": err.Error()})
		return
	}
	ext := regexp.MustCompile(`.(?:r\d\d|r\d\d\d|zip)`)

	if !ext.MatchString(file.Filename) {
		fh.Handler.logger.Error("Wrong extension error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "The file do not have .zip extension"})
		return
	}

	fh.Handler.logger.Info("File sucesfully read from input",
		zap.String("fileName", file.Filename))

	dst := fmt.Sprintf("%s/%s.zip", repository.Config.FilePath, teamName)

	// Upload the file to specific path
	err = ctx.SaveUploadedFile(file, dst)
	if err != nil {
		fh.Handler.logger.Error("Save file error")
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	fh.Handler.logger.Info("File sucesfully saved",
		zap.String("savedPath", dst))

	ctx.JSON(http.StatusOK, gin.H{"message": "File was sucesfully uploaded",
		"fileName": file.Filename,
	})

	// ctx.String(http.StatusAccepted, "Endpoint was reached")
}
