package main

import (
	"context"

	"github.com/ursulgwopp/payment-microservice/internal/app"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	app.Run(ctx, "./configs/docker.yaml")
}
