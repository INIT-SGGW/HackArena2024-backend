package repository

import (
	"INIT-SGGW/hackarena-backend/model"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

const (
	driver = "postgresql"
	host   = "localhost"
	port   = 800
	dbname = "hackarena"
)

func getConnectionString() string {
	user := os.Getenv("HACKDB_USER")
	password := os.Getenv("HACKDB_PWD")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return psqlInfo
}

func ConnectDataBase() {
	var err error
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	psqlinfo := getConnectionString()
	DB, err = gorm.Open(postgres.Open(psqlinfo))

	if err != nil {
		logger.Error("Cannot connect to database")
	} else {
		logger.Info("Sucesfully connected to Database")
	}
}
func SyncDB() {
	DB.AutoMigrate(&model.Team{}, &model.User{}, &model.File{})
}
