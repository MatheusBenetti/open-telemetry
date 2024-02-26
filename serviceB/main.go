package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

type Current struct {
	WeatherResponse `json:"current"`
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
	http.HandleFunc("/getTemperature", func(w http.ResponseWriter, r *http.Request) {
		cep := r.URL.Query().Get("cep")

		if len(cep) != 8 || !isStringNumeric(cep) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		}

		viaCEP, err := fetchViaCep(cep)
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

		removeSpaces := strings.ReplaceAll(viaCEP.Localidade, " ", "-")

		weather, err := fetchWeatherAPI(removeSpaces)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error getting weather data"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(weather.WeatherResponse)
	})

	log.Println("Server listening on port :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func isStringNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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

func fetchWeatherAPI(location string) (*Current, error) {
	req, err := http.Get("http://api.weatherapi.com/v1/current.json?q=" + location + "&key=50dbab8a6094453b8d4214401242301")

	if err != nil {
		return nil, fmt.Errorf("failed to make request to Weather API: %v", err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var data Current
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	var fahrenheit float64 = data.WeatherResponse.TempC
	var kelvin float64 = data.WeatherResponse.TempC

	formatString := strings.ReplaceAll(location, "-", " ")

	data.WeatherResponse.City = formatString
	data.WeatherResponse.TempF = celsiusToFahrenheit(fahrenheit)
	data.WeatherResponse.TempK = celsiusToKelvin(kelvin)
	return &data, nil
}

func celsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 1.8) + 32
}

func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273
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
