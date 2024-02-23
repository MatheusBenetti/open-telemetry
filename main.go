package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/viper"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initProvider(serviceName, collectorURL string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, collectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func init() {
	viper.AutomaticEnv()
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdownServiceA, err := initProvider("ServiceA", viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT_A"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownServiceA(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider for ServiceA: %w", err)
		}
	}()

	shutdownServiceB, err := initProvider("ServiceB", viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT_B"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownServiceB(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider for ServiceB: %w", err)
		}
	}()

	routerServiceA := http.NewServeMux()
	routerServiceA.HandleFunc("/serviceA", func(w http.ResponseWriter, r *http.Request) {
		// Your logic for Service A
	})

	routerServiceB := http.NewServeMux()
	routerServiceB.HandleFunc("/serviceB", func(w http.ResponseWriter, r *http.Request) {
		// Your logic for Service B
	})

	go func() {
		log.Println("Starting Service A on port", viper.GetString("HTTP_PORT_A"))
		if err := http.ListenAndServe(viper.GetString("HTTP_PORT_A"), routerServiceA); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Println("Starting Service B on port", viper.GetString("HTTP_PORT_B"))
		if err := http.ListenAndServe(viper.GetString("HTTP_PORT_B"), routerServiceB); err != nil {
			log.Fatal(err)
		}
	}()

	select {
	case <-sigCh:
		log.Println("Shutting down gracefully, CTRL+C pressed...")
	case <-ctx.Done():
		log.Println("Shutting down due to other reason...")
	}

	// Create a timeout context for the graceful shutdown
	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
}
