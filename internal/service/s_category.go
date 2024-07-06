package service

import (
	"context"
	"go-todolist-grpc/api/pb"
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

type ReqCreateCategory struct {
	Name string `json:"name" validate:"required,min=1,max=128"`
}

func (ins ReqCreateCategory) toFieldValues() model.CategoryFieldValues {
	now := time.Now().UTC()
	fv := model.CategoryFieldValues{}
	fv.Name = model.GiveColString(ins.Name)
	fv.CreatedAt = model.GiveColTime(now)
	fv.UpdatedAt = model.GiveColTime(now)
	return fv
}

func (s *Server) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Response, error) {
	conn := db.GetConn()

	// Validate request
	reqCategory := &ReqCreateCategory{}
	if err := bindRequest(req, reqCategory); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	// Check if the category name is already exists
	getCategory := model.GetCategoryByName(conn, reqCategory.Name)
	if getCategory != nil {
		return nil, status.Errorf(codes.AlreadyExists, "the category already exists")
	}

	insFields := reqCategory.toFieldValues()
	category, categoryErr := model.CreateCategory(conn, &insFields)
	if categoryErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", categoryErr)
	}

	categoryInfo := &pb.Category{
		Id:        int32(category.ID.Val),
		Name:      category.Name.Val,
		CreatedAt: util.GetFullDateStr(category.CreatedAt.Val),
		UpdatedAt: util.GetFullDateStr(category.UpdatedAt.Val),
	}

	return &pb.Response{
		Data: &pb.Response_Category{
			Category: categoryInfo,
		},
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}

func (s *Server) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.Response, error) {
	conn := db.GetConn()

	// Validate request
	reqGet := &ReqId{}
	if err := bindRequest(req, reqGet); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	categoryId := int(reqGet.Id)
	getCategory := model.GetCategoryByID(conn, categoryId)
	if getCategory == nil {
		return nil, status.Errorf(codes.NotFound, "category ID not found")
	}

	return &pb.Response{
		Data: &pb.Response_Category{
			Category: &pb.Category{
				Id:        int32(getCategory.ID),
				Name:      getCategory.Name,
				CreatedAt: util.GetFullDateStr(getCategory.CreatedAt),
				UpdatedAt: util.GetFullDateStr(getCategory.UpdatedAt),
			},
		},
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}

type ReqListCategory struct {
	Page       int32   `json:"page" validate:"required,min=1,max=100000"`
	PageSize   int32   `json:"page_size" validate:"required,min=5,max=1000"`
	SortBy     *string `json:"sort_by" validate:"omitempty,max=10"`
	CategoryId *int32  `json:"category_id" validate:"omitempty,min=1"`
	Name       *string `json:"name" validate:"omitempty,min=1,max=128"`
}

func (ins ReqListCategory) toConditions() *model.CategoryConditions {
	cons := &model.CategoryConditions{}

	if ins.CategoryId != nil {
		categoryId := int(*ins.CategoryId)
		cons.ID = &condition.Int{EQ: &categoryId}
	}

	if ins.Name != nil {
		categoryName := *ins.Name
		cons.Name = &condition.String{EQ: &categoryName}
	}

	return cons
}

func (s *Server) ListCategory(ctx context.Context, req *pb.ListCategoryRequest) (*pb.ListResponse, error) {
	conn := db.GetConn()

	reqList := &ReqListCategory{}
	if err := bindRequest(req, reqList); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	limit := int(reqList.PageSize)
	offset := int((reqList.Page - 1) * reqList.PageSize)
	cons := reqList.toConditions()
	reqOrderBy := &model.CategoryOrderBy{}
	if reqList.SortBy != nil {
		reqOrderBy.Parse(ParseSortBy(*reqList.SortBy))
	}

	listCategory := model.ListCategory(conn, cons, reqOrderBy, &limit, &offset)
	count, err := model.GetCategoryCount(conn, cons)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get category count: %v", err)
	}

	pbCategories := []*pb.Category{}
	for _, category := range listCategory {
		pbCategories = append(pbCategories, &pb.Category{
			Id:        int32(category.ID),
			Name:      category.Name,
			CreatedAt: util.GetFullDateStr(category.CreatedAt),
			UpdatedAt: util.GetFullDateStr(category.UpdatedAt),
		})
	}

	return &pb.ListResponse{
		Data: &pb.ListResponse_Categories{
			Categories: &pb.Categories{
				Data: pbCategories,
			},
		},
		TotalCount: count,
		Page:       reqList.Page,
		PageSize:   reqList.PageSize,
		Status:     http.StatusOK,
		Message:    "ok",
	}, nil
}

type ReqUpdateCategory struct {
	Id   int32   `json:"id" validate:"required,min=1"`
	Name *string `json:"name" validate:"omitempty,min=1,max=128"`
}

func (ins ReqUpdateCategory) toFieldValues() (model.CategoryFieldValues, bool) {
	requiredCheck := false
	fv := model.CategoryFieldValues{}
	fv.ID = model.GiveColInt(int(ins.Id))

	if ins.Name != nil {
		requiredCheck = true
		fv.Name = model.GiveColString(*ins.Name)
	}

	return fv, requiredCheck
}

func (s *Server) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Response, error) {
	conn := db.GetConn()

	// Validate request
	reqUpdate := &ReqUpdateCategory{}
	if err := bindRequest(req, reqUpdate); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	categoryId := int(reqUpdate.Id)
	if getCategory := model.GetCategoryByID(conn, categoryId); getCategory == nil {
		return nil, status.Errorf(codes.NotFound, "category ID not found")
	}

	insFields, insCheck := reqUpdate.toFieldValues()
	if insCheck {
		tx, txErr := conn.Begin()
		if txErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to open db transaction: %v", txErr)
		}
		defer tx.Rollback()

		if err := model.UpdateCategory(tx, categoryId, &insFields); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
		}

		getCategory := model.GetCategoryByID(tx, categoryId)
		if getCategory == nil {
			return nil, status.Errorf(codes.NotFound, "category ID not found")
		}

		comErr := tx.Commit()
		if comErr != nil {
			log.Error.Printf("failed to create user from db tx: %v", comErr)
			return nil, status.Errorf(codes.Internal, "failed to create user from db tx: %v", comErr)
		}

		return &pb.Response{
			Data: &pb.Response_Category{
				Category: &pb.Category{
					Id:        int32(getCategory.ID),
					Name:      getCategory.Name,
					CreatedAt: util.GetFullDateStr(getCategory.CreatedAt),
					UpdatedAt: util.GetFullDateStr(getCategory.UpdatedAt),
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

func (s *Server) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.Response, error) {
	conn := db.GetConn()

	// Validate request
	reqDelete := &ReqId{}
	if err := bindRequest(req, reqDelete); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	categoryId := int(reqDelete.Id)
	if getCategory := model.GetCategoryByID(conn, categoryId); getCategory == nil {
		return nil, status.Errorf(codes.NotFound, "category ID not found")
	}

	tx, txErr := conn.Begin()
	if txErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to open db transaction: %v", txErr)
	}
	defer tx.Rollback()

	if err := model.DeleteCategory(tx, categoryId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}

	comErr := tx.Commit()
	if comErr != nil {
		log.Error.Printf("failed to create user from db tx: %v", comErr)
		return nil, status.Errorf(codes.Internal, "failed to create user from db tx: %v", comErr)
	}

	return &pb.Response{
		Data:    nil,
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}
