package service_test

import (
	"bytes"
	"context"
	"go-todolist-grpc/api/pb"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/log"
	"go-todolist-grpc/internal/pkg/util"
	"go-todolist-grpc/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setUpUser() error {
	var mockConfigContent bytes.Buffer
	mockConfigContent.WriteString("HTTP_SERVER_PORT=" + config.HttpPort + "\n")
	mockConfigContent.WriteString("GRPC_SERVER_PORT=" + config.GrpcPort + "\n")
	mockConfigContent.WriteString("DB_HOST=" + config.SourceHost + "\n")
	mockConfigContent.WriteString("DB_PORT=" + config.SourcePort + "\n")
	mockConfigContent.WriteString("DB_USER=" + config.SourceUser + "\n")
	mockConfigContent.WriteString("DB_PASS=" + config.SourcePassword + "\n")
	mockConfigContent.WriteString("DB_NAME=" + config.TestSourceDataBase + "\n")
	mockConfigContent.WriteString("DB_CONN_MAX_LT_SEC=" + strconv.Itoa(config.SourceDBConnMaxLTSec) + "\n")
	mockConfigContent.WriteString("DB_MAX_CONN=" + strconv.Itoa(config.SourceMaxConn) + "\n")
	mockConfigContent.WriteString("DB_MAX_IDLE=" + strconv.Itoa(config.SourceMaxIdle) + "\n")
	mockConfigContent.WriteString("BCRYPT_COST=" + strconv.Itoa(config.BcryptCost) + "\n")
	mockConfigContent.WriteString("JWT_SECRET_KEY=" + config.JwtSecretKey + "\n")
	mockConfigContent.WriteString("JWT_TTL=" + strconv.Itoa(config.JwtTtl) + "\n")
	mockConfigContent.WriteString("LOG_LEVEL=" + strconv.Itoa(config.LogLevel) + "\n")
	mockConfigContent.WriteString("LOG_FOLDER_PATH=" + config.LogFolderPath + "\n")
	mockConfigContent.WriteString("ENABLE_CONSOLE_OUTPUT=" + strconv.FormatBool(config.EnableConsoleOutput) + "\n")
	mockConfigContent.WriteString("ENABLE_FILE_OUTPUT=" + strconv.FormatBool(config.EnableFileOutput) + "\n")

	// Create app.env file
	appFolderPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	mockConfigFile := filepath.Join(appFolderPath, "app.env")
	err := os.WriteFile(mockConfigFile, mockConfigContent.Bytes(), 0644)
	defer os.Remove(mockConfigFile)
	if err != nil {
		return err
	}

	// Init config
	loadErr := config.Load()
	if loadErr != nil {
		return loadErr
	}

	// Init log
	log.Init(config.LogLevel, config.LogFolderPath, strconv.Itoa(os.Getpid()), config.EnableConsoleOutput, config.EnableFileOutput)

	// Init sql
	opt := &db.Option{
		Host:     config.SourceHost,
		Port:     config.SourcePort,
		Username: config.SourceUser,
		Password: config.SourcePassword,
		DBName:   config.TestSourceDataBase,
	}

	err = db.Init(opt)
	if err != nil {
		return err
	}

	return nil
}

func TestRegisterUser(t *testing.T) {
	err := setUpUser()
	assert.NoError(t, err)

	s := service.Server{}
	email := util.RandomEmail()

	t.Run("Success", func(t *testing.T) {
		req := &pb.RegisterUserRequest{
			Email:    email,
			Username: util.RandomString(3),
			Password: util.RandomString(72),
		}

		res, err := s.RegisterUser(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetUser().Id)
	})

	t.Run("Failure_ExistingEmail", func(t *testing.T) {
		req := &pb.RegisterUserRequest{
			Email:    email, // Email already registered in previous test
			Username: util.RandomString(6),
			Password: util.RandomString(8),
		}

		res, err := s.RegisterUser(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = AlreadyExists desc = the email already exists")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Equal(t, "the email already exists", st.Message())
	})

	t.Run("Failure_InvalidEmail", func(t *testing.T) {
		req := &pb.RegisterUserRequest{
			Email:    "invalid-email",
			Username: util.RandomString(6),
			Password: util.RandomString(8),
		}

		res, err := s.RegisterUser(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqRegister.Email' Error:Field validation for 'Email' failed on the 'email' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_ShortPassword", func(t *testing.T) {
		req := &pb.RegisterUserRequest{
			Email:    util.RandomEmail(),
			Username: util.RandomString(6),
			Password: util.RandomString(7),
		}

		res, err := s.RegisterUser(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqRegister.Password' Error:Field validation for 'Password' failed on the 'min' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_HashLongPassword", func(t *testing.T) {
		req := &pb.RegisterUserRequest{
			Email:    util.RandomEmail(),
			Username: util.RandomString(6),
			Password: util.RandomString(73),
		}

		res, err := s.RegisterUser(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = Internal desc = failed to hash password: bcrypt: password length exceeds 72 bytes")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to hash password")
	})
}

func TestLogin(t *testing.T) {
	err := setUpUser()
	assert.NoError(t, err)

	// Insert a test user
	s := service.Server{}
	email := util.RandomEmail()
	password := util.RandomString(8)
	rReq := &pb.RegisterUserRequest{
		Email:    email,
		Username: util.RandomString(6),
		Password: password,
	}

	_, rErr := s.RegisterUser(context.Background(), rReq)
	assert.Nil(t, rErr)

	t.Run("Success", func(t *testing.T) {
		lReq := &pb.LoginRequest{
			Email:    email,
			Password: password,
		}

		res, err := s.Login(context.Background(), lReq)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetUser().Token)
	})

	t.Run("Failure_IncorrectPassword", func(t *testing.T) {
		req := &pb.LoginRequest{
			Email:    email,
			Password: "invalid-password",
		}

		res, err := s.Login(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = password is incorrect")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "password is incorrect", st.Message())
	})

	t.Run("Failure_UnregisteredEmail", func(t *testing.T) {
		req := &pb.LoginRequest{
			Email:    util.RandomEmail(),
			Password: util.RandomString(8),
		}

		res, err := s.Login(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = this email is unregistered")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "this email is unregistered", st.Message())
	})
}

func TestUpdateUser(t *testing.T) {
	err := setUpUser()
	assert.NoError(t, err)

	// Insert a test user
	s := service.Server{}
	rReq := &pb.RegisterUserRequest{
		Email:    util.RandomEmail(),
		Username: util.RandomString(6),
		Password: util.RandomString(8),
	}

	rRes, err := s.RegisterUser(context.Background(), rReq)
	assert.Nil(t, err)

	t.Run("Success", func(t *testing.T) {
		newUsername := util.RandomString(6)
		newPassword := util.RandomString(8)
		req := &pb.UpdateUserRequest{
			UserId:   rRes.GetUser().Id,
			Username: &newUsername,
			Password: &newPassword,
		}

		res, err := s.UpdateUser(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.Equal(t, newUsername, res.GetUser().Username)
		assert.NotEmpty(t, res.GetUser().Email)
	})

	t.Run("Failure_InvalidUserID", func(t *testing.T) {
		newUsername := util.RandomString(6)
		newPassword := util.RandomString(8)
		req := &pb.UpdateUserRequest{
			UserId:   999999,
			Username: &newUsername,
			Password: &newPassword,
		}

		res, err := s.UpdateUser(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = user ID not found")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "user ID not found", st.Message())
	})

	t.Run("Failure_EmptyUsername", func(t *testing.T) {
		newUsername := ""
		newPassword := util.RandomString(8)
		req := &pb.UpdateUserRequest{
			UserId:   rRes.GetUser().Id,
			Username: &newUsername,
			Password: &newPassword,
		}

		res, err := s.UpdateUser(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqUpdateUser.Username' Error:Field validation for 'Username' failed on the 'min' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})
}
