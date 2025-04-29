package main

import (
	"context"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/app"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := logger.New(cfg.Env, cfg.Kafka.Brokers)

	app.Run(ctx, log, cfg)
}
