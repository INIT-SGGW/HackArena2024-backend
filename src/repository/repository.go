package repository

import (
	"INIT-SGGW/hackarena-backend/model"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getConnectionString() string {
	viper.AddConfigPath("/etc/hackarena-backend/config") //Base config path for application
	viper.SetConfigName("dbconf")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error in reading config", viper.AllKeys())
		os.Exit(1)
	}

	user := os.Getenv("HACKDB_USER")
	password := os.Getenv("HACKDB_PWD")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		viper.GetString("DB_HOST"), viper.GetInt("DB_PORT"), user, password, viper.GetString("DB_NAME"))

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
