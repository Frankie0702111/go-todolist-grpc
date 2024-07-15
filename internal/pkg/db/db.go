package db

import (
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"go-todolist-grpc/internal/pkg/log"
	"os"
	"path/filepath"
	"runtime"
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
	SSLMode                  string
	ConnectionMaxLifeTimeSec *int
	MaxConn                  *int
	MaxIdle                  *int
}

func Init(opt *Option) error {
	var psqlInfo string
	if opt.SSLMode == "verify-full" {
		sslRootCert := getRootCertPath()
		err := verifyCertificate(sslRootCert)
		if err != nil {
			return fmt.Errorf("certificate verification failed: %w", err)
		}

		psqlInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslrootcert=%s timezone=UTC",
			opt.Host,
			opt.Port,
			opt.Username,
			opt.Password,
			opt.DBName,
			opt.SSLMode,
			sslRootCert,
		)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC",
			opt.Host,
			opt.Port,
			opt.Username,
			opt.Password,
			opt.DBName,
			opt.SSLMode,
		)
	}

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

func getRootCertPath() string {
	// For CI to run the unit tests.
	if certContent := os.Getenv("RDS_CA_CERT"); certContent != "" {
		tempDir, err := os.MkdirTemp("", "rds-cert")
		if err != nil {
			log.Error.Printf("Failed to create temp directory: %v", err)
			return ""
		}

		tempFile := filepath.Join(tempDir, "rds-ca-2019-root.pem")
		if err := os.WriteFile(tempFile, []byte(certContent), 0600); err != nil {
			log.Error.Printf("Failed to write cert content: %v", err)
			return ""
		}

		return tempFile
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return filepath.Join(basepath, "..", "..", "config", "certs", "rds-ca-2019-root.pem")
}

func GetRootCertPath() string {
	return getRootCertPath()
}

func verifyCertificate(certPath string) error {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Verify the certificate issuer
	expectedIssuer := "CN=Amazon RDS Root 2019 CA,OU=Amazon RDS,O=Amazon Web Services\\, Inc.,L=Seattle,ST=Washington,C=US"
	if cert.Issuer.String() != expectedIssuer {
		return fmt.Errorf("unexpected certificate issuer: got %s, want %s", cert.Issuer.String(), expectedIssuer)
	}

	// Verify the certificate's validity period
	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return fmt.Errorf("certificate is not valid at the current time")
	}

	return nil
}
