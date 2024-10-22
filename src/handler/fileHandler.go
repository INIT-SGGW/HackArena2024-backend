package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"INIT-SGGW/hackarena-backend/service"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	// if file exist do nothing as the file is already overwritten, just end the process and send succes message
	err = repository.DB.Model(&model.SolutionFile{}).Where("team_id = ?", teamID).First(&model.SolutionFile{}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		fh.Handler.Logger.Error("Error retreiving team from database ")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Error retreiving team from database",
			"teamName": teamName,
			"file":     file.Filename,
		})
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fh.Handler.Logger.Info("There is no filedata in database")
		fh.Handler.Logger.Info("Update team solution send status",
			zap.String("teamName", teamName))

		// update team status
		err = repository.DB.Model(&model.Team{}).Preload("File").Where("id = ?", teamID).Update("is_solution_send", true).Error
		if err != nil {
			fh.Handler.Logger.Error("Error during team update ")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    "error during team update",
				"teamName": teamName,
				"file":     file.Filename,
			})
			return
		}
		// Add file
		fh.Handler.Logger.Info("Adding file to database")
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
		fh.Handler.Logger.Info("File sucesfully added")

	} else {
		fh.Handler.Logger.Info("There is already file in databse for team, skip inserts",
			zap.String("teamName", teamName))
	}

	ctx.AbortWithStatus(http.StatusCreated)

}
