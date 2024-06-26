package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type TeamHandler struct {
	Handler Handler
}

// Team object send to registration
type InputTeam struct {
	TeamName     string       `json:"teamName" binding:"required"`
	TeamPassword string       `json:"password" binding:"required"`
	TeamMembers  []model.User `json:"teamMembers" binding:"required"`
}

// Team object send for update
type UpdateTeam struct {
	TeamName    string       `json:"teamName" binding:"required"`
	TeamMembers []model.User `json:"teamMembers" binding:"required"`
}

// User input on login endpoint
type UserCredential struct {
	ID       string `json:"-"`
	TeamName string `json:"-"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// User input for password change
type PasswordChange struct {
	Email       string `json:"email"`
	NewPassword string `json:"password"`
}

// Output team response
type TeamOutput struct {
	TeamName    string       `json:"teamName" binding:"required"`
	TeamMembers []model.User `json:"teamMembers" binding:"required"`
}

func NewTeamHandler(logger zap.Logger) *TeamHandler {
	return &TeamHandler{
		Handler: *NewHandler(logger),
	}
}

func (th TeamHandler) RegisterTeam(ctx *gin.Context) {
	var input InputTeam

	if err := ctx.ShouldBindJSON(&input); err != nil {
		fmt.Println(input)
		th.Handler.logger.Error("Register team error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := repository.HashPassword(input.TeamPassword)
	if err != nil {
		th.Handler.logger.Error("Hash password error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	th.Handler.logger.Info("JSON input is valid")
	team := &model.Team{TeamName: input.TeamName, Password: hash, Users: input.TeamMembers}

	result := repository.DB.Create(&team)
	if result.Error != nil {
		th.Handler.logger.Error("Cannot craete new Team")
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot create new team, duplicate"})
		return
	}
	th.Handler.logger.Info("Sucesfully created team")
	ctx.JSON(http.StatusCreated, gin.H{"message": "Sucesfully created team", "TeamName": team.TeamName})
}

// TODO spilt to single responsibility functions
func (th TeamHandler) LoginUser(ctx *gin.Context) {
	var input UserCredential

	if err := ctx.ShouldBindJSON(&input); err != nil {
		th.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//retreive password from database
	var dbObject UserCredential
	row := repository.DB.Table("users").
		Joins("INNER Join teams t ON t.id = users.team_id").Where("email = ?", input.Email).
		Select([]string{"t.id", "t.team_name", "users.email", "t.Password"}).Find(&dbObject)

	if row.Error != nil {
		th.Handler.logger.Info("Invalid team name")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or Team Name",
		})
		return
	}
	//Validate provided password
	isValid := repository.CheckPasswordHash(input.Password, dbObject.Password)
	if !isValid {
		th.Handler.logger.Error("Invalid password")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or Team Name",
		})
		return
	}
	//create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": dbObject.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	//sign token
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_JWT")))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	th.Handler.logger.Info("JWT token created")
	//Add cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("HACK-Arena-Authorization", tokenString, 3600*24, "", "", false, true)

	th.Handler.logger.Info("Sucesfully log in")
	ctx.JSON(http.StatusAccepted, gin.H{
		"message":  "Correct password",
		"teamName": dbObject.TeamName,
	})

}

func (th TeamHandler) ReteiveUsers(ctx *gin.Context) {
	teamName := ctx.Param("teamname")
	var teamOutput TeamOutput
	var team model.Team

	//Check if session have access to the resource
	cookieTeam, _ := ctx.Get("team")
	hasAccessTo := strings.ToLower(cookieTeam.(model.Team).TeamName)
	if hasAccessTo != strings.ToLower(teamName) {
		th.Handler.logger.Error("User have no access to this team")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "This user have no acces to this team",
			"teamName": teamName})
		return
	}
	th.Handler.logger.Info("User have acces to the resource")

	row := repository.DB.Select("team_name", "id").Where("team_name = ?", teamName).Find(&team)
	th.Handler.logger.Info("Retreive following team from DB",
		zap.String("teamName", team.TeamName),
		zap.Uint("team_id", team.ID))

	if team.ID == 0 || row.Error != nil {
		th.Handler.logger.Error("Invalid team name")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find team for the teamname",
			"teamName": teamName,
		})
		return
	}
	teamOutput.TeamName = team.TeamName
	var users []model.User
	repository.DB.Where("team_id = ?", team.ID).Find(&users)
	if len(users) < 1 {
		th.Handler.logger.Error("The team have 0 members")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find team members",
			"teamName": teamName})
		return
	}
	teamOutput.TeamMembers = users

	ctx.JSON(http.StatusOK, teamOutput)

}

func (th TeamHandler) UpdeteTeam(ctx *gin.Context) {

	teamName := ctx.Param("teamname")
	var teamUpdateBody UpdateTeam
	var team model.Team

	//Check if session have access to the resource
	cookieTeam, _ := ctx.Get("team")
	hasAccessTo := strings.ToLower(cookieTeam.(model.Team).TeamName)
	if hasAccessTo != strings.ToLower(teamName) {
		th.Handler.logger.Error("User have no access to this team")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "This user have no acces to this team",
			"teamName": teamName})
		return
	}
	th.Handler.logger.Info("User have acces to the resource")

	if err := ctx.ShouldBindJSON(&teamUpdateBody); err != nil {
		th.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := repository.DB.Select("team_name", "id").Where("team_name = ?", teamName).Find(&team)
	if team.ID == 0 || row.Error != nil {
		th.Handler.logger.Error("Invalid team name")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find team for the teamname",
			"teamName": teamName,
		})
		return
	}
	team.TeamName = teamUpdateBody.TeamName
	repository.DB.Model(&team).Association("Users")
	repository.DB.Unscoped().Model(&team).Association("Users").Unscoped().Clear()
	repository.DB.Model(&team).Association("Users").Append(teamUpdateBody.TeamMembers)
	repository.DB.Model(&team).Where("ID = ?", team.ID).Updates(team)

	ctx.JSON(http.StatusOK, gin.H{"message": "Team Sucesfully updated",
		"UpdateData": teamUpdateBody})

}

//TODO add email logic to send password reset link

func (th TeamHandler) ChangePassword(ctx *gin.Context) {
	teamName := ctx.Param("teamname")
	var newCredentials PasswordChange
	var team model.Team

	//Check if session have access to the resource
	cookieTeam, _ := ctx.Get("team")
	hasAccessTo := strings.ToLower(cookieTeam.(model.Team).TeamName)
	if hasAccessTo != strings.ToLower(teamName) {
		th.Handler.logger.Error("User have no access to this team")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "This user have no acces to this team",
			"teamName": teamName})
		return
	}
	th.Handler.logger.Info("User have acces to the resource")

	if err := ctx.ShouldBindJSON(&newCredentials); err != nil {
		th.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	th.Handler.logger.Info("Body was sucesfully binded")

	row := repository.DB.Select("team_name", "id").Where("team_name = ?", teamName).Find(&team)
	th.Handler.logger.Info("Retreive following team from DB",
		zap.String("teamName", team.TeamName),
		zap.Uint("team_id", team.ID))

	if team.ID == 0 || row.Error != nil {
		th.Handler.logger.Error("Invalid team name")
		ctx.JSON(http.StatusConflict, gin.H{
			"error":    "Cannot find team for the teamname",
			"teamName": teamName,
		})
		return
	}

	hash, err := repository.HashPassword(newCredentials.NewPassword)
	if err != nil {
		th.Handler.logger.Error("Hash password error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	th.Handler.logger.Info("Password was Hashed")
	team.Password = hash
	repository.DB.Model(&team).Updates(team)

	ctx.JSON(http.StatusAccepted, gin.H{
		"message":  "Sucesfully updated password",
		"teamName": team.TeamName,
	})
}
