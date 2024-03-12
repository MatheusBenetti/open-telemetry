package web

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type TemplateData struct {
	Title              string
	ResponseTime       time.Duration
	BackgroundColor    string
	ExternalCallMethod string
	ExternalCallURL    string
	Content            string
	RequestNameOtel    string
	OTELTracer         trace.Tracer
}
