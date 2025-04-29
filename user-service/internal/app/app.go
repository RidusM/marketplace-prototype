package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"userService/internal/config"
	"userService/internal/controller/grpcapp"
	"userService/internal/repository"
	"userService/internal/service"
	"userService/pkg/logger"
	"userService/pkg/storage/postgres"
	"userService/pkg/storage/redis"
)

//TODO: connect metrics + kafka + log handling

func Run(configPath string) {
	ctx := context.Background()

	// Load config
	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatal("loading config: ", err.Error())
	}

	// Logger
	log := logger.New(cfg.Env, cfg.KafkaConfig.Brokers)

	// Postgres Database
	db, err := postgres.New(cfg.PostgresConfig, postgres.MaxPoolSize(int32(cfg.PostgresConfig.MaxPool)), postgres.ConnTimeout(cfg.PostgresConfig.ConnTimeout), postgres.MaxConnAttempts(cfg.PostgresConfig.ConnAttempts))
	if err != nil {
		log.Error("creating storage:", logger.Err(err))
	}
	defer db.Close()
	log.Info("Successfully set up storage")

	// Redis Database
	rdb, err := redis.New(cfg.RedisConfig, redis.PoolSize(cfg.RedisConfig.PoolSize), redis.PoolTimeout(cfg.RedisConfig.ConnTimeout), redis.MinIdleCons(cfg.RedisConfig.MinIdleCons))
	if err != nil {
		log.Error("creating cache:", logger.Err(err))
	}
	log.Info("Successfully set up cache")

	// Repository
	userRepo := repository.New(db)
	cacheRepo := repository.NewCacheRepository(rdb)

	// Service
	svc := service.New(userRepo, cacheRepo)

	// GRPC Server
	serv, err := grpcapp.New(ctx, ":8080", ":8081", cfg.OTLPConfig, log, svc)
	if err != nil {
		return
	}

	go func() {
		serv.Start(context.Background())
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	serv.Stop(ctx)

	log.Info("Server Stopped")
}
