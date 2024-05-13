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

// func InitDB() *sql.DB {
// 	psqlInfo := getConnectionString()
// 	db, err := sql.Open("postgres", psqlInfo)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = db.Ping()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Successfully connected!")
// 	return db

// }
func ConnectDataBase() *gorm.DB {
	var err error
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	var DB *gorm.DB
	psqlinfo := getConnectionString()
	DB, err = gorm.Open(postgres.Open(psqlinfo))

	if err != nil {
		logger.Error("Cannot connect to database")
	} else {
		logger.Info("Sucesfully connected to Database")
	}
	DB.AutoMigrate(&model.Team{}, &model.User{})
	return DB
}
