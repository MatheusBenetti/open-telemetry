package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Erro        bool   `json:"erro"`
}

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
	http.HandleFunc("/getTemperature", func(w http.ResponseWriter, r *http.Request) {
		var data Input
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request body"))
			return
		}
		cep := data.Cep

		log.Println(cep)

		if len(cep) != 8 || !isStringNumeric(cep) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		}

		viaCEP, err := fetchViaCep(cep)
		log.Println(viaCEP)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("can not found zipcode"))
			return
		}

		if viaCEP.Erro {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("can not found zipcode"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(viaCEP.Localidade)
	})

	log.Println("Server listening on port :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func fetchViaCep(cep string) (*ViaCEP, error) {
	req, err := http.Get("http://viacep.com.br/ws/" + cep + "/json/")

	if err != nil {
		return nil, fmt.Errorf("failed to make request to ViaCEP API: %v", err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var data ViaCEP
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &data, nil
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
