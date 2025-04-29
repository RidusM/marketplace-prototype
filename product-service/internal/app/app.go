package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/repository"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/service"
	grpcapp "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/transport/grpc"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/logger"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/storage/postgres"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/storage/redis"
)

func Run(ctx context.Context, configPath string) {
	// Logger

	// Load config
	cfg, err := config.New(configPath)
	if err != nil {
		panic(err)
	}

	l := logger.New("prod", cfg.Kafka.Brokers)

	// Display config
	log.Println(cfg)

	// Postgres Database
	db, err := postgres.New(&cfg.Postgres, postgres.MaxPoolSize(int32(cfg.Postgres.MaxPool)), postgres.ConnTimeout(cfg.Postgres.ConnTimeout), postgres.MaxConnAttempts(cfg.Postgres.ConnAttempts))
	if err != nil {
		l.Error("creating storage:", logger.Err(err))
	}
	defer db.Close()

	// Redis Database
	rdb, err := redis.New(&cfg.Redis, redis.PoolSize(cfg.Postgres.MaxPool), redis.PoolTimeout(cfg.Redis.ConnTimeout), redis.MinIdleCons(cfg.Redis.MinIdleCons))
	if err != nil {
		l.Error("creating cache:", logger.Err(err))
	}

	// Repository
	repo := repository.NewPostgresRepository(db)

	// Cache
	cache := repository.NewRedisCache(rdb)

	// Service
	svc := service.NewProductService(repo, cache)

	// GRPC Server
	serv, err := grpcapp.New(ctx, ":8080", ":8081", cfg.OTLP, l, svc)
	l.Info("Starting server...")
	go func() {
		if err = serv.Start(context.Background()); err != nil {
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	serv.Stop(ctx)

	log.Print("Server Stopped")
}
