package authapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/Satori27/sso/internal/grpc/auth"
	"google.golang.org/grpc"
)


type App struct{
	log *slog.Logger
	gRPCServer *grpc.Server
	port int
}


func (a *App) MustRun(){
	if err:=a.Run();err!=nil{
		panic(err)
	}
}


func New(log *slog.Logger, port int, auth authgrpc.Auth) *App{
	gRPCServer :=grpc.NewServer()

	authgrpc.Register(gRPCServer, auth)

	return &App{log: log, gRPCServer: gRPCServer, port: port}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log:=a.log.With(slog.String("op", op))

	log.Info("starting gRPC server")

	l, err:=net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err!=nil{
		return fmt.Errorf("%s: %w", op, err)
	}

	err=a.gRPCServer.Serve(l)
	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))
	if err!=nil{
		return fmt.Errorf("%s: %w", op, err)
	}
	
	
	return nil
}


func (a *App) Stop(){
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}