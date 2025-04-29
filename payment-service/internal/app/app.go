package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ursulgwopp/payment-microservice/internal/config"
	"github.com/ursulgwopp/payment-microservice/internal/repository"
	"github.com/ursulgwopp/payment-microservice/internal/service"
	grpcServer "github.com/ursulgwopp/payment-microservice/internal/transport"

	"github.com/ursulgwopp/payment-microservice/pkg/logger"
)

func Run(ctx context.Context, configPath string) {
	cfg, err := config.New(configPath)
	if err != nil {
		panic(err)
	}

	log := logger.New("local", cfg.Kafka.Brokers)

	repo, err := repository.NewPostgresRepository(&cfg.Postgres)
	if err != nil {
		log.Error(err.Error())
	}

	cache, err := repository.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.Error(err.Error())
	}

	srv := service.New(repo, cache)

	serv, err := grpcServer.New(ctx, ":8080", ":8081", cfg.OTLP, log, srv)
	if err != nil {
		log.Error("user client error", "err", logger.Err(err))
		os.Exit(1)
	}

	go func() {
		if err = serv.Start(context.Background()); err != nil {
			log.Error("server error", "err", logger.Err(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	serv.Stop(ctx)

	log.Info("Server Stopped")
}
