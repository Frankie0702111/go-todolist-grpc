package service

import (
	"context"
	"go-todolist-grpc/api/pb"
	"go-todolist-grpc/internal/middleware"
	"go-todolist-grpc/internal/model"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/db/condition"
	"go-todolist-grpc/internal/pkg/log"
	"go-todolist-grpc/internal/pkg/util"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReqCreateTask struct {
	CategoryId      int32   `json:"category_id" validate:"required"`
	Title           string  `json:"title" validate:"required,max=100"`
	Note            *string `json:"note" validate:"omitempty,max=255"`
	Url             *string `json:"url" validate:"omitempty,max=255"`
	SpecifyDatetime *int64  `json:"specify_datetime" validate:"omitempty,min=1"`
	Priority        int32   `json:"priority" validate:"required,oneof=1 2 3"`
}

func (ins ReqCreateTask) toFieldValues() model.TaskFieldValues {
	now := time.Now().UTC()
	fv := model.TaskFieldValues{}
	fv.CategoryId = model.GiveColInt(int(ins.CategoryId))
	fv.Title = model.GiveColString(ins.Title)

	if ins.Note != nil {
		fv.Note = model.GiveColString(*ins.Note)
	}

	if ins.Url != nil {
		fv.Url = model.GiveColString(*ins.Url)
	}

	var t *time.Time
	isSpecifyTime := false
	if ins.SpecifyDatetime != nil {
		t = Pointer(time.Unix(*ins.SpecifyDatetime/1000, 0))
		isSpecifyTime = true
	}
	fv.SpecifyDatetime = model.GiveColNullTime(t)
	fv.IsSpecifyTime = model.GiveColBool(isSpecifyTime)

	fv.Priority = model.GiveColInt(int(ins.Priority))
	fv.CreatedAt = model.GiveColTime(now)
	fv.UpdatedAt = model.GiveColTime(now)

	return fv
}

func (s *Server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Response, error) {
	claims, err := middleware.GetClaimsFromContext(ctx)
	if err != nil {
		log.Error.Printf("Failed to get user ID: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
	}

	// Validate request
	reqTask := &ReqCreateTask{}
	if err := bindRequest(req, reqTask); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	// Check if the task title is already exists
	conn := db.GetConn()
	getTask := model.GetTaskByTitle(conn, reqTask.Title)
	if getTask != nil {
		return nil, status.Errorf(codes.AlreadyExists, "the task already exists")
	}

	insFields := reqTask.toFieldValues()
	insFields.UserId = model.GiveColInt(claims.UserID)

	task, taskErr := model.CreateTask(conn, &insFields)
	if taskErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", taskErr)
	}

	taskInfo := &pb.Task{
		Id:              int32(task.ID.Val),
		UserId:          int32(task.UserId.Val),
		CategoryId:      int32(task.CategoryId.Val),
		Title:           task.Title.Val,
		Note:            task.Note.Val,
		Url:             task.Url.Val,
		SpecifyDatetime: util.GetFullDateStrFromPtr(&task.SpecifyDatetime.Val),
		IsSpecifyTime:   task.IsSpecifyTime.Val,
		Priority:        int32(task.Priority.Val),
		IsComplete:      task.IsComplete.Val,
		CreatedAt:       util.GetFullDateStr(task.CreatedAt.Val),
		UpdatedAt:       util.GetFullDateStr(task.UpdatedAt.Val),
	}

	return &pb.Response{
		Data: &pb.Response_Task{
			Task: taskInfo,
		},
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}

func (s *Server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Response, error) {
	conn := db.GetConn()

	// Validate request
	reqGet := &ReqId{}
	if err := bindRequest(req, reqGet); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	taskId := int(reqGet.Id)
	getTask := model.GetTaskByID(conn, taskId)
	if getTask == nil {
		return nil, status.Errorf(codes.NotFound, "task ID not found")
	}

	return &pb.Response{
		Data: &pb.Response_Task{
			Task: &pb.Task{
				Id:              int32(getTask.ID),
				UserId:          int32(getTask.UserId),
				CategoryId:      int32(getTask.CategoryId),
				Title:           getTask.Title,
				Note:            getTask.Note,
				Url:             getTask.Url,
				SpecifyDatetime: util.GetFullDateStrFromPtr(&getTask.SpecifyDatetime),
				IsSpecifyTime:   getTask.IsSpecifyTime,
				Priority:        int32(getTask.Priority),
				IsComplete:      getTask.IsComplete,
				CreatedAt:       util.GetFullDateStr(getTask.CreatedAt),
				UpdatedAt:       util.GetFullDateStr(getTask.UpdatedAt),
			},
		},
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}

type ReqListTask struct {
	Page          int32   `json:"page" validate:"required,min=1,max=100000"`
	PageSize      int32   `json:"page_size" validate:"required,min=5,max=1000"`
	SortBy        *string `json:"sort_by" validate:"omitempty,max=15"`
	TaskId        *int32  `json:"task_id" validate:"omitempty,min=1"`
	CategoryId    *int32  `json:"category_id" validate:"omitempty,min=1"`
	Title         *string `json:"title" validate:"omitempty,max=100"`
	IsSpecifyTime *bool   `json:"is_specify_time" validate:"omitempty"`
	Priority      *int32  `json:"priority" validate:"omitempty,oneof=1 2 3"`
	IsComplete    *bool   `json:"is_complete" validate:"omitempty"`
}

func (ins ReqListTask) toConditions() *model.TaskConditions {
	cons := &model.TaskConditions{}

	if ins.TaskId != nil {
		taskId := int(*ins.TaskId)
		cons.ID = &condition.Int{EQ: &taskId}
	}
	if ins.CategoryId != nil {
		categoryId := int(*ins.CategoryId)
		cons.CategoryId = &condition.Int{EQ: &categoryId}
	}
	if ins.Title != nil {
		taskTitle := *ins.Title
		cons.Title = &condition.String{StartAt: &taskTitle}
	}
	if ins.IsSpecifyTime != nil {
		isSpecifyTime := *ins.IsSpecifyTime
		cons.IsSpecifyTime = &condition.Bool{EQ: &isSpecifyTime}
	}
	if ins.Priority != nil {
		priority := int(*ins.Priority)
		cons.Priority = &condition.Int{EQ: &priority}
	}
	if ins.IsComplete != nil {
		isComplete := *ins.IsComplete
		cons.IsComplete = &condition.Bool{EQ: &isComplete}
	}

	return cons
}

func (s *Server) ListTask(ctx context.Context, req *pb.ListTaskRequest) (*pb.ListResponse, error) {
	claims, err := middleware.GetClaimsFromContext(ctx)
	if err != nil {
		log.Error.Printf("Failed to get user ID: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
	}
	userId := claims.UserID

	reqList := &ReqListTask{}
	if err := bindRequest(req, reqList); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	limit := int(reqList.PageSize)
	offset := int((reqList.Page - 1) * reqList.PageSize)
	cons := reqList.toConditions()
	cons.UserId = &condition.Int{EQ: &userId}
	reqOrderBy := &model.TaskOrderBy{}
	if reqList.SortBy != nil {
		reqOrderBy.Parse(ParseSortBy(*reqList.SortBy))
	}

	conn := db.GetConn()
	listTask := model.ListTask(conn, cons, reqOrderBy, &limit, &offset)
	count, err := model.GetTaskCount(conn, cons)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get task count: %v", err)
	}

	pbTasks := []*pb.Task{}
	for _, task := range listTask {
		pbTasks = append(pbTasks, &pb.Task{
			Id:              int32(task.ID),
			UserId:          int32(task.UserId),
			CategoryId:      int32(task.CategoryId),
			Title:           task.Title,
			Note:            task.Note,
			Url:             task.Url,
			SpecifyDatetime: util.GetFullDateStrFromPtr(&task.SpecifyDatetime),
			IsSpecifyTime:   task.IsSpecifyTime,
			Priority:        int32(task.Priority),
			IsComplete:      task.IsComplete,
			CreatedAt:       util.GetFullDateStr(task.CreatedAt),
			UpdatedAt:       util.GetFullDateStr(task.UpdatedAt),
		})
	}

	return &pb.ListResponse{
		Data: &pb.ListResponse_Tasks{
			Tasks: &pb.Tasks{
				Data: pbTasks,
			},
		},
		TotalCount: count,
		Page:       reqList.Page,
		PageSize:   reqList.PageSize,
		Status:     http.StatusOK,
		Message:    "ok",
	}, nil
}

type ReqUpdateTask struct {
	Id              int32   `json:"id" validate:"required,min=1"`
	CategoryId      *int32  `json:"category_id" validate:"omitempty,min=1"`
	Title           *string `json:"title" validate:"omitempty,max=100"`
	Note            *string `json:"note" validate:"omitempty,max=255"`
	Url             *string `json:"url" validate:"omitempty,max=255"`
	SpecifyDatetime *int64  `json:"specify_datetime" validate:"omitempty,min=1"`
	Priority        *int32  `json:"priority" validate:"omitempty,oneof=1 2 3"`
	IsComplete      *bool   `json:"is_complete" validate:"omitempty"`
}

func (ins ReqUpdateTask) toFieldValues() (model.TaskFieldValues, bool) {
	requiredCheck := false
	fv := model.TaskFieldValues{}
	fv.ID = model.GiveColInt(int(ins.Id))

	if ins.CategoryId != nil {
		requiredCheck = true
		fv.CategoryId = model.GiveColInt(int(*ins.CategoryId))
	}
	if ins.Title != nil {
		requiredCheck = true
		fv.Title = model.GiveColString(*ins.Title)
	}
	if ins.Note != nil {
		requiredCheck = true
		fv.Note = model.GiveColString(*ins.Note)
	}
	if ins.Url != nil {
		requiredCheck = true
		fv.Url = model.GiveColString(*ins.Url)
	}
	var t *time.Time
	isSpecifyTime := false
	if ins.SpecifyDatetime != nil {
		requiredCheck = true
		t = Pointer(time.Unix(*ins.SpecifyDatetime/1000, 0))
		isSpecifyTime = true
	}
	fv.SpecifyDatetime = model.GiveColNullTime(t)
	fv.IsSpecifyTime = model.GiveColBool(isSpecifyTime)
	if ins.Priority != nil {
		requiredCheck = true
		fv.Priority = model.GiveColInt(int(*ins.Priority))
	}
	if ins.IsComplete != nil {
		requiredCheck = true
		fv.IsComplete = model.GiveColBool(*ins.IsComplete)
	}

	return fv, requiredCheck
}

func (s *Server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Response, error) {
	conn := db.GetConn()

	// Validate request
	reqUpdate := &ReqUpdateTask{}
	if err := bindRequest(req, reqUpdate); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	taskId := int(reqUpdate.Id)
	getTask := model.GetTaskByID(conn, taskId)
	if getTask == nil {
		return nil, status.Errorf(codes.NotFound, "task ID not found")
	}

	insFields, insCheck := reqUpdate.toFieldValues()
	if insCheck {
		tx, txErr := conn.Begin()
		if txErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to open db transaction: %v", txErr)
		}
		defer tx.Rollback()

		if err := model.UpdateTask(tx, taskId, &insFields); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
		}

		getTask := model.GetTaskByID(tx, taskId)
		if getTask == nil {
			return nil, status.Errorf(codes.NotFound, "task ID not found")
		}

		comErr := tx.Commit()
		if comErr != nil {
			log.Error.Printf("failed to create task from db tx: %v", comErr)
			return nil, status.Errorf(codes.Internal, "failed to create task from db tx: %v", comErr)
		}

		return &pb.Response{
			Data: &pb.Response_Task{
				Task: &pb.Task{
					Id:              int32(getTask.ID),
					UserId:          int32(getTask.UserId),
					CategoryId:      int32(getTask.CategoryId),
					Title:           getTask.Title,
					Note:            getTask.Note,
					Url:             getTask.Url,
					SpecifyDatetime: util.GetFullDateStrFromPtr(&getTask.SpecifyDatetime),
					IsSpecifyTime:   getTask.IsSpecifyTime,
					Priority:        int32(getTask.Priority),
					IsComplete:      getTask.IsComplete,
					CreatedAt:       util.GetFullDateStr(getTask.CreatedAt),
					UpdatedAt:       util.GetFullDateStr(getTask.UpdatedAt),
				},
			},
			Status:  http.StatusOK,
			Message: "ok",
		}, nil
	}

	return &pb.Response{
		Data:    nil,
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}

func (s *Server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.Response, error) {
	conn := db.GetConn()

	// Validate request
	reqDelete := &ReqId{}
	if err := bindRequest(req, reqDelete); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	taskId := int(reqDelete.Id)
	getTask := model.GetTaskByID(conn, taskId)
	if getTask == nil {
		return nil, status.Errorf(codes.NotFound, "task ID not found")
	}

	tx, txErr := conn.Begin()
	if txErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to open db transaction: %v", txErr)
	}
	defer tx.Rollback()

	if err := model.DeleteTask(tx, taskId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}

	comErr := tx.Commit()
	if comErr != nil {
		log.Error.Printf("failed to create task from db tx: %v", comErr)
		return nil, status.Errorf(codes.Internal, "failed to create task from db tx: %v", comErr)
	}

	return &pb.Response{
		Data:    nil,
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}
