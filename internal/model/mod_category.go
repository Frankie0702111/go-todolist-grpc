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
	tableNameCategory string = "categories"
)

type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (u Category) TableName() string {
	return tableNameCategory
}

type CategoryFieldValues struct {
	ID        field.Int    `db_col:"id"`
	Name      field.String `db_col:"name"`
	CreatedAt field.Time   `db_col:"created_at"`
	UpdatedAt field.Time   `db_col:"updated_at"`
}

func (val CategoryFieldValues) TableName() string {
	return tableNameCategory
}

type CategoryConditions struct {
	ID   *condition.Int    `db_col:"id"`
	Name *condition.String `db_col:"name"`
}

func (val CategoryConditions) TableName() string {
	return tableNameCategory
}

type CategoryOrderBy struct {
	ID   *builder.OrderBy `db_col:"id"`
	Name *builder.OrderBy `db_col:"name"`
}

func (ob CategoryOrderBy) TableName() string {
	return tableNameCategory
}

func (ob *CategoryOrderBy) Parse(params map[string]bool) {
	ParseOrderByParams(params, ob)
}

func CreateCategory(conn *sql.DB, values *CategoryFieldValues) (*CategoryFieldValues, error) {
	gormConn := db.GormDriver(conn)

	if err := gormConn.Create(values).Error; err != nil {
		return nil, err
	}

	return values, nil
}

func getCategory(conn DBExecutable, cons *CategoryConditions) *Category {
	category := &Category{}
	gormConn := db.GormDriver(conn)

	if err := gormConn.Where(BuildWhereClause(cons)).Take(category).Error; err != nil {
		return nil
	}

	return category
}

func GetCategoryByName(conn *sql.DB, name string) *Category {
	cons := &CategoryConditions{
		Name: &condition.String{
			EQ: &name,
		},
	}

	return getCategory(conn, cons)
}

func GetCategoryByID(conn DBExecutable, id int) *Category {
	cons := &CategoryConditions{
		ID: &condition.Int{
			EQ: &id,
		},
	}

	return getCategory(conn, cons)
}

func ListCategory(conn *sql.DB, cons *CategoryConditions, orderBys *CategoryOrderBy, limit *int, offset *int) []Category {
	categories := make([]Category, 0)

	stmt := db.GormDriver(conn).Model(Category{}).Preload(clause.Associations)

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

	if err := stmt.Find(&categories).Error; err != nil {
		return categories
	}

	return categories
}

func GetCategoryCount(conn *sql.DB, cons *CategoryConditions) (int64, error) {
	var count int64

	stmt := db.GormDriver(conn).Model(Category{})

	where := BuildWhereClause(cons)
	if len(where.Exprs) > 0 {
		stmt = stmt.Where(where)
	}

	if err := stmt.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func UpdateCategory(conn *sql.Tx, id int, values *CategoryFieldValues) error {
	return db.GormDriver(conn).Where(Category{ID: id}).Updates(values).Error
}

func DeleteCategory(conn DBExecutable, id int) error {
	return db.GormDriver(conn).Delete(&Category{}, id).Error
}
