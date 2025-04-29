package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/repository"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/service"
	grpcServer "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/transport/grpc"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/auth"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/email"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/logger"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/storage/postgres"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/storage/redis"
)

func Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	const op = "app.Run"

	db, err := postgres.New(&cfg.Postgres, postgres.MaxPoolSize(10))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rdb, err := redis.New(&cfg.Redis)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userRepo := repository.NewUserRepository(db)
	passHistoryRepo := repository.NewPasswordHistoryRepository(db)
	cacheRepo := repository.NewCacheRepository(rdb)

	mailer := email.New(&cfg.Email)

	tokens, err := auth.New(cfg.Auth.SecretKey)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	srv := service.NewAuthService(userRepo, cacheRepo, passHistoryRepo, tokens, mailer, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL, cfg.Auth.UserCacheTTL, cfg.Auth.VerifyTTL, log, &sync.WaitGroup{})

	serv, err := grpcServer.New(ctx, cfg.App.Port, cfg.Metrics.MetricsPort, cfg.OTLP, log, srv)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		serv.Start(context.Background())
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	serv.Stop(ctx)

	log.Info("Server Stopped")

	return nil
}
