package db_test

import (
	"fmt"
	"go-todolist-grpc/internal/config"
	mydb "go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/log"
	"os"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestInitSuccess(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
			Host:                     "db",
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
			ConnectionMaxLifeTimeSec: &connectionMaxLifeTimeSec,
			MaxConn:                  &maxConn,
			MaxIdle:                  &maxIdle,
		}

		err = mydb.Init(opt)
		assert.NoError(t, err)

		// we make sure that all expectations were met
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
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.Error(t, err, "dial tcp: lookup invalidhost: no such host")
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
			Host:                     "db",
			Port:                     "invalidport",
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		fmt.Printf("error: %v\n", err)
		assert.Error(t, err, "dial tcp: lookup tcp/invalidport: unknown port")
	})

	t.Run("Failure_InvalidUser", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     "db",
			Port:                     config.SourcePort,
			Username:                 "invalid",
			Password:                 "invalid",
			DBName:                   config.SourceDataBase,
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.Error(t, err, "pq: password authentication failed for user \"invalid\"")
	})

	t.Run("Failure_InvalidDBName", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPing()

		opt := &mydb.Option{
			Host:                     "db",
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   "invaliddbname",
			ConnectionMaxLifeTimeSec: nil,
			MaxConn:                  nil,
			MaxIdle:                  nil,
		}

		err = mydb.Init(opt)
		assert.Error(t, err, "pq: database \"invaliddbname\" does not exist")
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
			Host:                     "db",
			Port:                     config.SourcePort,
			Username:                 config.SourceUser,
			Password:                 config.SourcePassword,
			DBName:                   config.SourceDataBase,
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
