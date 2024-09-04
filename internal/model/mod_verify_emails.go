package model

import (
	"database/sql"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/db/condition"
	"go-todolist-grpc/internal/pkg/db/field"
	"time"
)

const (
	tableNameVerifyEmails string = "verify_emails"
)

type VerifyEmail struct {
	ID         int       `json:"id"`
	UserId     int       `json:"user_id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	SecretCode string    `json:"-"`
	IsUsed     bool      `json:"is_used"`
	ExpiredAt  time.Time `json:"-"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (u VerifyEmail) TableName() string {
	return tableNameVerifyEmails
}

type VerifyEmailFieldValues struct {
	ID         field.Int    `db_col:"id"`
	UserId     field.Int    `db_col:"user_id"`
	Username   field.String `db_col:"username"`
	Email      field.String `db_col:"email"`
	SecretCode field.String `db_col:"secret_code"`
	IsUsed     field.Bool   `db_col:"is_used"`
	ExpiredAt  field.Time   `db_col:"expired_at"`
	CreatedAt  field.Time   `db_col:"created_at"`
	UpdatedAt  field.Time   `db_col:"updated_at"`
}

func (val VerifyEmailFieldValues) TableName() string {
	return tableNameVerifyEmails
}

type VerifyEmailConditions struct {
	ID        *condition.Int    `db_col:"id"`
	Email     *condition.String `db_col:"email"`
	IsUsed    *condition.Bool   `db_col:"is_used"`
	ExpiredAt *condition.Time   `db_col:"expired_at"`
}

func (val VerifyEmailConditions) TableName() string {
	return tableNameVerifyEmails
}

func CreateVerifyEmail(conn DBExecutable, values *VerifyEmailFieldValues) (*VerifyEmailFieldValues, error) {
	gormConn := db.GormDriver(conn)

	err := gormConn.Create(values).Error
	if err != nil {
		return nil, err
	}

	return values, nil
}

func getVerifyEmail(conn DBExecutable, cons *VerifyEmailConditions) *VerifyEmail {
	verifyEmail := &VerifyEmail{}
	gormConn := db.GormDriver(conn)

	if err := gormConn.Where(BuildWhereClause(cons)).Take(verifyEmail).Error; err != nil {
		return nil
	}

	return verifyEmail
}

func GetVerifyEmailByID(conn DBExecutable, id int, is_used bool, expired_at *time.Time) *VerifyEmail {
	cons := &VerifyEmailConditions{
		ID: &condition.Int{
			EQ: &id,
		},
		IsUsed: &condition.Bool{
			EQ: &is_used,
		},
		ExpiredAt: &condition.Time{
			GTE: expired_at,
		},
	}

	return getVerifyEmail(conn, cons)
}

func UpdateVerifyEmail(conn *sql.Tx, id int, values *VerifyEmailFieldValues) error {
	return db.GormDriver(conn).Where(VerifyEmail{ID: id}).Updates(values).Error
}
