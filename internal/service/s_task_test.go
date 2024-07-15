package service_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func setUpTask() error {
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

func createAuthenticatedContext(id int) context.Context {
	claims := &util.CustomClaims{
		UserID: id,
	}
	claimsJSON, _ := json.Marshal(claims)
	md := metadata.New(map[string]string{
		"x-auth-claims": string(claimsJSON),
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

type setUpTaskInfo struct {
	s          *service.Server
	ctx        context.Context
	userId     int32
	categoryId int32
}

func createUserAndCategory(t *testing.T) *setUpTaskInfo {
	err := setUpTask()
	assert.NoError(t, err)

	s := &service.Server{}

	// Register a user
	rReq := &pb.RegisterUserRequest{
		Email:    util.RandomEmail(),
		Username: util.RandomString(3),
		Password: util.RandomString(8),
	}
	rRes, err := s.RegisterUser(context.Background(), rReq)
	assert.Nil(t, err)

	// Create authenticated context
	ctx := createAuthenticatedContext(int(rRes.GetUser().Id))

	// Create a category
	cReq := &pb.CreateCategoryRequest{
		Name: util.RandomString(6),
	}
	cRes, err := s.CreateCategory(context.Background(), cReq)
	assert.Nil(t, err)

	return &setUpTaskInfo{
		s:          s,
		ctx:        ctx,
		userId:     rRes.GetUser().Id,
		categoryId: cRes.GetCategory().Id,
	}
}

func createTask(t *testing.T, setUp *setUpTaskInfo) *pb.Response {
	req := &pb.CreateTaskRequest{
		CategoryId: setUp.categoryId,
		Title:      util.RandomString(10),
		Priority:   util.RandomInt(int32(1), int32(3)),
	}
	res, err := setUp.s.CreateTask(setUp.ctx, req)
	assert.Nil(t, err)

	return res
}

func TestCreateTask(t *testing.T) {
	setUp := createUserAndCategory(t)

	t.Run("Sussess", func(t *testing.T) {
		req := &pb.CreateTaskRequest{
			CategoryId: setUp.categoryId,
			Title:      util.RandomString(10),
			Priority:   util.RandomInt(int32(1), int32(3)),
		}

		res, err := setUp.s.CreateTask(setUp.ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetTask().Id)
		assert.Equal(t, setUp.categoryId, res.GetTask().CategoryId)
	})

	t.Run("Failure_InvalidRequest", func(t *testing.T) {
		req := &pb.CreateTaskRequest{
			Title: "Invalid-task",
		}

		res, err := setUp.s.CreateTask(setUp.ctx, req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqCreateTask.CategoryId' Error:Field validation for 'CategoryId' failed on the 'required' tag\nKey: 'ReqCreateTask.Priority' Error:Field validation for 'Priority' failed on the 'required' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_Unauthenticated", func(t *testing.T) {
		unauthCtx := context.Background()
		req := &pb.CreateTaskRequest{
			CategoryId: setUp.categoryId,
			Title:      util.RandomString(10),
			Priority:   util.RandomInt(int32(1), int32(3)),
		}

		res, err := setUp.s.CreateTask(unauthCtx, req)
		assert.EqualError(t, err, "rpc error: code = Unauthenticated desc = authentication failed: no metadata found in context")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "authentication failed")
	})
}

func TestGetTask(t *testing.T) {
	setUp := createUserAndCategory(t)
	cTRes := createTask(t, setUp)

	t.Run("Sussess", func(t *testing.T) {
		req := &pb.GetTaskRequest{
			Id: cTRes.GetTask().Id,
		}

		res, err := setUp.s.GetTask(setUp.ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetTask().Id)
	})

	t.Run("Failure_NonExistentID", func(t *testing.T) {
		req := &pb.GetTaskRequest{
			Id: 999999,
		}

		res, err := setUp.s.GetTask(setUp.ctx, req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = task ID not found")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "task ID not found", st.Message())
	})
}

func TestListTask(t *testing.T) {
	setUp := createUserAndCategory(t)
	for i := 0; i < 5; i++ {
		createTask(t, setUp)
	}

	t.Run("Sussess", func(t *testing.T) {
		sortBy := "-id"
		req := &pb.ListTaskRequest{
			Page:     1,
			PageSize: 999,
			SortBy:   &sortBy,
		}

		res, err := setUp.s.ListTask(setUp.ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.NotEmpty(t, res.GetTasks())
	})

	t.Run("Failure_InvalidRequest", func(t *testing.T) {
		req := &pb.ListTaskRequest{}
		res, err := setUp.s.ListTask(setUp.ctx, req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqListTask.Page' Error:Field validation for 'Page' failed on the 'required' tag\nKey: 'ReqListTask.PageSize' Error:Field validation for 'PageSize' failed on the 'required' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_LongSortBy", func(t *testing.T) {
		sortBy := "-Failure_LongSortBy"
		req := &pb.ListTaskRequest{
			Page:     1,
			PageSize: 999,
			SortBy:   &sortBy,
		}

		res, err := setUp.s.ListTask(setUp.ctx, req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqListTask.SortBy' Error:Field validation for 'SortBy' failed on the 'max' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})

	t.Run("Failure_SortPageSize", func(t *testing.T) {
		req := &pb.ListTaskRequest{
			Page:     1,
			PageSize: 1,
		}

		res, err := setUp.s.ListTask(setUp.ctx, req)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = failed to validate: Key: 'ReqListTask.PageSize' Error:Field validation for 'PageSize' failed on the 'min' tag")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "failed to validate")
	})
}

func TestUpdateTask(t *testing.T) {
	setUp := createUserAndCategory(t)
	cTRes := createTask(t, setUp)
	newTitle := util.RandomString(10)

	t.Run("Sussess", func(t *testing.T) {
		req := &pb.UpdateTaskRequest{
			Id:    cTRes.GetTask().Id,
			Title: &newTitle,
		}

		res, err := setUp.s.UpdateTask(setUp.ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
		assert.Equal(t, newTitle, res.GetTask().Title)
	})

	t.Run("Failure_Non-ExistentID", func(t *testing.T) {
		req := &pb.UpdateTaskRequest{
			Id:    999999,
			Title: &newTitle,
		}

		res, err := setUp.s.UpdateTask(setUp.ctx, req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = task ID not found")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "task ID not found", st.Message())
	})
}

func TestDeleteTask(t *testing.T) {
	setUp := createUserAndCategory(t)
	cTRes := createTask(t, setUp)

	t.Run("Sussess", func(t *testing.T) {
		req := &pb.DeleteTaskRequest{
			Id: cTRes.GetTask().Id,
		}

		res, err := setUp.s.DeleteTask(setUp.ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int32(http.StatusOK), res.Status)
		assert.Equal(t, "ok", res.Message)
	})

	t.Run("Failure", func(t *testing.T) {
		req := &pb.DeleteTaskRequest{
			Id: 999999,
		}

		res, err := setUp.s.DeleteTask(setUp.ctx, req)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = task ID not found")
		assert.Nil(t, res)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Equal(t, "task ID not found", st.Message())
	})
}
