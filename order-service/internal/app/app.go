package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/repository"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/service"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/transport/grpcapp"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/logger"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/storage/postgres"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/storage/redis"
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
	log := logger.New(cfg.Env, cfg.Kafka.Brokers)

	// Postgres Database
	db, err := postgres.New(cfg.Postgres, postgres.MaxPoolSize(int32(cfg.Postgres.MaxPool)), postgres.ConnTimeout(cfg.Postgres.ConnTimeout), postgres.MaxConnAttempts(cfg.Postgres.ConnAttempts))
	if err != nil {
		log.Error("creating storage:", logger.Err(err))
	}
	defer db.Close()
	log.Info("Successfully set up storage")

	// Redis Database
	rdb, err := redis.New(cfg.Redis, redis.PoolSize(cfg.Postgres.MaxPool), redis.PoolTimeout(cfg.Redis.ConnTimeout), redis.MinIdleCons(cfg.Redis.MinIdleCons))
	if err != nil {
		log.Error("creating cache:", logger.Err(err))
	}
	log.Info("Successfully set up cache")

	// Repository
	orderRepo := repository.NewOrderRepository(db)
	itemRepo := repository.NewItemRepository(db)

	// Cache
	cache := repository.NewRedisCache(rdb)

	// Service
	svc := service.New(itemRepo, orderRepo, cache)

	// GRPC Server
	serv, err := grpcapp.New(ctx, ":8080", ":8081", cfg.OTLP, log, svc)
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
