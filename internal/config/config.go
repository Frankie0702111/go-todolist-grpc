package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var configIns *Config

type Config struct {
	HttpServerPort string `mapstructure:"HTTP_SERVER_PORT"`
	GprcServerPort string `mapstructure:"GRPC_SERVER_PORT"`

	DBHost                     string `mapstructure:"DB_HOST"`
	DBPort                     string `mapstructure:"DB_PORT"`
	DBUser                     string `mapstructure:"DB_USER"`
	DBPassword                 string `mapstructure:"DB_PASS"`
	DBName                     string `mapstructure:"DB_NAME"`
	DBConnectionMaxLifeTimeSec *int   `mapstructure:"DB_CONN_MAX_LT_SEC"`
	DBMaxConnection            *int   `mapstructure:"DB_MAX_CONN"`
	DBMaxIdle                  *int   `mapstructure:"DB_MAX_IDLE"`

	BcryptCost   int    `mapstructure:"BCRYPT_COST"`
	JwtSecretKey string `mapstructure:"JWT_SECRET_KEY"`
	JwtTtl       int    `mapstructure:"JWT_TTL"`

	LogLevel            int    `mapstructure:"LOG_LEVEL"`
	LogFolderPath       string `mapstructure:"LOG_FOLDER_PATH"`
	EnableConsoleOutput bool   `mapstructure:"ENABLE_CONSOLE_OUTPUT"`
	EnableFileOutput    bool   `mapstructure:"ENABLE_FILE_OUTPUT"`
}

func Load() error {
	godotenv.Load("app.env")
	appFolderPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	viper.AddConfigPath(appFolderPath)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	configIns = &Config{}
	if err := viper.Unmarshal(&configIns); err != nil {
		return err
	}

	return nil
}

func Get() *Config {
	return configIns
}

func ResetConfig() {
	configIns = nil
}
