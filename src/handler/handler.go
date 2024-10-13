package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	Logger *zap.Logger
}

func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{
		Logger: logger,
	}
}

func (h Handler) ValidateTeamScope() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		defer h.Logger.Sync()
		teamName := ctx.Param("teamname")
		if teamName == "" {
			h.Logger.Error("Missing teamName parameter")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Mising teamName parameter"})
			ctx.Abort()
			return
		}

		h.Logger.Info("The input is valid",
			zap.String("teamName", teamName))

		h.Logger.Info("Checking if user have access to requested team")

		cookieUser, _ := ctx.Get("user")
		hasAccessToTeamWithId := cookieUser.(model.Member).TeamID
		team := &model.Team{}
		result := repository.DB.Select("team_name,id,is_verified").Where("id = ?", hasAccessToTeamWithId).First(&team)
		if result.Error != nil {
			h.Logger.Error("The team for provided user do not exist or another retreive error occure")
			ctx.JSON(http.StatusForbidden, gin.H{
				"error":  "The team for provided user",
				"teamId": hasAccessToTeamWithId})
			ctx.Abort()
			return
		}

		if !strings.EqualFold(team.TeamName, teamName) {
			h.Logger.Error("User have no acces to this team",
				zap.String("requestedTeam", teamName),
				zap.String("teamInCookie", team.TeamName))
			ctx.AbortWithStatus(401)
			return
		}
		h.Logger.Info("User have access to requested team")
		ctx.Set("team_id", team.ID)
		ctx.Set("team_is_veifird", team.IsVerified)
		ctx.Set("team_name", team.TeamName)
		ctx.Next()
	}

}
