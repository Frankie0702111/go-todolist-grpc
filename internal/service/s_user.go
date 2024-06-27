package service

import (
	"context"
	"go-todolist-grpc/api/pb"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/model"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/hash"
	"go-todolist-grpc/internal/pkg/log"
	"go-todolist-grpc/internal/pkg/util"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReqRegister struct {
	Email    string `json:"email" validate:"required,email,max=64"`
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8"`
}

func (ins ReqRegister) toFieldValues() model.UserFieldValues {
	now := time.Now().UTC()
	fv := model.UserFieldValues{}
	fv.Email = model.GiveColString(ins.Email)
	fv.Username = model.GiveColString(ins.Username)
	fv.Password = model.GiveColString(ins.Password)
	fv.Status = model.GiveColBool(true)
	fv.CreatedAt = model.GiveColTime(now)
	fv.UpdatedAt = model.GiveColTime(now)
	return fv
}

func (s *Server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.Response, error) {
	cnf := config.Get()
	conn := db.GetConn()

	// Check if the email is already registered
	getUser := model.GetUserByEmail(conn, req.Email)
	if getUser != nil {
		return nil, status.Errorf(codes.AlreadyExists, "the email already exists")
	}

	// Validate request
	reqRegister := &ReqRegister{}
	if err := bindRequest(req, reqRegister); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	// Hash password
	hashPassword, hashPasswordErr := hash.HashPassword(cnf.BcryptCost, reqRegister.Password)
	if hashPasswordErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", hashPasswordErr)
	}
	reqRegister.Password = hashPassword

	// Create user
	tx, txErr := conn.Begin()
	if txErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to open db transaction: %v", txErr)
	}
	defer tx.Rollback()

	insFields := reqRegister.toFieldValues()
	user, userErr := model.CreateUser(tx, &insFields)
	if userErr != nil {
		if userErr.Error() == "email already exists" {
			return nil, status.Errorf(codes.AlreadyExists, "email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", userErr)
	}

	comErr := tx.Commit()
	if comErr != nil {
		log.Error.Printf("failed to create user from db tx: %v", comErr)
		return nil, status.Errorf(codes.Internal, "failed to create user from db tx: %v", comErr)
	}

	userInfo := &pb.User{
		Id:        int64(user.ID.Val),
		Username:  user.Username.Val,
		Email:     user.Email.Val,
		CreatedAt: util.GetFullDateStr(user.CreatedAt.Val),
		UpdatedAt: util.GetFullDateStr(user.UpdatedAt.Val),
	}

	return &pb.Response{
		Data: &pb.Response_User{
			User: userInfo,
		},
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}

type ReqLogin struct {
	Email    string `validate:"required,email,max=64"`
	Password string `validate:"required,min=8"`
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.Response, error) {
	cnf := config.Get()
	conn := db.GetConn()

	// Check if this email is unregistered
	getUser := model.GetUserByEmail(conn, req.Email)
	if getUser == nil {
		return nil, status.Errorf(codes.NotFound, "this email is unregistered")
	}

	// Validate request
	reqLogin := &ReqLogin{}
	if err := bindRequest(req, reqLogin); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	// Check user password
	if checkPassword := hash.CheckPasswordHash(req.Password, getUser.Password); !checkPassword {
		return nil, status.Errorf(codes.InvalidArgument, "password is incorrect")
	}

	// Grnerate token
	token, tokenErr := util.GenerateToken(cnf.JwtTtl, cnf.JwtSecretKey, int64(getUser.ID))
	if tokenErr != nil {
		log.Error.Printf("failed to generate token: %v", tokenErr)
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", tokenErr)
	}

	return &pb.Response{
		Data: &pb.Response_User{
			User: &pb.User{
				Id:        int64(getUser.ID),
				Username:  getUser.Username,
				Email:     getUser.Email,
				CreatedAt: util.GetFullDateStr(getUser.CreatedAt),
				UpdatedAt: util.GetFullDateStr(getUser.UpdatedAt),
				Token:     &token,
			},
		},
		Status:  http.StatusOK,
		Message: "ok",
	}, nil
}

type ReqUpdate struct {
	UserId   int64   `json:"user_id" validate:"required"`
	Username *string `json:"username" validate:"omitempty,min=3,max=32"`
	Password *string `json:"password" validate:"omitempty,min=8"`
}

func (ins ReqUpdate) toFieldValues() (model.UserFieldValues, bool) {
	requiredCheck := false
	fv := model.UserFieldValues{}
	fv.ID = model.GiveColInt(int(ins.UserId))

	if ins.Username != nil {
		requiredCheck = true
		fv.Username = model.GiveColString(*ins.Username)
	}

	if ins.Password != nil {
		requiredCheck = true
		fv.Password = model.GiveColString(*ins.Password)
	}

	return fv, requiredCheck
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.Response, error) {
	cnf := config.Get()
	conn := db.GetConn()

	// Check if the user ID not found
	userId := int(req.UserId)
	if getUser := model.GetUserByID(conn, &userId); getUser == nil {
		return nil, status.Errorf(codes.NotFound, "user ID not found")
	}

	// Validate request
	reqUpdate := &ReqUpdate{}
	if err := bindRequest(req, reqUpdate); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	// Hash password
	if reqUpdate.Password != nil {
		hashPassword, hashPasswordErr := hash.HashPassword(cnf.BcryptCost, *reqUpdate.Password)
		if hashPasswordErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %v", hashPasswordErr)
		}
		reqUpdate.Password = &hashPassword
	}

	// Update user
	insFields, insCheck := reqUpdate.toFieldValues()
	if insCheck {
		tx, txErr := conn.Begin()
		if txErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to open db transaction: %v", txErr)
		}
		defer tx.Rollback()

		if err := model.UpdateUser(tx, userId, &insFields); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
		}

		getUser := model.GetUserByID(tx, &userId)
		if getUser == nil {
			return nil, status.Errorf(codes.NotFound, "user ID not found")
		}

		comErr := tx.Commit()
		if comErr != nil {
			log.Error.Printf("failed to create user from db tx: %v", comErr)
			return nil, status.Errorf(codes.Internal, "failed to create user from db tx: %v", comErr)
		}

		return &pb.Response{
			Data: &pb.Response_User{
				User: &pb.User{
					Id:        int64(getUser.ID),
					Username:  getUser.Username,
					Email:     getUser.Email,
					CreatedAt: util.GetFullDateStr(getUser.CreatedAt),
					UpdatedAt: util.GetFullDateStr(getUser.UpdatedAt),
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
