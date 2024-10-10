package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamHandler struct {
	Handler Handler
}

func NewTeamHandler(logger *zap.Logger) *TeamHandler {
	return &TeamHandler{
		Handler: *NewHandler(logger),
	}
}

func (th TeamHandler) RetreiveTeam(ctx *gin.Context) {
	defer th.Handler.logger.Sync()

	teamName := ctx.Param("teamname")
	if teamName == "" {
		th.Handler.logger.Error("Missing teamName parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Mising teamName parameter"})
		return
	}

	th.Handler.logger.Info("The input is valid",
		zap.String("teamName", teamName))

	th.Handler.logger.Info("Checking if user have access to requested team")

	cookieUser, _ := ctx.Get("user")
	hasAccessToTeamWithId := cookieUser.(model.Member).TeamID
	team := &model.Team{}
	result := repository.DB.Select("team_name,id,is_verified").Where("id = ?", hasAccessToTeamWithId).First(&team)
	if result.Error != nil {
		th.Handler.logger.Error("The team for provided user do not exist or another retreive error occure")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error":  "The team for provided user",
			"teamId": hasAccessToTeamWithId})
		return
	}

	if !strings.EqualFold(team.TeamName, teamName) {
		th.Handler.logger.Error("User have no acces to this team",
			zap.String("requestedTeam", teamName),
			zap.String("teamInCookie", team.TeamName))
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "User have no acces to this team"})
		return
	}
	th.Handler.logger.Info("User have access to requested team")

	th.Handler.logger.Info("Start retreiving data from database",
		zap.String("team", team.TeamName),
		zap.Uint("team_id", hasAccessToTeamWithId))

	members := []model.Member{}
	result = repository.DB.Model(&model.Member{}).Where("team_id = ?", hasAccessToTeamWithId).Find(&members)
	if result.Error != nil {
		th.Handler.logger.Error("Error while retreiving users for team",
			zap.String("teamName", team.TeamName))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while retreiving users"})
		return
	}
	th.Handler.logger.Info("Sucessfully get members data from database",
		zap.Int("recordCollected", len(members)))

	membersToResponse := []model.GetTeamMemberResponse{}
	for _, member := range members {
		toResponse := model.GetTeamMemberResponse{
			Email:     member.Email,
			FirstName: member.FirstName,
			LastName:  member.LastName,
			Verified:  member.IsVerified,
		}
		membersToResponse = append(membersToResponse, toResponse)
	}

	jsonBody, err := json.Marshal(model.GetTeamResponse{
		TeamName:    team.TeamName,
		IsVerified:  team.IsVerified,
		TeamMembers: membersToResponse,
	})
	if err != nil {
		th.Handler.logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshall failed",
		})
		return
	}

	ctx.Data(http.StatusAccepted, "application/json", jsonBody)
}

func (th TeamHandler) ConfirmTeam(ctx *gin.Context) {
	defer th.Handler.logger.Sync()

	teamName := ctx.Param("teamname")
	if teamName == "" {
		th.Handler.logger.Error("Missing teamName parameter")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Mising teamName parameter"})
		return
	}

	th.Handler.logger.Info("The input is valid",
		zap.String("teamName", teamName))

	th.Handler.logger.Info("Checking if user have access to requested team")

	cookieUser, _ := ctx.Get("user")
	hasAccessToTeamWithId := cookieUser.(model.Member).TeamID
	team := &model.Team{}
	result := repository.DB.Select("team_name,id,is_verified").Where("id = ?", hasAccessToTeamWithId).First(&team)
	if result.Error != nil {
		th.Handler.logger.Error("The team for provided user do not exist or another retreive error occure")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "The team for provided user do not exist or user have no acces to that team"})
		return
	}
	th.Handler.logger.Info("User have acces to the team")
	th.Handler.logger.Info("Update confirmation value")

	err := repository.DB.Model(&model.Team{}).Where("id = ?", team.ID).Update("is_confirmed", true).Error
	if err != nil {
		th.Handler.logger.Error("Error inserting to database")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database insert failed",
		})
		return
	}

	ctx.AbortWithStatus(200)
}

func (th TeamHandler) GetAllTeamsAsAdmin(ctx *gin.Context) {
	teams := []model.Team{}
	err := repository.DB.Model(&model.Team{}).Preload("Members").Find(&teams).Error
	if err != nil {
		th.Handler.logger.Error("Error when retreiving teams from database",
			zap.Error(err))
		ctx.AbortWithStatus(500)
		return
	}
	th.Handler.logger.Info("Sucesfully retreive teams")

	responseTeams := []model.TeamResponse{}

	for _, team := range teams {
		newTeam := model.TeamResponse{
			TeamName:         team.TeamName,
			IsVerified:       team.IsVerified,
			ApproveSatatus:   team.ApproveStatus,
			TeamMembersCount: len(team.Members),
		}
		responseTeams = append(responseTeams, newTeam)
	}
	response := model.GetAllTeamsResponse{Teams: responseTeams}

	jsonBody, err := json.Marshal(response)
	if err != nil {
		th.Handler.logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshall failed",
		})
		return
	}

	ctx.Data(http.StatusAccepted, "application/json", jsonBody)
}
