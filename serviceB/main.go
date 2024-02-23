package web

import (
	"encoding/json"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type TemperatureResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	initTracer()

	http.HandleFunc("/temperature", handleTemperature)
	log.Println("Service B listening on port :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func handleTemperature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Cep string `json:"cep"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cep := req.Cep
	if len(cep) != 8 || !isStringNumeric(cep) {
		http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// Simulate finding location and temperature
	location := "São Paulo" // Assuming it's São Paulo for demonstration
	temperature := 28.5     // Assuming it's 28.5°C for demonstration

	resp := TemperatureResponse{
		City:  location,
		TempC: temperature,
		TempF: celsiusToFahrenheit(temperature),
		TempK: celsiusToKelvin(temperature),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func isStringNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func celsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 9 / 5) + 32
}

func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273.15
}

func initTracer() {
	exporter, err := zipkin.New(
		"http://localhost:9411/api/v2/spans",
		zipkin.WithSDKOptions(sdktrace.WithSampler(sdktrace.AlwaysSample())),
	)
	if err != nil {
		log.Fatalf("Failed to create Zipkin exporter: %v", err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
}
