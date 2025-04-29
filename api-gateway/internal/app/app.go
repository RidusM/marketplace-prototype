package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/service"
	orderClient "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/grpc/clients/order"
	paymentClient "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/grpc/clients/payment"
	productClient "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/grpc/clients/product"
	authClient "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/grpc/clients/rbacAuth"
	userClient "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/grpc/clients/user"
	httpServ "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/transport/http"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/client"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/order"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/payment"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/product"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/rbacAuth"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/logger"
)

func Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	mux := runtime.NewServeMux()

	paymentCl, err := paymentClient.New(ctx, cfg.PaymentService.Host, cfg.PaymentService.MetricsPort, cfg.OTLP, log)
	if err != nil {
		return fmt.Errorf("failed to create payment client: %w", err)
	}
	err = payment.RegisterPaymentServiceHandlerClient(ctx, mux, paymentCl.PaymentServiceClient)
	if err != nil {
		fmt.Errorf("%v:%w", "payment.RegisterPaymentServiceHandlerClient", err)
	}

	authCl, err := authClient.New(ctx, cfg.AuthService.Host, cfg.AuthService.MetricsPort, cfg.OTLP, log)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}
	err = rbacAuth.RegisterAuthServiceHandlerClient(ctx, mux, authCl.AuthServiceClient)
	if err != nil {
		fmt.Errorf("%v:%w", "payment.RegisterPaymentServiceHandlerClient", err)
	}
	userCl, err := userClient.New(ctx, cfg.UserService.Host, cfg.UserService.MetricsPort, cfg.OTLP, log)
	if err != nil {
		return fmt.Errorf("failed to create user client: %w", err)
	}
	err = client.RegisterUserServiceHandlerClient(ctx, mux, userCl.UserServiceClient)
	if err != nil {
		return fmt.Errorf("failed to create user client: %w", err)
	}
	if err != nil {
		fmt.Errorf("%v:%w", "payment.RegisterPaymentServiceHandlerClient", err)
	}
	productCl, err := productClient.New(ctx, cfg.ProductService.Host, cfg.ProductService.MetricsPort, cfg.OTLP, log)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}
	err = product.RegisterProductServiceHandlerClient(ctx, mux, productCl.ProductServiceClient)
	if err != nil {
		fmt.Errorf("%v:%w", "payment.RegisterPaymentServiceHandlerClient", err)
	}

	orderCl, err := orderClient.New(ctx, cfg.OrderService.Host, cfg.OrderService.MetricsPort, cfg.OTLP, log)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}
	err = order.RegisterOrderServiceHandlerClient(ctx, mux, orderCl.OrderServiceClient)
	if err != nil {
		fmt.Errorf("%v:%w", "order.RegisterOrderServiceHandlerClient", err)
	}

	authService := authClient.NewAuthService(authCl.AuthServiceClient)
	userService := userClient.NewUserService(userCl.UserServiceClient)

	aggregatorService := service.NewAggregatorService(authService, userService, log)

	// Создаем HTTP обработчик
	aggregatorHandler := httpServ.NewAggregatorHandler(aggregatorService)

	// Создаем маршрутизатор Gorilla Mux
	//router := httpServ.NewRouter(aggregatorHandler)

	// Объединяем Gorilla Mux с gRPC Gateway ServeMux
	mainMux := httpServ.NewRouter(aggregatorHandler)
	mainMux.Handle("/", mux) // gRPC Gateway

	// Запускаем HTTP сервер
	log.Info("Starting API Gateway on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}
