package grpcapp

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/config"
	order "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/api/client"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/logger"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/metric"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/oteltrace"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

var (
	sleep = flag.Duration("sleep", time.Second*5, "duration between changes in health")

	kaep = keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: false,
	}

	kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
	}
)

// Server encapsulates the gRPC and HTTP servers along with related resources
type Server struct {
	grpcServer        *grpc.Server
	httpServer        *http.Server
	grpcAddr          string
	log               *logger.Logger
	traceProvider     trace.TracerProvider
	prometheusFactory metric.PrometheusFactory
}

func New(ctx context.Context, port, metricsPort string, otlpConfig config.OTLPConfig, log *logger.Logger, service Service) (*Server, error) {

	logSpanTraceID := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{
				"traceID", span.TraceID().String(),
				"spanID", span.SpanID().String(),
			}
		}
		return nil
	}

	spanTraceFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{
				"traceID": span.TraceID().String(),
				"spanID":  span.SpanID().String(),
			}
		}
		return nil
	}

	exp, err := oteltrace.NewOTLPExporter(ctx, otlpConfig.OTLPEndpoint)
	if err != nil {
		log.Error("failed to create OTLP exporter", logger.Err(err))
		return nil, err
	}

	tp, err := oteltrace.NewTraceProvider(exp, "user-service")
	if err != nil {
		log.Error("failed to create trace provider", logger.Err(err))
		return nil, err
	}
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	panicsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})

	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		log.Error("recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	cert, err := tls.LoadX509KeyPair("/app/x509/server-cert.pem", "/app/x509/server-key.pem")
	if err != nil {
		log.Error("failed to load key pair: %s", logger.Err(err))
		return nil, err
	}

	grpcSrv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(spanTraceFromContext)),
			logging.UnaryServerInterceptor(InterceptorLogger(log), logging.WithFieldsFromContext(logSpanTraceID)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	)

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(grpcSrv, healthcheck)
	order.RegisterOrderServiceServer(grpcSrv, NewOrderService(service))

	go func() {
		next := healthpb.HealthCheckResponse_SERVING
		for {
			healthcheck.SetServingStatus("auth-service", next)
			if next == healthpb.HealthCheckResponse_SERVING {
				next = healthpb.HealthCheckResponse_NOT_SERVING
			} else {
				next = healthpb.HealthCheckResponse_SERVING
			}
			time.Sleep(*sleep)
		}
	}()

	httpSrv := &http.Server{Addr: metricsPort}

	srvMetrics.InitializeMetrics(grpcSrv)

	return &Server{
		grpcServer:        grpcSrv,
		httpServer:        httpSrv,
		grpcAddr:          port,
		log:               log,
		traceProvider:     tp,
		prometheusFactory: metric.NewPrometheusFactory(srvMetrics, panicsTotal),
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		l, err := net.Listen("tcp", s.grpcAddr)
		if err != nil {
			return err
		}

		s.log.Info("starting gRPC server", "addr", l.Addr().String())

		return s.grpcServer.Serve(l)
	})

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

func (s *Server) Stop(ctx context.Context) error {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
	}
	return nil
}
