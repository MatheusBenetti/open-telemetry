package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	conf "github.com/MatheusBenetti/opentelemetry/config"
	"github.com/MatheusBenetti/opentelemetry/internal/inputHandle/infra/opentel"
	"github.com/MatheusBenetti/opentelemetry/internal/inputHandle/infra/web"
	"go.opentelemetry.io/otel"
)

func main() {
	var cfg conf.Config
	viperCfg := conf.NewViper("env.json")
	viperCfg.ReadViper(&cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	providerShutdown, provErr := opentel.InitProvider(
		"service_a_orchestration",
		cfg.Zipkin.Endpoint,
	)
	if provErr != nil {
		return
	}

	defer func() {
		if err := providerShutdown(ctx); err != nil {
			log.Printf("failed shuting down the tracer provider %s\n", err.Error())
		}
	}()

	server := web.Server{
		TemplateData: web.TemplateData{
			Title:           "Service A: Orchestration",
			ExternalCallURL: cfg.ServiceB.Host,
			RequestNameOtel: "service_a:all",
			OTELTracer:      otel.Tracer("service_a"),
		},
	}

	server.Execute()
}
