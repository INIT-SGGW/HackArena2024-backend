package service

import (
	"fmt"
	"net/smtp"
	"os"

	"go.uber.org/zap"
)

type EmailService struct {
	logger                      *zap.Logger
	email                       string
	password                    string
	emailHost                   string
	emailPort                   string
	websiteURL                  string
	passwordResetEmailBodyStart string
	passwordResetEmailBodyEnd   string
}

// TODO: Change declaration from file
func NewEmailService(logger *zap.Logger) *EmailService {
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
	templateBodyStart := ` 
	<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Zmiana Hasla - Hackarena</title>
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
					<h1>Zmaiana hasla</h1>
				</div>
				<div class="content">
					<p>Drogi Uczestniku,</p>
					<p>Kliknij w link aby ustawic nowe haslo dla twojego konta </p>
	`

	templateBodyEnd := `
					<p>Zaobserwuj nasze social media, aby być na bieżąco! <br><br>
						 <a href="https://www.facebook.com/profile.php?id=61559358943109&is_tour_dismissed">Facebook</a> | <a href="https://www.instagram.com/kn_init_/">Instagram</a> | <a href="https://www.linkedin.com/company/102567955">LinkedIn</a>
						</p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Hackarena. Wszelkie prawa zastrzeżone.</p>
					<p>W razie pytań, prosimy o kontakt: <a href="mailto:kontakt@hackarena.pl">kontkat@hackarena.pl</a>.</p>
				</div>
			</div>
		</body>
		</html>
		`
	return &EmailService{
		logger:                      logger,
		email:                       email,
		password:                    password,
		emailHost:                   emailHost,
		emailPort:                   emailPort,
		websiteURL:                  "https://test.hackarena.pl",
		passwordResetEmailBodyStart: templateBodyStart,
		passwordResetEmailBodyEnd:   templateBodyEnd,
	}
}

func (es EmailService) SendResetPasswordEmail(email, oneTimePassword string) error {
	defer es.logger.Sync()

	es.logger.Info("Start authentication")

	auth := smtp.PlainAuth("", es.email, es.password, es.emailHost)
	es.logger.Info("Authenticated")

	es.logger.Info("Start creating the email")
	link := fmt.Sprintf("%s/reset?email=%s&token=%s", es.websiteURL, email, oneTimePassword)
	body := es.passwordResetEmailBodyStart + link + es.passwordResetEmailBodyEnd
	to := []string{email}
	message := fmt.Sprintf("From: %s\r\n", es.email)
	message += fmt.Sprintf("To: %s\r\n", email)
	message += "Subject: HackArena2.0 Password restart\r\n"
	message += "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message += fmt.Sprintf("\r\n%s\r\n", body)

	es.logger.Info("Email created")
	es.logger.Info("Sending email",
		zap.String("recipient", email))
	err := smtp.SendMail(es.emailHost+":"+es.emailPort, auth, es.email, to, []byte(message))
	if err != nil {
		es.logger.Error("Error sending the email",
			zap.Error(err))
		return err
	}
	es.logger.Info("Email was sucesfully send",
		zap.String("recipient", email))

	return nil
}
