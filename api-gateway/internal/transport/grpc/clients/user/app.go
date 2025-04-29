package userClient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker/v2"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/api/client"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/logger"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/metric"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/api-gateway/pkg/oteltrace"
	"golang.org/x/sync/errgroup"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/keepalive"
)

var (
	serviceConfig = `{
		"healthCheckConfig": {
			"serviceName": ""
		}
	}`
	kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}
)

type Client struct {
	UserServiceClient client.UserServiceClient
	httpServer        *http.Server
	log               *logger.Logger
	traceProvider     trace.TracerProvider
	prometheusFactory metric.PrometheusFactory
}

func New(ctx context.Context, target, metricsPort string, otlpConfig config.OTLPConfig, log *logger.Logger) (*Client, error) {

	rpcLogger := log.With(
		"service",
		"gRPC/server",
		"component",
		"user-client")

	logSpanTraceID := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{
				"traceID", span.TraceID().String(),
				"spanID", span.SpanID().String(),
			}
		}
		return nil
	}

	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{
				"traceID": span.TraceID().String(),
				"spanID":  span.SpanID().String(),
			}
		}
		return nil
	}

	cb := gobreaker.NewCircuitBreaker[interface{}](gobreaker.Settings{
		Name:        "demo",
		MaxRequests: 25,
		Timeout:     10000,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= 0.6
		},

		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Info("Circuit Breaker: %s, changed from %v, to %v", name, from, to)
		},
	})

	exp, err := oteltrace.NewOTLPExporter(ctx, otlpConfig.Endpoint)
	if err != nil {
		log.Error("failed to create OTLP exporter: %v", logger.Err(err))
	}

	tp, err := oteltrace.NewTraceProvider(exp, "user-client")
	if err != nil {
		log.Error("failed to create trace provider: %v", logger.Err(err))
	}

	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	clMetrics := grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	creds, err := credentials.NewClientTLSFromFile("/app/x509/ca-cert.pem", "/app/x509/server-key.pem")
	if err != nil {
		log.Error("failed to load key pair: %s", logger.Err(err))
	}

	grpcCl, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(creds),
		grpc.WithStatsHandler(otelgrpc.NewServerHandler()),
		grpc.WithChainUnaryInterceptor(
			CircuitBreakerClientInterceptor(cb),
			timeout.UnaryClientInterceptor(500*time.Millisecond),
			clMetrics.UnaryClientInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.UnaryClientInterceptor(InterceptorLogger(rpcLogger), logging.WithFieldsFromContext(logSpanTraceID)),
		),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		panic(err)
	}

	defer grpcCl.Close()

	cl := client.NewUserServiceClient(grpcCl)

	httpSrv := &http.Server{Addr: metricsPort}

	return &Client{
		UserServiceClient: cl,
		httpServer:        httpSrv,
		log:               log,
		traceProvider:     tp,
		prometheusFactory: metric.NewPrometheusFactory(clMetrics),
	}, nil
}

func (s *Client) Start(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", s.prometheusFactory.InitHandler())
		s.httpServer.Handler = mux
		s.log.Info("starting HTTP server", "addr", s.httpServer.Addr)
		return s.httpServer.ListenAndServe()
	})

	eg.Go(func() error {
		<-ctx.Done()
		return s.Stop(ctx)
	})

	return eg.Wait()
}

func (s *Client) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
	}
	return nil
}
