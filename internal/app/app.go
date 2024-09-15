package app

import (
	"log/slog"

	authapp "github.com/Satori27/sso/internal/app/grpc"
	"github.com/Satori27/sso/internal/config"
	"github.com/Satori27/sso/internal/services/auth"
	"github.com/Satori27/sso/internal/storage/postgres"
)

type App struct{
	GRPCServer *authapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App{
	storage, err := postgres.New(cfg)
	if err!=nil{
		panic(err)
	}
	authService := auth.New(log, storage, storage, storage,storage, cfg.TokenTTL)
	authapp:=authapp.New(log, cfg.GRPC.Port, authService)

	return &App{GRPCServer: authapp}
}