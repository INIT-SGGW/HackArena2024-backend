package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"INIT-SGGW/hackarena-backend/service"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserAccountHandler struct {
	Handler Handler
	service *service.EmailService
}

func NewUserAccountHandler(logger *zap.Logger) *UserAccountHandler {
	return &UserAccountHandler{
		Handler: *NewHandler(logger),
		service: service.NewEmailService(logger),
	}
}

func (uh UserAccountHandler) LoginUser(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	var input model.LoginRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		uh.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//retreive password from database
	var dbObject model.Member
	row := repository.DB.Table("members").Where("email = ?", input.Email).
		Select([]string{"id", "email", "password"}).Find(&dbObject)

	if row.Error != nil {
		uh.Handler.logger.Info("Invalid email")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or email",
		})
		return
	}
	//Validate provided password
	isValid := repository.CheckPasswordHash(input.Password, dbObject.Password)
	if !isValid {
		uh.Handler.logger.Error("Invalid password")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password or email",
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

	uh.Handler.logger.Info("JWT token created")
	//Add cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("HACK-Arena-Authorization", tokenString, 3600*24, "", "", false, true)

	var team model.Team
	result := repository.DB.Model(&model.Team{}).Select("teams.team_name").Joins("inner join members on members.team_id = teams.id").Where("members.email = ? ", dbObject.Email).First(&team)

	if result.Error != nil {
		uh.Handler.logger.Error("Error marshaling response",
			zap.String("email", dbObject.Email))

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to retreive the team",
		})
		return
	}
	jsonBody, err := json.Marshal(model.LoginResponse{
		TeamName: team.TeamName,
		Email:    dbObject.Email,
	})
	if err != nil {
		uh.Handler.logger.Error("Error marshaling response")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Response marshal failed",
		})
		return
	}

	uh.Handler.logger.Info("Sucesfully log in")
	ctx.Data(http.StatusAccepted, "application/json", jsonBody)

}

//Chanhe to member not team pasword

func (uh UserAccountHandler) ChangePassword(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	var changePasswordRequest model.ChangePasswordRequest

	if err := ctx.ShouldBindJSON(&changePasswordRequest); err != nil {
		uh.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uh.Handler.logger.Info("The JSON is valid")

	uh.Handler.logger.Info("Checking if user have access to requested team")

	cookieUser, _ := ctx.Get("user")
	userEmail := cookieUser.(model.Member).Email
	member := &model.Member{}
	result := repository.DB.Select("email,password").Where("email = ?", userEmail).First(&member)
	if result.Error != nil {
		uh.Handler.logger.Error("The member for login user do not exist or another retreive error occure")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":     "The team for provided user",
			"userEmail": userEmail})
		return
	}
	uh.Handler.logger.Info("User have acces to the resource")

	isValid := repository.CheckPasswordHash(changePasswordRequest.OldPassword, member.Password)
	if !isValid {
		uh.Handler.logger.Error("The password do not match with the one in database")
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Password Invalid"})
		return
	}
	uh.Handler.logger.Info("Password is correct")

	hash, err := repository.HashPassword(changePasswordRequest.NewPassword)
	if err != nil {
		uh.Handler.logger.Error("Error when hashing password",
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error when hashing new password"})
		return
	}
	uh.Handler.logger.Info("Password was hashed")
	row := repository.DB.Model(&model.Member{}).Where("email = ?", userEmail).Update("password", hash)
	if row.Error != nil {
		uh.Handler.logger.Error("Error updating password to database",
			zap.Error(row.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database update error",
		})
		return
	}
	uh.Handler.logger.Info("Password was sucesfully updated")

	ctx.AbortWithStatus(201)
}

func (uh UserAccountHandler) RestartForgotPassword(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	var forgotPasswordRequest model.ForgotPasswordRequest

	if err := ctx.ShouldBindJSON(&forgotPasswordRequest); err != nil {
		uh.Handler.logger.Error("Input body error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uh.Handler.logger.Info("The JSON is valid")

	var dbObject model.Member
	row := repository.DB.Table("members").Where("email = ?", forgotPasswordRequest.Email).
		Select([]string{"id", "email", "password", "is_verified"}).Find(&dbObject)
	if row.Error != nil {
		uh.Handler.logger.Error("Invalid email")
		ctx.JSON(http.StatusNotAcceptable, gin.H{
			"error": "There is no such email in database",
		})
		return
	}
	if !dbObject.IsVerified {
		uh.Handler.logger.Error("The user is not verified")
		ctx.JSON(http.StatusNotAcceptable, gin.H{
			"error": "Unverified user",
		})
		return
	}
	uh.Handler.logger.Info("The user is verified")

	uh.Handler.logger.Info("Generate one time password")
	uniqueVerificationToken := uuid.NewString()
	oneTimePassword, err := repository.HashPassword(uniqueVerificationToken)
	if err != nil {
		uh.Handler.logger.Error("Error in password hashing",
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unverified user",
		})
		return
	}

	uh.Handler.logger.Info("Start transaction")
	tx := repository.DB.Begin()
	tx.Model(&model.Member{}).Where("email = ? ", dbObject.Email).Update("password", oneTimePassword)
	err = uh.service.SendResetPasswordEmail(dbObject.Email, uniqueVerificationToken)
	if err != nil {
		tx.Rollback()
		uh.Handler.logger.Error("Error in sending email",
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Email send error",
		})
		return
	}
	tx.Commit()
	uh.Handler.logger.Info("Send the new password email")
	uh.Handler.logger.Info("Commit the transaction one time password has been set")

	ctx.AbortWithStatus(201)
}

func (uh UserAccountHandler) ResetPassword(ctx *gin.Context) {
	defer uh.Handler.logger.Sync()

	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Dummy endpoint ResetPassword",
	})
}
