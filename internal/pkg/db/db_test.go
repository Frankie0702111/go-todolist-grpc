package db_test

import (
	"crypto/x509"
	"encoding/pem"
	"go-todolist-grpc/internal/config"
	mydb "go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestInitSuccess(t *testing.T) {
	// Performing this unit test may result in the exposure of AWS user information, which is only open to the local.
	// t.Run("Success_AWS", func(t *testing.T) {
	// 	certPath := mydb.GetRootCertPath()
	// 	if _, err := os.Stat(certPath); os.IsNotExist(err) {
	// 		t.Skip("Skipping AWS test: RDS certificate file not found")
	// 	}

	// 	connectionMaxLifeTimeSec := config.SourceDBConnMaxLTSec
	// 	maxConn := config.SourceMaxConn
	// 	maxIdle := config.SourceMaxIdle
	// 	log.Init(config.LogLevel, config.LogFolderPath, strconv.Itoa(os.Getpid()), config.EnableConsoleOutput, config.EnableFileOutput)

	// 	db, mock, err := sqlmock.New()
	// 	if err != nil {
	// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	// 	}
	// 	defer db.Close()

	// 	mock.ExpectPing()

	// 	// Simulate the SSL query
	// 	rows := sqlmock.NewRows([]string{"ssl_is_used"}).AddRow(true)
	// 	mock.ExpectQuery("SELECT ssl_is_used()").WillReturnRows(rows)

	// 	opt := &mydb.Option{
	// 		Host:                     config.AWSSourceHost,
	// 		Port:                     config.SourcePort,
	// 		Username:                 config.AWSSourceUser,
	// 		Password:                 config.AWSSourcePassword,
	// 		DBName:                   config.AWSSourceDataBase,
	// 		SSLMode:                  config.AWSSourceSSLMode,
	// 		ConnectionMaxLifeTimeSec: &connectionMaxLifeTimeSec,
	// 		MaxConn:                  &maxConn,
	// 		MaxIdle:                  &maxIdle,
	// 	}

	// 	err = mydb.Init(opt)
	// 	assert.NoError(t, err)

	// 	// Checking the SSL status
	// 	var sslUsed bool
	// 	err = db.QueryRow("SELECT ssl_is_used()").Scan(&sslUsed)
	// 	assert.NoError(t, err)
	// 	assert.True(t, sslUsed, "SSL should be used")

	// 	// We make sure that all expectations were met
	// 	if err := mock.ExpectationsWereMet(); err != nil {
	// 		t.Errorf("there were unfulfilled expectations: %s", err)
	// 	}
	// })

	t.Run("Success_General", func(t *testing.T) {
		connectionMaxLifeTimeSec := config.SourceDBConnMaxLTSec
		maxConn := config.SourceMaxConn
		maxIdle := config.SourceMaxIdle
		log.Init(config.LogLevel, config.LogFolderPath, strconv.Itoa(os.Getpid()), config.EnableConsoleOutput, config.EnableFileOutput)

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		// Local set the SourceHost to "127.0.0.1"
		// Docker set the SourceHost to "db"
		opt := &mydb.Option{
			Host:                     config.SourceHost,
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
			SSLMode:                  config.SourceSSLMode,
			ConnectionMaxLifeTimeSec: &connectionMaxLifeTimeSec,
			MaxConn:                  &maxConn,
			MaxIdle:                  &maxIdle,
		}

		err = mydb.Init(opt)
		assert.NoError(t, err)

		// We make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failure_InvalidHost", func(t *testing.T) {
		log.Init(config.LogLevel, config.LogFolderPath, strconv.Itoa(os.Getpid()), config.EnableConsoleOutput, config.EnableFileOutput)

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     "invalidhost",
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
			SSLMode:                  config.SourceSSLMode,
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.Contains(t, err.Error(), "dial tcp: lookup invalidhost")
	})

	t.Run("Failure_InvalidPort", func(t *testing.T) {
		log.Init(config.LogLevel, config.LogFolderPath, strconv.Itoa(os.Getpid()), config.EnableConsoleOutput, config.EnableFileOutput)

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     config.SourceHost,
			Port:                     "invalidport",
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
			SSLMode:                  config.SourceSSLMode,
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.EqualError(t, err, "dial tcp: lookup tcp/invalidport: unknown port")
	})

	t.Run("Failure_InvalidUser", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     config.SourceHost,
			Port:                     config.SourcePort,
			Username:                 "invalid",
			Password:                 "invalid",
			DBName:                   config.SourceDataBase,
			SSLMode:                  config.SourceSSLMode,
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.EqualError(t, err, "pq: password authentication failed for user \"invalid\"")
	})

	t.Run("Failure_InvalidDBName", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     config.SourceHost,
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   "invaliddbname",
			SSLMode:                  config.SourceSSLMode,
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.EqualError(t, err, "pq: database \"invaliddbname\" does not exist")
	})

	t.Run("Failure_InvalidSSLMode", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     config.SourceHost,
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
			SSLMode:                  "invalidsslmode",
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.EqualError(t, err, "pq: unsupported sslmode \"invalidsslmode\"; only \"require\" (default), \"verify-full\", \"verify-ca\", and \"disable\" supported")
	})
}

func TestGetConn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		log.Init(config.LogLevel, config.LogFolderPath, strconv.Itoa(os.Getpid()), config.EnableConsoleOutput, config.EnableFileOutput)

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     config.SourceHost,
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
			SSLMode:                  config.SourceSSLMode,
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.NoError(t, err)

		conn := mydb.GetConn()
		assert.NotNil(t, conn)

		err = conn.Ping()
		assert.NoError(t, err)
	})

	t.Run("Failure_NotInitialized", func(t *testing.T) {
		mydb.ResetConn()
		conn := mydb.GetConn()
		assert.Nil(t, conn)
	})
}

func TestVerifyCertificate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		certPath := mydb.GetRootCertPath()
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			t.Skip("Skipping test: RDS certificate file not found")
		}

		// Should be able to read the certificate file
		certData, err := os.ReadFile(certPath)
		assert.NoError(t, err)

		// Should be able to decode the PEM block
		block, _ := pem.Decode(certData)
		assert.NotNil(t, block)

		// Should be able to parse the certificate
		cert, err := x509.ParseCertificate(block.Bytes)
		assert.NoError(t, err)

		// Certificate issuer should match
		assert.Equal(t, "CN=Amazon RDS Root 2019 CA,OU=Amazon RDS,O=Amazon Web Services\\, Inc.,L=Seattle,ST=Washington,C=US", cert.Issuer.String())

		// Certificate should be valid
		now := time.Now()
		assert.True(t, now.After(cert.NotBefore) && now.Before(cert.NotAfter))
	})
}
