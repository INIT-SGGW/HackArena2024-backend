package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"INIT-SGGW/hackarena-backend/service"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FileHandler struct {
	Handler           Handler
	pathToFileStorage string
	service           *service.FileService
}

func NewFileHandler(Logger *zap.Logger, pathToFileStorage string) *FileHandler {
	return &FileHandler{
		Handler:           *NewHandler(Logger),
		service:           service.NewFileService(Logger, pathToFileStorage),
		pathToFileStorage: pathToFileStorage,
	}
}

func (fh FileHandler) UploadFile(ctx *gin.Context) {
	defer fh.Handler.Logger.Sync()

	teamName := ctx.MustGet("team_name").(string)
	teamID := ctx.MustGet("team_id").(uint)
	teamVerificationToken := ctx.MustGet("team_verification_token").(string)

	var fileModel model.SolutionFile

	file, err := ctx.FormFile("file")
	if err != nil {
		fh.Handler.Logger.Error("There is no file in form")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check file extension
	ext := regexp.MustCompile(`.(?:r\d\d|r\d\d\d|zip)`)
	if !ext.MatchString(file.Filename) {
		fh.Handler.Logger.Error("Wrong extension error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "The file do not have .zip extension"})
		return
	}

	fh.Handler.Logger.Info("File sucesfully read from input",
		zap.String("fileName", file.Filename))

	dst := fmt.Sprintf("%s/%s.zip", fh.pathToFileStorage, teamVerificationToken)

	// Upload the file to specific path
	err = ctx.SaveUploadedFile(file, dst)
	if err != nil {
		fh.Handler.Logger.Error("Save file error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fh.Handler.Logger.Info("File sucesfully saved to disk",
		zap.String("savedPath", dst))

	fileModel.TeamID = teamID
	fileModel.FileName = dst
	err = repository.DB.Model(&model.SolutionFile{}).Save(&fileModel).Error
	if err != nil {
		fh.Handler.Logger.Error("Error during DB save")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Error during DB save",
			"teamName": teamName,
			"file":     file.Filename,
		})
		return
	}

	ctx.AbortWithStatus(http.StatusCreated)

}
