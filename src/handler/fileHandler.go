package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"INIT-SGGW/hackarena-backend/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FileHandler struct {
	Handler                      Handler
	pathToSolutionFileStorage    string
	pathToMatchFileStorage       string
	pathToAllSolutionTempStorage string
	service                      *service.FileService
}

func NewFileHandler(Logger *zap.Logger, pathToSolutionFileStorage string) *FileHandler {
	defer Logger.Sync()
	pathToMatchFileStorage, exist := os.LookupEnv("HA_ADMIN_FILE_STORAGE")
	if !exist {
		Logger.Error("The HA_ADMIN_FILE_STORAGE environmental variable is missing")
		os.Exit(2)
	}
	pathToAllSolutionTempStorage, exist := os.LookupEnv("HA_ALL_FILE_STORAGE")
	if !exist {
		Logger.Error("The HA_ALL_FILE_STORAGE environmental variable is missing")
		os.Exit(2)
	}
	return &FileHandler{
		Handler:                      *NewHandler(Logger),
		service:                      service.NewFileService(Logger, pathToSolutionFileStorage, pathToAllSolutionTempStorage),
		pathToSolutionFileStorage:    pathToSolutionFileStorage,
		pathToMatchFileStorage:       pathToMatchFileStorage,
		pathToAllSolutionTempStorage: pathToAllSolutionTempStorage,
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
		fh.Handler.Logger.Info("There is already file in databse for team, update the data",
			zap.String("teamName", teamName))
		fileModel.FileName = dst
		err = repository.DB.Model(&model.SolutionFile{}).Where("team_id = ?", teamID).Update("file_name", dst).Error
		if err != nil {
			fh.Handler.Logger.Error("Error during DB save")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    "Error during updating file name data",
				"teamName": teamName,
				"file":     file.Filename,
			})
			return
		}
		fh.Handler.Logger.Info("File sucesfully added")
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

func (fh FileHandler) DownloadSingleMatchFile(ctx *gin.Context) {
	defer fh.Handler.Logger.Sync()

	teamName := ctx.Param("teamname")
	var team model.Team
	var file model.MatchFile

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
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", teamName))
	ctx.Header("Content-Type", fileContentType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	ctx.File(file.FileName)

	ctx.String(http.StatusOK, "Endpoint sucesfully reached")
}

// Check Match status endpoints

func (fh FileHandler) UserCheckMatchFile(ctx *gin.Context) {
	defer fh.Handler.Logger.Sync()

	teamID := ctx.MustGet("team_id").(uint)
	teamVerificationToken := ctx.MustGet("team_verification_token").(string)
	response := &model.CheckMatchResponse{
		IsMatchFieldExist: false,
		MatchFileName:     "",
	}

	var matchFile model.MatchFile
	err := repository.DB.Model(&model.MatchFile{}).Select("id,file_name").Where("team_id = ?", teamID).First(&matchFile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fh.Handler.Logger.Error("Error in retreiving data from database",
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed in databse connection"})
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		jsonBody, err := json.Marshal(response)
		if err != nil {
			fh.Handler.Logger.Error("Error marshaling response")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Response marshall failed",
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", jsonBody)
		return
	}

	response = &model.CheckMatchResponse{
		IsMatchFieldExist: true,
		MatchFileName:     fmt.Sprintf("%s.json", teamVerificationToken),
	}
	jsonBody, err := json.Marshal(response)
	if err != nil {
		fh.Handler.Logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshall failed",
		})
		return
	}

	ctx.Data(http.StatusOK, "application/json", jsonBody)
}

func (fh FileHandler) AdminCheckMatchFile(ctx *gin.Context) {
	defer fh.Handler.Logger.Sync()

	teamName := ctx.Param("teamname")
	response := &model.CheckMatchResponse{
		IsMatchFieldExist: false,
		MatchFileName:     "",
	}

	var team model.Team
	err := repository.DB.Model(&model.Team{}).Where("teams.team_name = ?", teamName).Preload("MatchFile").First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || team.ID < 1 {
		fh.Handler.Logger.Error("There is no such team in databse",
			zap.String("teamName", teamName))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":        "There is nos such team in database",
			"sendTeamName": teamName,
		})
		return
	}
	if err != nil {
		fh.Handler.Logger.Error("Error retreiving team data from database")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database query fail failed",
		})
		return
	}
	if team.MatchFile.TeamID != team.ID {
		jsonBody, err := json.Marshal(response)
		if err != nil {
			fh.Handler.Logger.Error("Error marshaling response")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Response marshall failed",
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", jsonBody)
		return
	}

	response = &model.CheckMatchResponse{
		IsMatchFieldExist: true,
		MatchFileName:     fmt.Sprintf("%s.json", team.VerificationToken),
	}
	jsonBody, err := json.Marshal(response)
	if err != nil {
		fh.Handler.Logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshall failed",
		})
		return
	}

	ctx.Data(http.StatusOK, "application/json", jsonBody)
}

func (fh FileHandler) GetAllFiles(ctx *gin.Context) {
	defer fh.Handler.Logger.Sync()

	teams := []model.Team{}
	err := repository.DB.Model(&model.Team{}).Where("is_solution_send = true").Preload("SolutionFile").Find(&teams).Error
	if err != nil {
		fh.Handler.Logger.Error("Error querying database")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database query error",
		})
		return
	}
	fh.Handler.Logger.Info("Sucesfully retreive teams from database")
	t := time.Now()

	directoryToZip := fmt.Sprintf("%s/temp-%s", fh.pathToAllSolutionTempStorage, t.Format(time.Kitchen))

	err = os.Mkdir(directoryToZip, 0755)
	if err != nil {
		fh.Handler.Logger.Error("Error creating temp directory")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Directory create error",
		})
		return
	}
	fh.Handler.Logger.Info("Create temp directory")

	for _, team := range teams {
		if team.SolutionFile.ID > 0 {
			srcFile := team.SolutionFile.FileName
			dstFile := fmt.Sprintf("%s/%s.zip", directoryToZip, team.TeamName)
			fh.Handler.Logger.Info("Start copying the file",
				zap.String("srcFile", srcFile),
				zap.String("dstFile", dstFile))
			err = fh.service.CopyFile(srcFile, dstFile)
			if err != nil {
				fh.Handler.Logger.Error("Error copying the file",
					zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Copy file error",
					"srcFile": "team.SolutionFile.FileName",
					"dstFile": fmt.Sprintf("%s/%s", directoryToZip, team.TeamName),
				})
				return
			}
			fh.Handler.Logger.Info("File sucesfully copied",
				zap.String("srcFile", srcFile),
				zap.String("dstFile", dstFile))
		}
	}
	fh.Handler.Logger.Info("All files sucesfully copied")

	fh.Handler.Logger.Info("Start archiving files",
		zap.String("dir", directoryToZip))

	archivePath := fmt.Sprintf("%s/solutions-%s", fh.pathToAllSolutionTempStorage, t.Format(time.Kitchen))
	err = fh.service.ZipArchive(directoryToZip, archivePath)
	if err != nil {
		fh.Handler.Logger.Error("Error archiving directory",
			zap.String("dir", directoryToZip),
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Zip archive error",
		})
		return
	}
	fh.Handler.Logger.Info("Files sucesfully archive",
		zap.String("archive", archivePath))

	// Add attachment to the reponse
	fileData, err := os.Open(archivePath)
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
	ctx.Header("Content-Disposition", "attachment; filename=solutions.zip")
	ctx.Header("Content-Type", fileContentType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	ctx.File(archivePath)

	ctx.String(http.StatusOK, "All solutions zip sucesfully send")

}
