package service

import (
	"context"
	"go-todolist-grpc/api/pb"
	"go-todolist-grpc/internal/model"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReqVerifyEmail struct {
	Id         int32  `json:"id" validate:"required,min=1"`
	SecretCode string `json:"secret_code" validate:"required,max=32"`
}

func (ins ReqVerifyEmail) toFieldValues() model.VerifyEmailFieldValues {
	fv := model.VerifyEmailFieldValues{}
	fv.ID = model.GiveColInt(int(ins.Id))
	fv.SecretCode = model.GiveColString(ins.SecretCode)
	fv.IsUsed = model.GiveColBool(true)

	return fv
}

func (s *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.Response, error) {
	conn := db.GetConn()
	now := time.Now().UTC()

	// Validate request
	reqVerifyEmail := &ReqVerifyEmail{}
	if err := bindRequest(req, reqVerifyEmail); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate: %v", err.Error())
	}

	verifyEmailId := int(reqVerifyEmail.Id)
	getVerifyEmail := model.GetVerifyEmailByID(conn, verifyEmailId, false, &now)
	if getVerifyEmail == nil {
		return nil, status.Errorf(codes.NotFound, "verify email ID not found")
	}

	insFields := reqVerifyEmail.toFieldValues()
	tx, txErr := conn.Begin()
	if txErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to open db transaction: %v", txErr)
	}
	defer tx.Rollback()

	if err := model.UpdateVerifyEmail(tx, verifyEmailId, &insFields); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update verify email: %v", err)
	}

	if err := model.UpdateUser(tx, getVerifyEmail.UserId, &model.UserFieldValues{
		IsEmailVerified: model.GiveColBool(true),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	comErr := tx.Commit()
	if comErr != nil {
		log.Error.Printf("failed to verify email from db tx: %v", comErr)
		return nil, status.Errorf(codes.Internal, "failed to verify email from db tx: %v", comErr)
	}

	return &pb.Response{
		Data: &pb.Response_VerifyEmail{
			VerifyEmail: &pb.VerifyEmail{
				IsUsed: true,
			},
		},
	}, nil
}
