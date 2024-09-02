package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/newmohr/example/api"
	"github.com/newmohr/example/internal/domain/repository"
	grpcRepo "github.com/newmohr/example/internal/infra/grpc"
	"github.com/newmohr/example/internal/infra/local/memori"
	"github.com/newmohr/example/internal/infra/telemetry"
	"github.com/newmohr/example/internal/presentation/gql"
	"github.com/newmohr/example/internal/presentation/gql/middleware"
	"github.com/newmohr/example/internal/presentation/scheduler"
	"github.com/newmohr/example/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracer := initTracer(ctx)
	defer shutdownTracer(ctx)

	cacheUpdateInterval := parseFlags()

	conn, geoGrpcClient := initGRPCClient(ctx)
	defer conn.Close()

	locationCache := memori.NewLocationCache()
	locationUseCase := usecase.NewLocationUseCase(geoGrpcClient, locationCache)

	jobController := scheduler.NewLocationScheduler(locationUseCase)
	resolver := gql.NewResolver(locationUseCase)

	srv := handler.NewDefaultServer(api.NewExecutableSchema(api.Config{Resolvers: resolver}))
	srv.Use(middleware.NewLoggingMiddleware())

	jobController.StartCacheUpdater(ctx, cacheUpdateInterval)

	startHTTPServer(srv)

	waitForShutdown()
}

func initTracer(ctx context.Context) func(context.Context) error {
	shutdownTracer, err := telemetry.InitTracer(ctx)
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	return shutdownTracer
}

func parseFlags() time.Duration {
	defaultInterval := getEnvAsInt("DEFAULT_INTERVAL", 30)
	updateInterval := flag.Int("update-interval", defaultInterval, "Cache update interval in seconds")
	flag.Parse()
	return time.Duration(*updateInterval) * time.Second
}

func initGRPCClient(ctx context.Context) (*grpc.ClientConn, repository.LocationRepository) {
	grpcHost := getEnv("GRPC_SERVER_HOST", "grpc-server")
	grpcPort := getEnv("GRPC_SERVER_PORT", "50051")
	grpcServerAddress := grpcHost + ":" + grpcPort
	conn, err := grpc.DialContext(ctx, grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to create gRPC client: %v", err)
	}
	geoGrpcClient := grpcRepo.NewGeoGrpc(conn)
	return conn, geoGrpcClient
}

func startHTTPServer(srv *handler.Server) {
	httpPort := getEnv("HTTP_PORT", "8080")
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", httpPort)
		log.Fatal(http.ListenAndServe(":"+httpPort, nil))
	}()
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return fallback
}
