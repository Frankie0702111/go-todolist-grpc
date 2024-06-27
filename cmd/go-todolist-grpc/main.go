package main

import (
	"context"
	"errors"
	"go-todolist-grpc/api/pb"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/middleware"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/log"
	"go-todolist-grpc/internal/service"
	logger "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	// Load configuration
	cnfErr := config.Load()
	if cnfErr != nil {
		logger.Fatal(cnfErr)
	}
	cnf := config.Get()

	// Init log
	log.Init(cnf.LogLevel, cnf.LogFolderPath, strconv.Itoa(os.Getpid()), cnf.EnableConsoleOutput, cnf.EnableFileOutput)

	// Init database
	initDBErr := db.Init(&db.Option{
		Host:                     cnf.DBHost,
		Port:                     cnf.DBPort,
		Username:                 cnf.DBUser,
		Password:                 cnf.DBPassword,
		DBName:                   cnf.DBName,
		ConnectionMaxLifeTimeSec: cnf.DBConnectionMaxLifeTimeSec,
		MaxConn:                  cnf.DBMaxConnection,
		MaxIdle:                  cnf.DBMaxIdle,
	})
	if initDBErr != nil {
		log.Error.Printf("initial DB connection error: %v", initDBErr)
	}

	dbConn := db.GetConn()
	defer dbConn.Close()

	// Create a context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()
	waitGroup, ctx := errgroup.WithContext(ctx)

	// Init Http server
	runGatewayServer(cnf, ctx, waitGroup)

	// Init gRPC server
	runGrpcServer(cnf, ctx, waitGroup)

	err := waitGroup.Wait()
	if err != nil {
		logger.Fatalf("error from wait group: %v", err)
	}
}

func runGrpcServer(cnf *config.Config, ctx context.Context, waitGroup *errgroup.Group) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			middleware.Verify(cnf),
		)),
	)
	pb.RegisterToDoListServer(grpcServer, &service.Server{})
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cnf.GprcServerPort)
	if err != nil {
		log.Error.Printf("cannot create listener: %v", err)
	}

	waitGroup.Go(func() error {
		log.Info.Printf("start gRPC server at %s", listener.Addr().String())

		err = grpcServer.Serve(listener)
		if err != nil {
			log.Error.Printf("gRPC server failed to serve: %v", err)
			return err
		}

		return nil
	})

	// Graceful shutdown gRPC
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info.Println("graceful shutdown gRPC server")

		grpcServer.GracefulStop()
		log.Info.Println("gRPC server is stopped")

		return nil
	})
}

func runGatewayServer(cnf *config.Config, ctx context.Context, waitGroup *errgroup.Group) {
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	if err := pb.RegisterToDoListHandlerServer(ctx, grpcMux, &service.Server{}); err != nil {
		log.Error.Printf("cannot register handler server: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	httpServer := &http.Server{
		Handler: grpcMux,
		Addr:    ":" + cnf.HttpServerPort,
	}

	waitGroup.Go(func() error {
		log.Info.Printf("start HTTP gateway server at %s", httpServer.Addr)

		if err := httpServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error.Printf("HTTP gateway server failed to serve: %v", err)

			return err
		}

		return nil
	})

	// Graceful shutdown gateway
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info.Println("graceful shutdown HTTP gateway server")

		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Error.Printf("failed to shutdown HTTP gateway server: %v", err)
			return err
		}

		log.Info.Println("HTTP gateway server is stopped")
		return nil
	})
}
