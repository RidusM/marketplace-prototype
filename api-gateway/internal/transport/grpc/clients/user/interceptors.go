package userClient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
)

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func CircuitBreakerClientInterceptor(cb *gobreaker.CircuitBreaker[interface{}]) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		const op = "grpc.clients.user.interceptors.CircuitBreakerClientInterceptor"
		_, cbErr := cb.Execute(func() (interface{}, error) {
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			return nil, nil
		})

		return cbErr
	}
}
