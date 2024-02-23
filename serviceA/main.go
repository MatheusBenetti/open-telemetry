package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Input struct {
	Cep string `json:"cep"`
}

type ValidResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	initTracer()

	http.HandleFunc("/input", handleInput)
	log.Println("Service A listening on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cep := input.Cep
	if len(cep) != 8 || !isStringNumeric(cep) {
		http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// Forward the request to Service B
	resp, err := http.Post("http://localhost:8090/temperature", "application/json", strings.NewReader(fmt.Sprintf(`{"cep": "%s"}`, cep)))
	if err != nil {
		log.Printf("Error forwarding request to Service B: %v\n", err)
		http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			log.Printf("Error decoding error response from Service B: %v\n", err)
			http.Error(w, "Error decoding error response", http.StatusInternalServerError)
			return
		}
		http.Error(w, errResp.Message, resp.StatusCode)
		return
	}

	var successResp ValidResponse
	if err := json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		log.Printf("Error decoding success response from Service B: %v\n", err)
		http.Error(w, "Error decoding success response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(successResp)
}

func isStringNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
