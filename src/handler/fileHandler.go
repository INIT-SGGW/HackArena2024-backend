package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"INIT-SGGW/hackarena-backend/service"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FileHandler struct {
	Handler                   Handler
	pathToSolutionFileStorage string
	pathToMatchFileStorage    string
	service                   *service.FileService
}

func NewFileHandler(Logger *zap.Logger, pathToSolutionFileStorage string) *FileHandler {
	defer Logger.Sync()
	pathToMatchFileStorage, exist := os.LookupEnv("HA_ADMIN_FILE_STORAGE")
	if !exist {
		Logger.Error("The HA_ADMIN_FILE_STORAGE environmental variable is missing")
		os.Exit(2)
	}
	return &FileHandler{
		Handler:                   *NewHandler(Logger),
		service:                   service.NewFileService(Logger, pathToSolutionFileStorage),
		pathToSolutionFileStorage: pathToSolutionFileStorage,
		pathToMatchFileStorage:    pathToMatchFileStorage,
	}
}

func (fh FileHandler) UploadSolutionFile(ctx *gin.Context) {
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

	dst := fmt.Sprintf("%s/%s.zip", fh.pathToSolutionFileStorage, teamVerificationToken)

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
		err = repository.DB.Model(&model.Team{}).Preload("SolutionFile").Where("id = ?", teamID).Update("is_solution_send", true).Error
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

// TODO Change declaration of this to function to use UploadFileToDB interface and just call it instead of two method from handler
func (fh FileHandler) UploadMatchFile(ctx *gin.Context) {
	defer fh.Handler.Logger.Sync()

	teamName := ctx.Param("teamname")

	var team model.Team
	err := repository.DB.Model(&model.Team{}).Select("id,team_name,verification_token").Where("team_name = ?", teamName).First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fh.Handler.Logger.Error("There is no such team in database")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),
			"providedTeamName": teamName})
		return
	}
	if err != nil {
		fh.Handler.Logger.Error("Error retreiving team data from database")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var fileModel model.MatchFile

	file, err := ctx.FormFile("file")
	if err != nil {
		fh.Handler.Logger.Error("There is no file in form")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check file extension
	ext := regexp.MustCompile(`.(?:r\d\d|r\d\d\d|json)`)
	if !ext.MatchString(file.Filename) {
		fh.Handler.Logger.Error("Wrong extension error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "The file do not have .json extension"})
		return
	}

	fh.Handler.Logger.Info("File sucesfully read from input",
		zap.String("fileName", file.Filename))

	dst := fmt.Sprintf("%s/%s.json", fh.pathToMatchFileStorage, team.VerificationToken)

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
	err = repository.DB.Model(&model.MatchFile{}).Where("team_id = ?", team.ID).First(&model.MatchFile{}).Error
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
		// Add file
		fh.Handler.Logger.Info("Adding file to database")
		fileModel.TeamID = team.ID
		fileModel.FileName = dst
		err = repository.DB.Model(&model.MatchFile{}).Save(&fileModel).Error
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

func (fh FileHandler) DownloadSingleSolutionFile(ctx *gin.Context) {
	defer fh.Handler.Logger.Sync()

	teamName := ctx.Param("teamname")
	var team model.Team
	var file model.SolutionFile

	row := repository.DB.Select("id").Where("team_name = ?", teamName).Find(&team)

	if team.ID == 0 || row.Error != nil {
		fh.Handler.Logger.Error("Invalid team name")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    "Cannot find team for the provided teamname",
			"teamName": teamName,
		})
		return
	}

	row = repository.DB.Select("file_name").Where("team_id = ?", team.ID).Find(&file)
	if errors.Is(row.Error, gorm.ErrRecordNotFound) {
		fh.Handler.Logger.Error("Invalid team name")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    "Cannot find file for the teamname",
			"teamName": teamName,
		})
		return
	}
	fh.Handler.Logger.Info("Start processing file",
		zap.String("teamName", teamName),
		zap.String("filePath", file.FileName))

	fileData, err := os.Open(file.FileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file descriptor"})
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
	ctx.File(file.FileName)

	ctx.String(http.StatusOK, "Endpoint sucesfully reached")
}
