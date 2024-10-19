package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"INIT-SGGW/hackarena-backend/service"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailHandler struct {
	Handler          Handler
	TemplateBasePath string
	service          *service.EmailService
}

func NewEmailHandler(logger *zap.Logger) *EmailHandler {
	pathToTemplates, exist := os.LookupEnv("HA_EMAIL_TEMP_PATH")
	if !exist {
		logger.Error("The HA_EMAIL_TEMP_PATH environmental variable is missing, it should contain base path for templates directory")
		os.Exit(2)
	}
	return &EmailHandler{
		Handler:          *NewHandler(logger),
		TemplateBasePath: pathToTemplates,
		service:          service.NewEmailService(logger),
	}
}

func (eh EmailHandler) SendEmail(ctx *gin.Context) {
	defer eh.Handler.Logger.Sync()

	var emailRequest model.SendEmailRequest

	if err := ctx.ShouldBindJSON(&emailRequest); err != nil {
		eh.Handler.Logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	eh.Handler.Logger.Info("The input body is valid")
	emails, err := eh.GetEmailsByFilter(emailRequest.Recipients.MailingGroups)
	if err != nil {
		eh.Handler.Logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	eh.Handler.Logger.Info("Emails",
		zap.Strings("emails", emails))
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"message": "The endpoint is not implemented",
	})

	// Find user by filter

}

func (eh EmailHandler) GetEmailsByFilter(filter string) ([]string, error) {
	defer eh.Handler.Logger.Sync()
	var emails []string

	query, err := eh.GetFilterQuery(filter)
	if err != nil {
		eh.Handler.Logger.Error("GetFilterQuery error",
			zap.Error(err))
		wraped := fmt.Errorf("[GetEmailsByFilter] error in geting filter %w", err)
		return []string{}, wraped
	}
	eh.Handler.Logger.Info("Query sucesfully retreive",
		zap.String("query", query))

	return emails, nil
}

func (eh EmailHandler) GetFilterQuery(filter string) (string, error) {
	defer eh.Handler.Logger.Sync()
	mailingFilter := &model.MailingGroupFilter{}
	query := ""
	err := repository.DB.Model(&model.MailingGroupFilter{}).Where("filter_name = ?", filter).First(&mailingFilter).Error
	if err != nil {
		eh.Handler.Logger.Error("GetFilterQuery error",
			zap.Error(err))
		wraped := fmt.Errorf("[GetFilterQuery] error in geting filter %w", err)
		return query, wraped
	}
	query = mailingFilter.Query

	eh.Handler.Logger.Info("Sucesfully get query",
		zap.String("usedFilter", filter),
		zap.String("query", query))

	return query, nil

}
