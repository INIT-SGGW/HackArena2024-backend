package repository

import (
	"INIT-SGGW/hackarena-backend/model"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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

	configPath, exist := os.LookupEnv("HACKDB_CONFIG_PATH")
	if !exist {
		fmt.Println("The HACKDB_CONFIG_PATH environmental variable is missing")
		os.Exit(2)
	}

	viper.AddConfigPath(configPath) //Base config path for application
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

func ConnectDataBase(logger *zap.Logger) {
	defer logger.Sync()
	var err error

	psqlinfo := getConnectionString()
	DB, err = gorm.Open(postgres.Open(psqlinfo))

	if err != nil {
		logger.Error("Cannot connect to database")
	} else {
		logger.Info("Sucesfully connected to Database")
	}
}
func SyncDB() {
	DB.AutoMigrate(&model.Team{}, &model.Member{}, &model.SolutionFile{}, &model.MatchFile{}, &model.HackArenaAdmin{}, &model.MailingGroupFilter{}, &model.EmailTemplates{})
}

func CreateLogger() *zap.Logger {
	stdout := zapcore.AddSync(os.Stdout)

	logPath, exist := os.LookupEnv("HA_LOG_PATH")
	if !exist {
		fmt.Println("The HA_LOG_PATH environmental variable is missing")
		os.Exit(2)
	}

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // megabytes
		MaxBackups: 1,
		MaxAge:     3, // days
	})

	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)

	return zap.New(core)
}
