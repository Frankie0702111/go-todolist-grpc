package middleware

import (
	"context"
	"errors"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/pkg/log"
	"go-todolist-grpc/internal/pkg/util"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// List of methods that require authentication
var authRequiredMethods = map[string]bool{
	"/pb.ToDoList/UpdateUser":     true,
	"/pb.ToDoList/CreateCategory": true,
	"/pb.ToDoList/GetCategory":    true,
	"/pb.ToDoList/ListCategory":   true,
	"/pb.ToDoList/UpdateCategory": true,
	"/pb.ToDoList/DeleteCategory": true,
}

func Verify(cnf *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if the method requires authentication
		if !authRequiredMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		tokens := md["authorization"]
		if len(tokens) == 0 {
			return nil, errors.New("missing token")
		}

		token := tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")

		// Validate token
		claims, err := util.ParseToken(cnf.JwtSecretKey, token)
		log.Info.Printf("claims = %v\n", claims)
		if err != nil {
			log.Error.Printf("invalid token: %v", err)
			return nil, errors.New("invalid token")
		}

		// Proceed with the handler
		return handler(ctx, req)
	}
}
