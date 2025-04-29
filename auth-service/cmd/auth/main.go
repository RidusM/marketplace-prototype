package main

import (
	"context"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/app"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/logger"
	_ "go.uber.org/automaxprocs"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := logger.New(cfg.Env, cfg.Kafka.Brokers)

	if err := app.Run(ctx, log, cfg); err != nil {
		panic(err)
	}
}
