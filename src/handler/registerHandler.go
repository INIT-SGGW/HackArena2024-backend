package handler

import (
	"INIT-SGGW/hackarena-backend/model"
	"INIT-SGGW/hackarena-backend/repository"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RegisterHandler struct {
	Handler    Handler
	email      string
	password   string
	emailHost  string
	emailPort  string
	websiteUrl string
}

func NewRegisterHandler(logger *zap.Logger) *RegisterHandler {
	defer logger.Sync()

	email, exist := os.LookupEnv("HA_EMAIL_USER")
	if !exist {
		logger.Error("The HA_EMAIL_USER environmental variable is missing")
		os.Exit(2)
	}

	password, exist := os.LookupEnv("HA_EMAIL_PWD")
	if !exist {
		logger.Error("The HA_EMAIL_PWD environmental variable is missing")
		os.Exit(2)
	}
	emailHost, exist := os.LookupEnv("HA_EMAIL_HOST")
	if !exist {
		logger.Error("The HA_EMAIL_HOST environmental variable is missing")
		os.Exit(2)
	}
	emailPort, exist := os.LookupEnv("HA_EMAIL_PORT")
	if !exist {
		logger.Error("The HA_EMAIL_PORT environmental variable is missing")
		os.Exit(2)
	}
	websiteURL, exist := os.LookupEnv("HA_WEB_URL")
	if !exist {
		logger.Error("The HA_WEB_URL environmental variable is missing")
		os.Exit(2)
	}

	return &RegisterHandler{
		Handler:    *NewHandler(logger),
		email:      email,
		password:   password,
		emailHost:  emailHost,
		emailPort:  emailPort,
		websiteUrl: websiteURL,
	}
}

func (rh RegisterHandler) RegisterTeam(ctx *gin.Context) {
	defer rh.Handler.logger.Sync()

	var input model.RegisterTeamRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		rh.Handler.logger.Error("Register team error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rh.Handler.logger.Info("JSON input is valid")

	TeamMembers := make([]model.Member, 0)
	for _, email := range input.TeamMembersEmails {
		newMember := model.Member{Email: email, Password: ""}
		TeamMembers = append(TeamMembers, newMember)
	}
	rh.Handler.logger.Info("Team members are created")

	uniqueVerificationToken := uuid.NewString()
	team := &model.Team{TeamName: input.TeamName, Members: TeamMembers, VerificationToken: uniqueVerificationToken, IsVerified: false}

	rh.Handler.logger.Info("Start registration transaction")
	tx := repository.DB.Begin()
	result := tx.Create(&team)
	if result.Error != nil {
		tx.Rollback()
		rh.Handler.logger.Error("Cannot craete new Team")
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot create new team, duplicate"})
		return
	}
	rh.Handler.logger.Info("Team is sucesfully inserted to database")

	// send verification email
	err := rh.SendVerificationEmail(team)
	if err != nil {
		tx.Rollback()
		rh.Handler.logger.Error("Error when sending the emails the insert was rollbacked",
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending emails"})
		return
	}
	tx.Commit()

	rh.Handler.logger.Info("Sucesfully created team")
	ctx.AbortWithStatus(201)
}

func (rh RegisterHandler) RegisterMember(ctx *gin.Context) {
	defer rh.Handler.logger.Sync()

	var input model.RegisterTeamMemberRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		rh.Handler.logger.Error("Register team member error body do not match request pattern")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rh.Handler.logger.Info("JSON input is valid")

	memberTeam := &model.Team{}
	result := repository.DB.Model(&memberTeam).Select("teams.verification_token,teams.team_name,teams.id").Joins("inner join members on members.team_id = teams.id").Where("members.email = ? ", input.Email).First(&memberTeam)
	if result.Error == gorm.ErrRecordNotFound {
		rh.Handler.logger.Error("There is no email or no team for that email in database",
			zap.String("email", input.Email),
			zap.Error(result.Error))
		ctx.JSON(http.StatusForbidden, gin.H{"error": "There is no email or no team for that email in database",
			"email": input.Email})
		return
	}
	if result.Error != nil {
		rh.Handler.logger.Error("Error when retreiving record from database",
			zap.String("email", input.Email),
			zap.Error(result.Error))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error when retreiving team data from database"})
		return
	}
	if memberTeam.VerificationToken != input.VerificationToken {
		rh.Handler.logger.Error("Verification token for this email do not match with the database one",
			zap.String("email", input.Email))
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Verification token for this email do not match with the database one",
			"email": input.Email})
		return
	}
	if !isOccupationValid(input.Occupation) {
		rh.Handler.logger.Error("Occupation not valid",
			zap.String("email", input.Email),
			zap.String("occupation", input.Occupation))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Occupation not valid"})
		return
	}
	if !isDietPreferenceValid(input.DietPreference) {
		rh.Handler.logger.Error("DietPreference not valid",
			zap.String("email", input.Email),
			zap.String("dietPreference", input.DietPreference))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "DietPreference not valid"})
		return
	}

	hash, err := repository.HashPassword(input.Password)
	if err != nil {
		rh.Handler.logger.Error("Error when hashing password",
			zap.String("email", input.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Hashing password error",
			"email": input.Email})
	}

	rh.Handler.logger.Info("Member verification token match with Team verification token and data is valid",
		zap.String("email", input.Email),
		zap.String("team", memberTeam.TeamName))

	newMember := &model.Member{
		TeamID:         memberTeam.ID,
		Email:          input.Email,
		Password:       hash,
		FirstName:      &input.FirstName,
		LastName:       &input.LastName,
		DateOfBirth:    (*datatypes.Date)(&input.DateOfBirth),
		Occupation:     &input.Occupation,
		DietPrefernces: &input.DietPreference,
		Agreement:      input.Agreement,
		School:         &input.School,
		IsVerified:     true,
	}

	result = repository.DB.Model(&model.Member{}).Where("email = ?", newMember.Email).Updates(&newMember)
	if result.Error != nil {
		rh.Handler.logger.Error("Error when updating the member",
			zap.String("email", input.Email),
			zap.String("teamName", memberTeam.TeamName),
			zap.Error(result.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error when updating the member"})
		return
	}

	ctx.AbortWithStatus(201)
}

// TODO: Change email to lazy (or not so lazy) initial from file and mayby separate method
func (rh RegisterHandler) SendVerificationEmail(team *model.Team) error {
	defer rh.Handler.logger.Sync()

	rh.Handler.logger.Info("Start authentication")

	auth := smtp.PlainAuth("", rh.email, rh.password, rh.emailHost)
	rh.Handler.logger.Info("Authenticated")

	for _, member := range team.Members {
		rh.Handler.logger.Info("Start sending email",
			zap.String("recipient", member.Email))
		baseUrl, err := url.Parse(rh.websiteUrl)
		if err != nil {
			rh.Handler.logger.Error("Invalid parsing of base URL",
				zap.String("baseURL", rh.websiteUrl))
			return err
		}
		baseUrl.Path += fmt.Sprintf("/rejestracja/%s", team.TeamName)
		params := url.Values{}
		params.Add("token", team.VerificationToken)
		params.Add("email", member.Email)
		baseUrl.RawQuery += params.Encode()

		link := baseUrl.String()
		to := []string{member.Email}
		message := fmt.Sprintf("From: Hackarena <%s>\r\n", rh.email)
		message += fmt.Sprintf("To: %s\r\n", member.Email)
		message += "Subject: Rejestracja\r\n"
		message += "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		body := `<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Dziękujemy za Rejestrację - Hackarena</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background-color: #ffffff;
					color: #ffffff;
					margin: 0;
					padding: 0;
				}
				.container {
					width: 100%;
					max-width: 600px;
					margin: 0 auto;
					background-color: #000314;
					padding: 20px;
					border-radius: 10px;
					box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
				}
				.header {
					text-align: center;
					padding: 10px 0;
					background-color: #000314;
					color: #FFD200;
					border-radius: 10px;
					border-width: 2px;
					border-color: #FFD200;
					border-style: solid;
				}
				.header img {
					max-width: 300px;
					margin-bottom: 10px;
				}
				.content {
					margin-top: 20px;
					text-align: center;
				}
				a {
					color: #FFD200;
					text-decoration: none;
				}
				.footer {
					text-align: center;
					padding: 10px 0;
					font-size: 12px;
					color: #757c9f;
				}
				.bold {
					font-weight: bold;
				}
				.content > * {
					color: #ffffff;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<img src="https://i.imgur.com/PoKUDEF.png" alt="Hackarena Logo">
					<h1>Dziękujemy za Rejestrację!</h1>
				</div>
				<div class="content">
					<p>Drogi Uczestniku,</p>
					<p>Dziękujemy za zarejestrowanie się na hackathon Hackarena 2.0! Cieszymy się, że chcesz dołączyć do nas w tym ekscytującym wydarzeniu.</p>
					<p>Do zakończenia procesu rejestracji, kliknij w link i wypełnij wszystkie niezbędne dane:</p>` +
			fmt.Sprintf("<a href=%s>%s</a>", link, link) +
			`<p class="bold">Informację o tym, czy twoja drużyna zakwalifikowała się na wydarzenie prześlemy do 20.10, dlatego sprawdzaj swoją skrzynkę mailową!
					</p>
					<p>Czekając zaobserwuj nasze social media, aby być na bieżąco! <br><br>
						 <a href="https://www.facebook.com/profile.php?id=61559358943109&is_tour_dismissed">Facebook</a> | <a href="https://www.instagram.com/kn_init_/">Instagram</a> | <a href="https://www.linkedin.com/company/102567955">LinkedIn</a>
						</p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Hackarena. Wszelkie prawa zastrzeżone.</p>
					<p>W razie pytań, prosimy o kontakt: <a href="mailto:kontakt@hackarena.pl">kontakt@hackarena.pl</a>.</p>
				</div>
			</div>
		</body>
		</html>`
		message += fmt.Sprintf("\r\n%s\r\n", body)

		rh.Handler.logger.Info("Send email",
			zap.Strings("recipient", to))
		err = smtp.SendMail(rh.emailHost+":"+rh.emailPort, auth, rh.email, to, []byte(message))
		if err != nil {
			rh.Handler.logger.Error("Error sending the email",
				zap.Error(err))
			return err
		}
		rh.Handler.logger.Info("Email was sucesfully send",
			zap.String("recipient", member.Email))
	}

	return nil
}

func isOccupationValid(occupation string) bool {
	return (occupation == "student" || occupation == "undergraduate" || occupation == "postgraduate" || occupation == "other")
}

func isDietPreferenceValid(dietPreference string) bool {
	return (dietPreference == "vegetarian" || dietPreference == "vegan" || dietPreference == "none")
}
