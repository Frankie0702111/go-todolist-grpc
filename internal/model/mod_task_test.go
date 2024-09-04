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

	"github.com/stretchr/testify/assert"
)

var sqlDBTask *sql.DB
var sqlTxTask *sql.Tx

func setUpModTask() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC",
		config.SourceHost,
		config.SourcePort,
		config.SourceUser,
		config.SourcePassword,
		config.SourceDataBase,
		config.SourceSSLMode,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	tx, txErr := db.Begin()
	if txErr != nil {
		panic(txErr)
	}

	sqlDBTask = db
	sqlTxTask = tx
}

func setDownModTask() {
	sqlDBTask.Close()
}

func createTask(userId, categoryId int) (*model.TaskFieldValues, error) {
	now := time.Now().UTC()
	taskValues := &model.TaskFieldValues{
		UserId:          model.GiveColInt(userId),
		CategoryId:      model.GiveColInt(categoryId),
		Title:           model.GiveColString(util.RandomString(10)),
		Note:            model.GiveColString(util.RandomString(20)),
		Url:             model.GiveColString("http://example.com/" + util.RandomString(3)),
		SpecifyDatetime: model.GiveColNullTime(util.Pointer(now.Add(24 * time.Hour))),
		IsSpecifyTime:   model.GiveColBool(true),
		Priority:        model.GiveColInt(util.RandomInt(1, 3)),
		IsComplete:      model.GiveColBool(util.RandomBool()),
		CreatedAt:       model.GiveColTime(now),
		UpdatedAt:       model.GiveColTime(now),
	}

	return model.CreateTask(sqlDBTask, taskValues)
}

func createTestUserForTask(email, username, password string) (*model.UserFieldValues, error) {
	hashPassword, err := util.HashPassword(config.BcryptCost, password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	userValues := &model.UserFieldValues{
		Email:           model.GiveColString(email),
		Username:        model.GiveColString(username),
		Password:        model.GiveColString(hashPassword),
		Status:          model.GiveColBool(true),
		CreatedAt:       model.GiveColTime(now),
		UpdatedAt:       model.GiveColTime(now),
		IsEmailVerified: model.GiveColBool(true),
	}

	return model.CreateUser(sqlDBTask, userValues)
}

func createCategoryForTask(name string) (*model.CategoryFieldValues, error) {
	now := time.Now().UTC()
	categoryValues := &model.CategoryFieldValues{
		Name:      model.GiveColString(name),
		CreatedAt: model.GiveColTime(now),
		UpdatedAt: model.GiveColTime(now),
	}

	return model.CreateCategory(sqlDBTask, categoryValues)
}

func TestCreateTask(t *testing.T) {
	setUpModTask()
	defer setDownModTask()

	t.Run("Success", func(t *testing.T) {
		user, err := createTestUserForTask(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, err)

		category, err := createCategoryForTask(util.RandomString(6))
		assert.Nil(t, err)

		task, err := createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)
		assert.NotNil(t, task)
		assert.NotZero(t, task.ID.Val)
	})

	t.Run("Failure_InvalidID", func(t *testing.T) {
		_, err := createTask(-1, 1)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "pq: insert or update on table \"tasks\" violates foreign key constraint \"users_user_id_foreign\"")
	})
}

func TestGetTaskByTitle(t *testing.T) {
	setUpModTask()
	defer setDownModTask()

	t.Run("Success", func(t *testing.T) {
		user, err := createTestUserForTask(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, err)

		category, err := createCategoryForTask(util.RandomString(6))
		assert.Nil(t, err)

		task, err := createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)

		getTask := model.GetTaskByTitle(sqlDBTask, task.Title.Val)
		assert.NotNil(t, getTask)
		assert.Equal(t, task.Title.Val, getTask.Title)
	})

	t.Run("Failure_NonExistentTitle", func(t *testing.T) {
		task := model.GetTaskByTitle(sqlDBTask, "nonexistent-title")
		assert.Nil(t, task)
	})
}

func TestGetTaskByID(t *testing.T) {
	setUpModTask()
	defer setDownModTask()

	t.Run("Success", func(t *testing.T) {
		user, err := createTestUserForTask(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, err)

		category, err := createCategoryForTask(util.RandomString(6))
		assert.Nil(t, err)

		task, err := createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)

		getTask := model.GetTaskByID(sqlDBTask, task.ID.Val)
		assert.NotNil(t, getTask)
		assert.Equal(t, task.ID.Val, getTask.ID)
	})

	t.Run("Failure", func(t *testing.T) {
		task := model.GetTaskByID(sqlDBTask, 999999)
		assert.Nil(t, task)
	})
}

func TestListTask(t *testing.T) {
	setUpModTask()
	defer setDownModTask()

	t.Run("Success", func(t *testing.T) {
		user, userErr := createTestUserForTask(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, userErr)

		category, categoryErr := createCategoryForTask(util.RandomString(6))
		assert.Nil(t, categoryErr)

		_, err := createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)
		_, err = createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)

		// Get task count for assertions
		getTaskCount, err := model.GetTaskCount(sqlDBTask, &model.TaskConditions{
			UserId: &condition.Int{
				EQ: &user.ID.Val,
			},
		})
		assert.Nil(t, err)

		conditions := &model.TaskConditions{
			UserId: &condition.Int{
				EQ: &user.ID.Val,
			},
		}
		orderBys := &model.TaskOrderBy{}
		orderBys.Parse(service.ParseSortBy("-id"))
		limit := 999999
		offset := 0

		tasks := model.ListTask(sqlDBTask, conditions, orderBys, &limit, &offset)
		assert.NotNil(t, tasks)
		assert.Len(t, tasks, int(getTaskCount))
		assert.GreaterOrEqual(t, len(tasks), int(getTaskCount))
	})

	t.Run("Failure_EmptyResult", func(t *testing.T) {
		nonExistentUserID := 99999
		conditions := &model.TaskConditions{
			UserId: &condition.Int{
				EQ: &nonExistentUserID,
			},
		}
		orderBys := &model.TaskOrderBy{}
		limit := 999999
		offset := 0

		tasks := model.ListTask(sqlDBTask, conditions, orderBys, &limit, &offset)
		assert.NotNil(t, tasks)
		assert.Len(t, tasks, 0)
	})
}

func TestGetTaskCount(t *testing.T) {
	setUpModTask()
	defer setDownModTask()

	t.Run("Success", func(t *testing.T) {
		user, userErr := createTestUserForTask(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, userErr)

		category, categoryErr := createCategoryForTask(util.RandomString(6))
		assert.Nil(t, categoryErr)

		initialCount, initialCountErr := model.GetTaskCount(sqlDBTask, &model.TaskConditions{
			UserId: &condition.Int{
				EQ: &user.ID.Val,
			},
		})
		assert.Nil(t, initialCountErr)

		_, err := createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)

		newCount, err := model.GetTaskCount(sqlDBTask, &model.TaskConditions{
			UserId: &condition.Int{
				EQ: &user.ID.Val,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, initialCount+1, newCount)
	})
}

func TestUpdateTask(t *testing.T) {
	setUpModTask()
	defer setDownModTask()

	t.Run("Success", func(t *testing.T) {
		user, userErr := createTestUserForTask(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, userErr)

		category, categoryErr := createCategoryForTask(util.RandomString(6))
		assert.Nil(t, categoryErr)

		task, err := createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)

		newTitle := util.RandomString(10)
		updateTaskErr := model.UpdateTask(sqlTxTask, task.ID.Val, &model.TaskFieldValues{
			Title: model.GiveColString(newTitle),
		})
		assert.Nil(t, updateTaskErr)

		getTask := model.GetTaskByID(sqlTxTask, task.ID.Val)
		assert.NotNil(t, getTask)
		assert.Equal(t, newTitle, getTask.Title)
	})

	t.Run("Failure_NonExistentID", func(t *testing.T) {
		err := model.UpdateTask(sqlTxTask, 99999, &model.TaskFieldValues{
			Title: model.GiveColString(util.RandomString(10)),
		})
		assert.Nil(t, err)
	})
}

func TestDeleteTask(t *testing.T) {
	setUpModTask()
	defer setDownModTask()

	t.Run("Success", func(t *testing.T) {
		user, userErr := createTestUserForTask(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, userErr)

		category, categoryErr := createCategoryForTask(util.RandomString(6))
		assert.Nil(t, categoryErr)

		task, err := createTask(user.ID.Val, category.ID.Val)
		assert.Nil(t, err)

		deleteErr := model.DeleteTask(sqlTxTask, task.ID.Val)
		assert.Nil(t, deleteErr)

		getTask := model.GetTaskByID(sqlTxTask, task.ID.Val)
		assert.Nil(t, getTask)
	})

	t.Run("Failure", func(t *testing.T) {
		err := model.DeleteTask(sqlTxTask, 99999)
		assert.Nil(t, err)
	})
}
