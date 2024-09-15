package main

import (
	// "fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Satori27/sso/internal/app"
	"github.com/Satori27/sso/internal/config"
)

const (
	localEnv = "local"
	prodEnv = "production"
)


func main(){
	cfg:=config.MustLoad()
	log:=setupLogger(cfg.Env)

	application:= app.New(log, cfg)

	go application.GRPCServer.MustRun()

	stop:=make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()
}

func setupLogger(env string) *slog.Logger{
	var log *slog.Logger

	switch env{
	case localEnv: log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case prodEnv:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default: panic("invalid value for env key in config : "+env)
	}
	return log
}