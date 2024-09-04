package model_test

import (
	"database/sql"
	"fmt"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/model"
	"go-todolist-grpc/internal/pkg/util"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var sqlDBUser *sql.DB
var sqlTxUser *sql.Tx

func setUpModUser() {
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

	sqlDBUser = db
	sqlTxUser = tx
}

func setDownModUser() {
	sqlDBUser.Close()
}

func createTestUser(email, username, password string) (*model.UserFieldValues, error) {
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

	return model.CreateUser(sqlDBUser, userValues)
}

func TestCreateUser(t *testing.T) {
	setUpModUser()
	defer setDownModUser()

	t.Run("Success", func(t *testing.T) {
		_, err := createTestUser(util.RandomEmail(), util.RandomString(6), util.RandomString(8))
		assert.Nil(t, err)
	})

	t.Run("Failure_DuplicateEmail", func(t *testing.T) {
		// Insert a test user
		email := util.RandomEmail()
		password := util.RandomString(8)

		// Create the first user
		_, err := createTestUser(email, util.RandomString(6), password)
		assert.Nil(t, err)

		// Attempt to create a user with the same email
		_, err = createTestUser(email, util.RandomString(6), password)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "pq: duplicate key value violates unique constraint \"email_uidx\"")
	})
}

func TestGetUserByEmail(t *testing.T) {
	setUpModUser()
	defer setDownModUser()

	t.Run("Success", func(t *testing.T) {
		// Insert a test user
		email := util.RandomEmail()
		_, err := createTestUser(email, util.RandomString(6), util.RandomString(8))
		assert.Nil(t, err)

		// Test fetching the user by email
		user := model.GetUserByEmail(sqlDBUser, email)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
	})

	t.Run("Failure_Non-ExistentEmail", func(t *testing.T) {
		// Test fetching a non-existent user
		user := model.GetUserByEmail(sqlDBUser, "nonexistent@example.com")
		assert.Nil(t, user)
	})
}

func TestGetUserByID(t *testing.T) {
	setUpModUser()
	defer setDownModUser()

	t.Run("Success", func(t *testing.T) {
		// Insert a test user
		email := util.RandomEmail()
		createdUser, err := createTestUser(email, util.RandomString(6), util.RandomString(8))
		assert.Nil(t, err)

		// Test fetching the user by ID
		user := model.GetUserByID(sqlDBUser, createdUser.ID.Val)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID.Val, user.ID)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		// Test fetching a non-existent user
		id := 999999
		user := model.GetUserByID(sqlDBUser, id)
		assert.Nil(t, user)
	})
}

func TestGetUserCount(t *testing.T) {
	setUpModUser()
	defer setDownModUser()

	t.Run("Success", func(t *testing.T) {
		// Count users before inserting a new user
		initialCount, err := model.GetUserCount(sqlDBUser, &model.UserConditions{})
		assert.Nil(t, err)

		// Insert a test user
		email := util.RandomEmail()
		_, createUserErr := createTestUser(email, util.RandomString(6), util.RandomString(8))
		assert.Nil(t, createUserErr)

		// Count users after inserting a new user
		newCount, err := model.GetUserCount(sqlDBUser, &model.UserConditions{})
		assert.Nil(t, err)
		assert.Equal(t, initialCount+1, newCount)
	})
}

func TestUpdateUser(t *testing.T) {
	setUpModUser()
	defer setDownModUser()

	t.Run("Success", func(t *testing.T) {
		// Insert a test user
		email := util.RandomEmail()
		createdUser, err := createTestUser(email, util.RandomString(6), util.RandomString(8))
		assert.Nil(t, err)

		// Update the user's email
		newUsername := util.RandomString(6)
		updateUserErr := model.UpdateUser(sqlTxUser, createdUser.ID.Val, &model.UserFieldValues{
			Username: model.GiveColString(newUsername),
		})
		assert.Nil(t, updateUserErr)

		// Verify the update
		getdUser := model.GetUserByID(sqlTxUser, createdUser.ID.Val)
		assert.NotNil(t, getdUser)
		assert.Equal(t, newUsername, getdUser.Username)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		// Attempt to update a non-existent user
		err := model.UpdateUser(sqlTxUser, 999999, &model.UserFieldValues{
			Username: model.GiveColString(util.RandomString(6)),
		})
		assert.Nil(t, err)
	})
}
