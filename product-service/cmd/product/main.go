package main

import (
	"context"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/app"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	app.Run(ctx, "./configs/local.yaml")
}
