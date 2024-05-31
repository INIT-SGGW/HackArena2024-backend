package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	var team model.Team
	var fileModel model.File

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

	// Check file extension
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

	// Save file path to database
	row := repository.DB.Select("team_name", "id").Where("team_name = ?", teamName).Find(&team)
	fh.Handler.logger.Info("Retreive following team from DB",
		zap.String("teamName", team.TeamName),
		zap.Uint("team_id", team.ID))

	if team.ID == 0 || row.Error != nil {
		fh.Handler.logger.Error("Invalid team name")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find team for the teamname",
			"teamName": teamName,
		})
		return
	}
	row = repository.DB.Where("team_id = ?", team.ID).First(&fileModel)

	if !errors.Is(row.Error, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, gin.H{"message": "File was sucesfully overwritten",
			"fileName": file.Filename,
		})
		return
	}

	fileModel.TeamID = team.ID
	fileModel.FilePath = dst
	result := repository.DB.Save(&fileModel)
	if result.Error != nil {
		fh.Handler.logger.Error("Error during DB save")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Error during DB save",
			"teamName": teamName,
			"file":     file.Filename,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "File was sucesfully uploaded and saved in DB",
		"fileName": file.Filename,
	})

}
