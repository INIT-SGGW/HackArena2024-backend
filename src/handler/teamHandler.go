package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"encoding/json"
	"net/http"

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
	defer th.Handler.Logger.Sync()

	teamName := ctx.MustGet("team_name").(string)
	teamIsVerified := ctx.MustGet("team_is_veifird").(bool)
	hasAccessToTeamWithId := ctx.MustGet("team_id").(uint)

	th.Handler.Logger.Info("Start retreiving data from database",
		zap.String("team", teamName),
		zap.Uint("team_id", hasAccessToTeamWithId))

	members := []model.Member{}
	result := repository.DB.Model(&model.Member{}).Where("team_id = ?", hasAccessToTeamWithId).Find(&members)
	if result.Error != nil {
		th.Handler.Logger.Error("Error while retreiving users for team",
			zap.String("teamName", teamName))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while retreiving users"})
		return
	}
	th.Handler.Logger.Info("Sucessfully get members data from database",
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
		TeamName:    teamName,
		IsVerified:  teamIsVerified,
		TeamMembers: membersToResponse,
	})
	if err != nil {
		th.Handler.Logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshall failed",
		})
		return
	}

	ctx.Data(http.StatusAccepted, "application/json", jsonBody)
}

func (th TeamHandler) ConfirmTeam(ctx *gin.Context) {
	defer th.Handler.Logger.Sync()

	hasAccessToTeamWithId := ctx.MustGet("team_id").(uint)

	th.Handler.Logger.Info("Update confirmation value")

	err := repository.DB.Model(&model.Team{}).Where("id = ?", hasAccessToTeamWithId).Update("is_confirmed", true).Error
	if err != nil {
		th.Handler.Logger.Error("Error inserting to database")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database insert failed",
		})
		return
	}

	ctx.AbortWithStatus(200)
}

func (th TeamHandler) GetAllTeamsAsAdmin(ctx *gin.Context) {
	defer th.Handler.Logger.Sync()

	teams := []model.Team{}
	err := repository.DB.Model(&model.Team{}).Preload("Members").Find(&teams).Error
	if err != nil {
		th.Handler.Logger.Error("Error when retreiving teams from database",
			zap.Error(err))
		ctx.AbortWithStatus(500)
		return
	}
	th.Handler.Logger.Info("Sucesfully retreive teams")

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
		th.Handler.Logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshall failed",
		})
		return
	}

	ctx.Data(http.StatusAccepted, "application/json", jsonBody)
}

func (th TeamHandler) GetAllUsersAsAdmin(ctx *gin.Context) {
	defer th.Handler.Logger.Sync()

	teams := []model.Team{}
	err := repository.DB.Model(&model.Team{}).Preload("Members").Find(&teams).Error
	if err != nil {
		th.Handler.Logger.Error("Error when retreiving teams from database",
			zap.Error(err))
		ctx.AbortWithStatus(500)
		return
	}
	th.Handler.Logger.Info("Sucesfully retreive teams")

	memberResponses := []model.UserResponse{}

	for _, team := range teams {
		for _, member := range team.Members {
			var firstName string
			var lastName string
			if member.FirstName == nil {
				firstName = "not verified"
			} else {
				firstName = *member.FirstName
			}
			if member.LastName == nil {
				lastName = "not verified"
			} else {
				lastName = *member.LastName
			}

			newMemberResponseEntry := model.UserResponse{
				TeamName:   team.TeamName,
				Email:      member.Email,
				FirstName:  firstName,
				LastName:   lastName,
				IsVerified: member.IsVerified,
			}
			memberResponses = append(memberResponses, newMemberResponseEntry)
		}
	}
	response := model.GetAllUsersResponse{
		Users: memberResponses,
	}

	jsonBody, err := json.Marshal(response)
	if err != nil {
		th.Handler.Logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshall failed",
		})
		return
	}

	ctx.Data(http.StatusAccepted, "application/json", jsonBody)

}
