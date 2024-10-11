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

type DBConfig struct {
	Driver   string `mapstucture:"DB_DRIVER"`
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	Name     string `mapstructure:"DB_NAME"`
	FilePath string `mapstructure:"DB_FILE_STORAGE"`
}

var DB *gorm.DB
var Config *DBConfig

func InitializeConfig() {
	viper.AddConfigPath("/etc/hackarena-backend/config") //Base config path for application
	viper.SetConfigName("dbconf")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error in reading config", viper.AllKeys())
		os.Exit(1)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		fmt.Println("Error in marshaling config", viper.AllKeys())
		os.Exit(1)
	}
}

func getConnectionString() string {
	user := os.Getenv("HACKDB_USER")
	password := os.Getenv("HACKDB_PWD")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		Config.Host, Config.Port, user, password, Config.Name)

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
	DB.AutoMigrate(&model.Team{}, &model.Member{}, &model.HackArenaAdmin{})
}
