package db

import (
	"database/sql"
	"fmt"
	"go-todolist-grpc/internal/pkg/log"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gRPCDB *sql.DB

type Option struct {
	Host                     string
	Port                     string
	Username                 string
	Password                 string
	DBName                   string
	ConnectionMaxLifeTimeSec *int
	MaxConn                  *int
	MaxIdle                  *int
}

func Init(opt *Option) error {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC",
		opt.Host,
		opt.Port,
		opt.Username,
		opt.Password,
		opt.DBName,
	)

	log.Info.Printf("initial gRPC DB: %s:%s", opt.Host, opt.Port)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	if opt.ConnectionMaxLifeTimeSec != nil {
		log.Info.Printf("initial gRPC DB: set connection max life time: %d seconds", *opt.ConnectionMaxLifeTimeSec)
		db.SetConnMaxLifetime(time.Second * time.Duration(*opt.ConnectionMaxLifeTimeSec))
	}

	if opt.MaxConn != nil {
		log.Info.Printf("initial gRPC DB: set max open connections: %d", *opt.MaxConn)
		db.SetMaxOpenConns(*opt.MaxConn)
	}

	if opt.MaxIdle != nil {
		log.Info.Printf("initial gRPC DB: set max idle connections: %d", *opt.MaxIdle)
		db.SetMaxIdleConns(*opt.MaxIdle)
	}

	gRPCDB = db

	log.Info.Println("gRPC DB connection successfully")

	return nil
}

func GetConn() *sql.DB {
	if gRPCDB == nil {
		return nil
	}
	return gRPCDB
}

func GormDriver(db gorm.ConnPool) *gorm.DB {
	conn, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})

	if err != nil {
		log.Error.Printf("Open gorm connection failed: %v", err)
		panic(err)
	}

	return conn
}

func ResetConn() {
	gRPCDB = nil
}
