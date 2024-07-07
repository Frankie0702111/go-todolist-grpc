package model

import (
	"database/sql"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/db/builder"
	"go-todolist-grpc/internal/pkg/db/condition"
	"go-todolist-grpc/internal/pkg/db/field"
	"time"

	"gorm.io/gorm/clause"
)

const (
	tableNameTask string = "tasks"
)

type Task struct {
	ID              int       `json:"id"`
	UserId          int       `json:"user_id"`
	CategoryId      int       `json:"category_id"`
	Title           string    `json:"title"`
	Note            string    `json:"note"`
	Url             string    `json:"url"`
	SpecifyDatetime time.Time `json:"specify_datetime"`
	IsSpecifyTime   bool      `json:"is_specify_time"`
	Priority        int       `json:"priority"`
	IsComplete      bool      `json:"is_complete"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (u Task) TableName() string {
	return tableNameTask
}

type TaskFieldValues struct {
	ID              field.Int      `db_col:"id"`
	UserId          field.Int      `db_col:"user_id"`
	CategoryId      field.Int      `db_col:"category_id"`
	Title           field.String   `db_col:"title"`
	Note            field.String   `db_col:"note"`
	Url             field.String   `db_col:"url"`
	SpecifyDatetime field.NullTime `db_col:"specify_datetime"`
	IsSpecifyTime   field.Bool     `db_col:"is_specify_time"`
	Priority        field.Int      `db_col:"priority"`
	IsComplete      field.Bool     `db_col:"is_complete"`
	CreatedAt       field.Time     `db_col:"created_at"`
	UpdatedAt       field.Time     `db_col:"updated_at"`
}

func (val TaskFieldValues) TableName() string {
	return tableNameTask
}

type TaskConditions struct {
	ID            *condition.Int    `db_col:"id"`
	UserId        *condition.Int    `db_col:"user_id"`
	CategoryId    *condition.Int    `db_col:"category_id"`
	Title         *condition.String `db_col:"title"`
	IsSpecifyTime *condition.Bool   `db_col:"is_specify_time"`
	Priority      *condition.Int    `db_col:"priority"`
	IsComplete    *condition.Bool   `db_col:"is_complete"`
}

func (val TaskConditions) TableName() string {
	return tableNameTask
}

type TaskOrderBy struct {
	ID         *builder.OrderBy `db_col:"id"`
	CategoryId *builder.OrderBy `db_col:"category_id"`
	Priority   *builder.OrderBy `db_col:"priority"`
}

func (ob TaskOrderBy) TableName() string {
	return tableNameTask
}

func (ob *TaskOrderBy) Parse(params map[string]bool) {
	ParseOrderByParams(params, ob)
}

func CreateTask(conn *sql.DB, values *TaskFieldValues) (*TaskFieldValues, error) {
	gormConn := db.GormDriver(conn)

	if err := gormConn.Create(values).Error; err != nil {
		return nil, err
	}

	return values, nil
}

func getTask(conn DBExecutable, cons *TaskConditions) *Task {
	task := &Task{}
	gormConn := db.GormDriver(conn)

	if err := gormConn.Where(BuildWhereClause(cons)).Take(task).Error; err != nil {
		return nil
	}

	return task
}

func GetTaskByTitle(conn *sql.DB, title string) *Task {
	cons := &TaskConditions{
		Title: &condition.String{
			EQ: &title,
		},
	}

	return getTask(conn, cons)
}

func GetTaskByID(conn DBExecutable, id int) *Task {
	cons := &TaskConditions{
		ID: &condition.Int{
			EQ: &id,
		},
	}

	return getTask(conn, cons)
}

func ListTask(conn *sql.DB, cons *TaskConditions, orderBys *TaskOrderBy, limit *int, offset *int) []Task {
	tasks := make([]Task, 0)

	stmt := db.GormDriver(conn).Model(Task{}).Preload(clause.Associations)

	// conditions
	where := BuildWhereClause(cons)
	if len(where.Exprs) > 0 {
		stmt = stmt.Where(where)
	}

	// sorting
	orderBy := BuildOrderByClause(orderBys)
	if len(orderBy.Columns) > 0 {
		stmt = stmt.Clauses(orderBy)
	}

	if limit != nil {
		stmt = stmt.Limit(*limit)
	}

	if offset != nil {
		stmt = stmt.Offset(*offset)
	}

	if err := stmt.Find(&tasks).Error; err != nil {
		return tasks
	}

	return tasks
}

func GetTaskCount(conn *sql.DB, cons *TaskConditions) (int32, error) {
	var count int64

	stmt := db.GormDriver(conn).Model(Task{})

	where := BuildWhereClause(cons)
	if len(where.Exprs) > 0 {
		stmt = stmt.Where(where)
	}

	if err := stmt.Count(&count).Error; err != nil {
		return 0, err
	}

	return int32(count), nil
}

func UpdateTask(conn *sql.Tx, id int, values *TaskFieldValues) error {
	return db.GormDriver(conn).Where(Task{ID: id}).Updates(values).Error
}

func DeleteTask(conn DBExecutable, id int) error {
	return db.GormDriver(conn).Delete(&Task{}, id).Error
}
