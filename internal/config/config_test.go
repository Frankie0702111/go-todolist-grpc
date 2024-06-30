package config_test

import (
	"bytes"
	"go-todolist-grpc/internal/config"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAndGet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup app.env variables
		var mockConfigContent bytes.Buffer
		mockConfigContent.WriteString("HTTP_SERVER_PORT=" + config.HttpPort + "\n")
		mockConfigContent.WriteString("GRPC_SERVER_PORT=" + config.GrpcPort + "\n")
		mockConfigContent.WriteString("DB_HOST=" + config.SourceHost + "\n")
		mockConfigContent.WriteString("DB_PORT=" + config.SourcePort + "\n")
		mockConfigContent.WriteString("DB_USER=" + config.SourceUser + "\n")
		mockConfigContent.WriteString("DB_PASS=" + config.SourcePassword + "\n")
		mockConfigContent.WriteString("DB_NAME=" + config.SourceDataBase + "\n")
		mockConfigContent.WriteString("DB_CONN_MAX_LT_SEC=" + strconv.Itoa(config.SourceDBConnMaxLTSec) + "\n")
		mockConfigContent.WriteString("DB_MAX_CONN=" + strconv.Itoa(config.SourceMaxConn) + "\n")
		mockConfigContent.WriteString("DB_MAX_IDLE=" + strconv.Itoa(config.SourceMaxIdle) + "\n")
		mockConfigContent.WriteString("BCRYPT_COST=" + strconv.Itoa(config.BcryptCost) + "\n")
		mockConfigContent.WriteString("JWT_SECRET_KEY=" + config.JwtSecretKey + "\n")
		mockConfigContent.WriteString("JWT_TTL=" + strconv.Itoa(config.JwtTtl) + "\n")
		mockConfigContent.WriteString("LOG_LEVEL=" + strconv.Itoa(config.LogLevel) + "\n")
		mockConfigContent.WriteString("LOG_FOLDER_PATH=" + config.LogFolderPath + "\n")
		mockConfigContent.WriteString("ENABLE_CONSOLE_OUTPUT=" + strconv.FormatBool(config.EnableConsoleOutput) + "\n")
		mockConfigContent.WriteString("ENABLE_FILE_OUTPUT=" + strconv.FormatBool(config.EnableFileOutput) + "\n")

		// Create app.env file
		appFolderPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		mockConfigFile := filepath.Join(appFolderPath, "app.env")
		err := os.WriteFile(mockConfigFile, mockConfigContent.Bytes(), 0644)
		assert.NoError(t, err)
		defer os.Remove(mockConfigFile)

		// Init config
		loadErr := config.Load()
		assert.NoError(t, loadErr)

		// Check config
		cnf := config.Get()
		assert.Equal(t, config.HttpPort, cnf.HttpServerPort)
		assert.Equal(t, config.GrpcPort, cnf.GprcServerPort)
		assert.Equal(t, config.SourceHost, cnf.DBHost)
		assert.Equal(t, config.SourcePort, cnf.DBPort)
		assert.Equal(t, config.SourceUser, cnf.DBUser)
		assert.Equal(t, config.SourcePassword, cnf.DBPassword)
		assert.Equal(t, config.SourceDataBase, cnf.DBName)
		assert.Equal(t, config.SourceDBConnMaxLTSec, *cnf.DBConnectionMaxLifeTimeSec)
		assert.Equal(t, config.SourceMaxConn, *cnf.DBMaxConnection)
		assert.Equal(t, config.SourceMaxIdle, *cnf.DBMaxIdle)
		assert.Equal(t, config.BcryptCost, cnf.BcryptCost)
		assert.Equal(t, config.JwtSecretKey, cnf.JwtSecretKey)
		assert.Equal(t, config.JwtTtl, cnf.JwtTtl)
		assert.Equal(t, config.LogLevel, cnf.LogLevel)
		assert.Equal(t, config.LogFolderPath, cnf.LogFolderPath)
		assert.True(t, cnf.EnableConsoleOutput)
		assert.True(t, cnf.EnableFileOutput)
	})

	t.Run("Failure_InvalidEnvFile", func(t *testing.T) {
		config.ResetConfig()

		appFolderPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		os.Chdir(appFolderPath)
		os.Setenv("BCRYPT_COST", "invalid")

		err := config.Load()
		assert.Error(t, err)

		cfg := config.Get()
		assert.Nil(t, cfg)
	})
}
