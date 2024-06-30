package model

import (
	"database/sql"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/db/condition"
	"go-todolist-grpc/internal/pkg/db/field"
	"time"
)

const (
	tableNameUser string = "users"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Status    bool      `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Token     string    `json:"token,omitempty" gorm:"-"`
}

func (u User) TableName() string {
	return tableNameUser
}

type UserFieldValues struct {
	ID        field.Int    `db_col:"id"`
	Username  field.String `db_col:"username"`
	Email     field.String `db_col:"email"`
	Password  field.String `db_col:"password"`
	Status    field.Bool   `db_col:"status"`
	CreatedAt field.Time   `db_col:"created_at"`
	UpdatedAt field.Time   `db_col:"updated_at"`
}

func (val UserFieldValues) TableName() string {
	return tableNameUser
}

type UserConditions struct {
	ID    *condition.Int    `db_col:"id"`
	Email *condition.String `db_col:"email"`
}

func (val UserConditions) TableName() string {
	return tableNameUser
}

func CreateUser(conn DBExecutable, values *UserFieldValues) (*UserFieldValues, error) {
	gormConn := db.GormDriver(conn)

	err := gormConn.Create(values).Error
	if err != nil {
		// if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"email_uidx\"") {
		// 	return nil, errors.New("email already exists")
		// }
		return nil, err
	}

	return values, nil
}

func getUser(conn DBExecutable, cons *UserConditions) *User {
	user := &User{}
	gormConn := db.GormDriver(conn)

	if err := gormConn.Where(BuildWhereClause(cons)).Take(user).Error; err != nil {
		return nil
	}

	return user
}

func GetUserByEmail(conn *sql.DB, email string) *User {
	cons := &UserConditions{
		Email: &condition.String{
			EQ: &email,
		},
	}

	return getUser(conn, cons)
}

func GetUserByID(conn DBExecutable, id int) *User {
	cons := &UserConditions{
		ID: &condition.Int{
			EQ: &id,
		},
	}

	return getUser(conn, cons)
}

func GetUserCount(conn *sql.DB, cons *UserConditions) (int, error) {
	var count int64

	gormConn := db.GormDriver(conn)

	stmt := gormConn.Model(User{})

	where := BuildWhereClause(cons)
	if len(where.Exprs) > 0 {
		stmt = stmt.Where(where)
	}

	if err := stmt.Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func UpdateUser(conn *sql.Tx, id int, values *UserFieldValues) error {
	return db.GormDriver(conn).Where(User{ID: id}).Updates(values).Error
}
