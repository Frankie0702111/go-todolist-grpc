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

func setUpCategory() error {
	var mockConfigContent bytes.Buffer
	mockConfigContent.WriteString("HTTP_SERVER_PORT=" + config.HttpPort + "\n")
	mockConfigContent.WriteString("GRPC_SERVER_PORT=" + config.GrpcPort + "\n")
	mockConfigContent.WriteString("DB_HOST=" + config.SourceHost + "\n")
	mockConfigContent.WriteString("DB_PORT=" + config.SourcePort + "\n")
	mockConfigContent.WriteString("DB_USER=" + config.SourceUser + "\n")
	mockConfigContent.WriteString("DB_PASS=" + config.SourcePassword + "\n")
	mockConfigContent.WriteString("DB_NAME=" + config.SourceDataBase + "\n")
	mockConfigContent.WriteString("SSL_MODE=" + config.SourceSSLMode + "\n")
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
		DBName:   config.SourceDataBase,
		SSLMode:  config.SourceSSLMode,
	}

	err = db.Init(opt)
	if err != nil {
		return err
	}

	return nil
}

func TestCreateCategory(t *testing.T) {
	err := setUpCategory()
	assert.NoError(t, err)

	name := util.RandomString(6)
	s := service.Server{}

	t.Run("Sussess", func(t *testing.T) {
		req := &pb.CreateCategoryRequest{
			Name: name,
		}

		res, err := s.CreateCategory(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetCategory().Id)
	})

	t.Run("Failure_ExistingName", func(t *testing.T) {
		req := &pb.CreateCategoryRequest{
			Name: name,
		}

		res, err := s.CreateCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = AlreadyExists desc = the category already exists")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Contains(t, "the category already exists", st.Message())
	})

	t.Run("Failure_InvalidName", func(t *testing.T) {
		req := &pb.CreateCategoryRequest{
			Name: "",
		}

		res, err := s.CreateCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqCreateCategory.Name' Error:Field validation for 'Name' failed on the 'required' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_LongName", func(t *testing.T) {
		req := &pb.CreateCategoryRequest{
			Name: util.RandomString(129),
		}

		res, err := s.CreateCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqCreateCategory.Name' Error:Field validation for 'Name' failed on the 'max' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})
}

func TestGetCategory(t *testing.T) {
	err := setUpCategory()
	assert.NoError(t, err)

	s := service.Server{}
	name := util.RandomString(6)
	rReq := &pb.CreateCategoryRequest{
		Name: name,
	}

	gRes, rErr := s.CreateCategory(context.Background(), rReq)
	assert.Nil(t, rErr)

	t.Run("Sussess", func(t *testing.T) {
		req := &pb.GetCategoryRequest{
			Id: gRes.GetCategory().Id,
		}

		res, err := s.GetCategory(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetCategory().Id)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		req := &pb.GetCategoryRequest{
			Id: 999999,
		}

		res, err := s.GetCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = category ID not found")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "category ID not found", st.Message())
	})
}

func TestListCategory(t *testing.T) {
	err := setUpCategory()
	assert.NoError(t, err)

	s := service.Server{}

	t.Run("Sussess", func(t *testing.T) {
		sortBy := "-id"
		req := &pb.ListCategoryRequest{
			Page:     1,
			PageSize: 999,
			SortBy:   &sortBy,
		}

		res, err := s.ListCategory(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetCategories())
	})

	t.Run("Failure_InvalidRequest", func(t *testing.T) {
		req := &pb.ListCategoryRequest{}
		res, err := s.ListCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqListCategory.Page' Error:Field validation for 'Page' failed on the 'required' tag\nKey: 'ReqListCategory.PageSize' Error:Field validation for 'PageSize' failed on the 'required' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_LongSortBy", func(t *testing.T) {
		sortBy := "-Failure_LongSortBy"
		req := &pb.ListCategoryRequest{
			Page:     1,
			PageSize: 999,
			SortBy:   &sortBy,
		}

		res, err := s.ListCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqListCategory.SortBy' Error:Field validation for 'SortBy' failed on the 'max' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_ShortPageSize", func(t *testing.T) {
		req := &pb.ListCategoryRequest{
			Page:     1,
			PageSize: 1,
		}

		res, err := s.ListCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqListCategory.PageSize' Error:Field validation for 'PageSize' failed on the 'min' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})
}

func TestUpdateCategory(t *testing.T) {
	err := setUpCategory()
	assert.NoError(t, err)

	s := service.Server{}
	name := util.RandomString(6)
	rReq := &pb.CreateCategoryRequest{
		Name: name,
	}

	gRes, rErr := s.CreateCategory(context.Background(), rReq)
	assert.Nil(t, rErr)

	t.Run("Sussess", func(t *testing.T) {
		newName := util.RandomString(6)
		req := &pb.UpdateCategoryRequest{
			Id:   gRes.GetCategory().Id,
			Name: &newName,
		}

		res, err := s.UpdateCategory(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.Equal(t, newName, res.GetCategory().Name)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		req := &pb.UpdateCategoryRequest{
			Id:   999999,
			Name: &name,
		}

		res, err := s.UpdateCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = category ID not found")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "category ID not found", st.Message())
	})
}

func TestDeleteCategory(t *testing.T) {
	err := setUpCategory()
	assert.NoError(t, err)

	s := service.Server{}
	name := util.RandomString(6)
	rReq := &pb.CreateCategoryRequest{
		Name: name,
	}

	gRes, rErr := s.CreateCategory(context.Background(), rReq)
	assert.Nil(t, rErr)

	t.Run("Sussess", func(t *testing.T) {
		req := &pb.DeleteCategoryRequest{
			Id: gRes.GetCategory().Id,
		}

		res, err := s.DeleteCategory(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		req := &pb.DeleteCategoryRequest{
			Id: 999999,
		}

		res, err := s.DeleteCategory(context.Background(), req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = category ID not found")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "category ID not found", st.Message())
	})
}
