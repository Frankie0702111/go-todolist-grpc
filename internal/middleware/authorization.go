package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/pkg/log"
	"go-todolist-grpc/internal/pkg/util"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// List of methods that require authentication
var authRequiredMethods = map[string]bool{
	// gRPC
	"/pb.ToDoList/UpdateUser":     true,
	"/pb.ToDoList/CreateCategory": true,
	"/pb.ToDoList/GetCategory":    true,
	"/pb.ToDoList/ListCategory":   true,
	"/pb.ToDoList/UpdateCategory": true,
	"/pb.ToDoList/DeleteCategory": true,

	// gateway
	"/v1/user/update":     true,
	"/v1/category/create": true,
	"/v1/category/get":    true,
	"/v1/category/list":   true,
	"/v1/category/update": true,
	"/v1/category/delete": true,
}

func VerifyTokenByGrpc(cnf *config.Config) grpc.UnaryServerInterceptor {
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

		token := strings.TrimPrefix(tokens[0], "Bearer ")

		// Validate token
		claims, err := util.ParseToken(cnf.JwtSecretKey, token)
		if err != nil {
			log.Error.Printf("invalid token: %v", err)
			return nil, errors.New("invalid token")
		}

		claimsJSON, _ := json.Marshal(claims)
		newMD := metadata.New(map[string]string{
			"x-auth-claims": string(claimsJSON),
		})
		// Merge the new metadata with the existing one
		newCtx := metadata.NewIncomingContext(ctx, metadata.Join(md, newMD))

		// Proceed with the handler
		return handler(newCtx, req)
	}
}

func VerifyTokenByGateway(cnf *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the method requires authentication
			if !authRequiredMethods[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			claims, err := util.ParseToken(cnf.JwtSecretKey, token)
			if err != nil {
				log.Error.Printf("invalid token: %v", err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Convert claims to JSON string
			claimsJSON, _ := json.Marshal(claims)
			// Add claims to the request header
			r.Header.Set("X-Auth-Claims", string(claimsJSON))

			next.ServeHTTP(w, r)
		})
	}
}

func GetClaimsFromContext(ctx context.Context) (map[string]interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata found in context")
	}

	claimsJSON := md.Get("x-auth-claims")
	if len(claimsJSON) == 0 {
		return nil, errors.New("no claims found in metadata")
	}

	var claims map[string]interface{}
	if err := json.Unmarshal([]byte(claimsJSON[0]), &claims); err != nil {
		return nil, errors.New("failed to parse claims")
	}

	return claims, nil
}
