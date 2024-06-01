package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"errors"
	"fmt"
	"net/http"
	"os"
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
		ctx.JSON(http.StatusCreated, gin.H{"message": "File was sucesfully overwritten",
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

	ctx.JSON(http.StatusCreated, gin.H{"message": "File was sucesfully uploaded and saved in DB",
		"fileName": file.Filename,
	})

}

func (fh FileHandler) DownloadFiles(ctx *gin.Context) {
	teamName := ctx.Param("teamname")
	var team model.Team
	var file model.File

	row := repository.DB.Select("id").Where("team_name = ?", teamName).Find(&team)

	if team.ID == 0 || row.Error != nil {
		fh.Handler.logger.Error("Invalid team name")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find team for the teamname",
			"teamName": teamName,
		})
		return
	}

	row = repository.DB.Select("file_path").Where("team_id = ?", team.ID).Find(&file)
	if errors.Is(row.Error, gorm.ErrRecordNotFound) {
		fh.Handler.logger.Error("Invalid team name")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find file for the teamname",
			"teamName": teamName,
		})
		return
	}
	fh.Handler.logger.Info("Start processing file",
		zap.String("teamName", teamName),
		zap.String("filePath", file.FilePath))

	fileData, err := os.Open(file.FilePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileData.Close()

	fileHeader := make([]byte, 512)
	_, err = fileData.Read(fileHeader)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	fileContentType := http.DetectContentType(fileHeader)
	//Get the file info
	fileInfo, err := fileData.Stat()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", teamName))
	ctx.Header("Content-Type", fileContentType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	ctx.File(file.FilePath)

	ctx.String(http.StatusOK, "Endpoint sucesfully reached")
}
