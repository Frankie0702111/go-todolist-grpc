package model_test

import (
	"database/sql"
	"fmt"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/model"
	"go-todolist-grpc/internal/pkg/db/condition"
	"go-todolist-grpc/internal/pkg/util"
	"go-todolist-grpc/internal/service"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var sqlDBCategory *sql.DB
var sqlTxCategory *sql.Tx

func setUpModCategory() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s application_name=otter sslmode=disable timezone=UTC",
		config.SourceHost,
		config.SourcePort,
		config.SourceUser,
		config.SourcePassword,
		config.TestSourceDataBase,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	tx, txErr := db.Begin()
	if txErr != nil {
		panic(txErr)
	}

	sqlDBCategory = db
	sqlTxCategory = tx
}

func setDownModCategory() {
	sqlDBCategory.Close()
}

func createCategory(name string) (*model.CategoryFieldValues, error) {
	now := time.Now().UTC()
	categoryValues := &model.CategoryFieldValues{
		Name:      model.GiveColString(name),
		CreatedAt: model.GiveColTime(now),
		UpdatedAt: model.GiveColTime(now),
	}

	return model.CreateCategory(sqlDBCategory, categoryValues)
}

func TestCreateCategory(t *testing.T) {
	setUpModCategory()
	defer setDownModCategory()

	t.Run("Success", func(t *testing.T) {
		category, err := createCategory(util.RandomString(6))
		assert.Nil(t, err)
		assert.NotNil(t, category)
		assert.NotZero(t, category.ID.Val)
	})

	t.Run("Failure_DuplicateName", func(t *testing.T) {
		name := util.RandomString(6)
		_, err := createCategory(name)
		assert.Nil(t, err)

		_, err = createCategory(name)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "pq: duplicate key value violates unique constraint \"name_uidx\"")
	})
}

func TestGetCategoryByName(t *testing.T) {
	setUpModCategory()
	defer setDownModCategory()

	t.Run("Success", func(t *testing.T) {
		name := util.RandomString(6)
		_, err := createCategory(name)
		assert.Nil(t, err)

		category := model.GetCategoryByName(sqlDBCategory, name)
		assert.NotNil(t, category)
		assert.Equal(t, name, category.Name)
	})

	t.Run("Failure_Non-ExistentName", func(t *testing.T) {
		category := model.GetCategoryByName(sqlDBCategory, "nonexistent-category")
		assert.Nil(t, category)
	})
}

func TestGetCategoryByID(t *testing.T) {
	setUpModCategory()
	defer setDownModCategory()

	t.Run("Success", func(t *testing.T) {
		name := util.RandomString(6)
		createCategory, err := createCategory(name)
		assert.Nil(t, err)

		category := model.GetCategoryByID(sqlDBCategory, createCategory.ID.Val)
		assert.NotNil(t, category)
		assert.Equal(t, createCategory.ID.Val, category.ID)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		category := model.GetCategoryByID(sqlDBCategory, 999999)
		assert.Nil(t, category)
	})
}

func TestListCategory(t *testing.T) {
	setUpModCategory()
	defer setDownModCategory()

	t.Run("Success", func(t *testing.T) {
		// Create categories for testing
		_, err := createCategory(util.RandomString(6))
		assert.Nil(t, err)
		_, err = createCategory(util.RandomString(6))
		assert.Nil(t, err)

		// Get category count for assertions
		getCategoryCount, err := model.GetCategoryCount(sqlDBCategory, &model.CategoryConditions{})
		assert.Nil(t, err)

		// Define conditions for listing categories
		conditions := &model.CategoryConditions{}
		orderBys := &model.CategoryOrderBy{}
		orderBys.Parse(service.ParseSortBy("-id"))
		limit := 999999
		offset := 0

		categories := model.ListCategory(sqlDBCategory, conditions, orderBys, &limit, &offset)
		assert.NotNil(t, categories)
		assert.Len(t, categories, int(getCategoryCount))
		assert.GreaterOrEqual(t, len(categories), int(getCategoryCount))
	})

	t.Run("Failure_EmptyResult", func(t *testing.T) {
		// Define conditions that should return an empty result
		name := "nonexistent-category"
		conditions := &model.CategoryConditions{
			Name: &condition.String{
				EQ: &name,
			},
		}
		orderBys := &model.CategoryOrderBy{}
		limit := 999999
		offset := 0

		categories := model.ListCategory(sqlDBCategory, conditions, orderBys, &limit, &offset)
		assert.NotNil(t, categories)
		assert.Len(t, categories, 0)
	})
}

func TestGetCategoryCount(t *testing.T) {
	setUpModCategory()
	defer setDownModCategory()

	t.Run("Success", func(t *testing.T) {
		initialCount, err := model.GetCategoryCount(sqlDBCategory, &model.CategoryConditions{})
		assert.Nil(t, err)

		name := util.RandomString(6)
		_, createCategoryErr := createCategory(name)
		assert.Nil(t, createCategoryErr)

		newCount, err := model.GetCategoryCount(sqlDBCategory, &model.CategoryConditions{})
		assert.Nil(t, err)
		assert.Equal(t, initialCount+1, newCount)
	})
}

func TestUpdateCategory(t *testing.T) {
	setUpModCategory()
	defer setDownModCategory()

	t.Run("Success", func(t *testing.T) {
		name := util.RandomString(6)
		createCategory, err := createCategory(name)
		assert.Nil(t, err)

		newName := util.RandomString(6)
		updateCategoryErr := model.UpdateCategory(sqlTxCategory, createCategory.ID.Val, &model.CategoryFieldValues{
			Name: model.GiveColString(newName),
		})
		assert.Nil(t, updateCategoryErr)

		getCategory := model.GetCategoryByID(sqlTxCategory, createCategory.ID.Val)
		assert.NotNil(t, getCategory)
		assert.Equal(t, newName, getCategory.Name)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		err := model.UpdateCategory(sqlTxCategory, 999999, &model.CategoryFieldValues{
			Name: model.GiveColString(util.RandomString(6)),
		})
		assert.Nil(t, err)
	})
}

func TestDeleteCategory(t *testing.T) {
	setUpModCategory()
	defer setDownModCategory()

	t.Run("Success", func(t *testing.T) {
		name := util.RandomString(6)
		createCategory, err := createCategory(name)
		assert.Nil(t, err)

		deleteErr := model.DeleteCategory(sqlTxCategory, createCategory.ID.Val)
		assert.Nil(t, deleteErr)

		getategory := model.GetCategoryByID(sqlTxCategory, createCategory.ID.Val)
		assert.Nil(t, getategory)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		err := model.DeleteCategory(sqlTxCategory, 999999)
		assert.Nil(t, err)
	})
}
