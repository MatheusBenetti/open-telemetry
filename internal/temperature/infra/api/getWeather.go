package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MatheusBenetti/opentelemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/MatheusBenetti/opentelemetry/internal/temperature/dto"
	"github.com/MatheusBenetti/opentelemetry/internal/temperature/entity"
)

var createWeatherEndpoint = func(baseUrl string) string {
	return strings.Join([]string{baseUrl, "v1", "current.json"}, "/")
}

type WeatherFromAPI struct {
	config *config.Config
}

func NewWeatherFromAPI(config *config.Config) *WeatherFromAPI {
	return &WeatherFromAPI{
		config: config,
	}
}

func (wap *WeatherFromAPI) Get(ctx context.Context, location string) (entity.Temperature, error) {
	hCtx := otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier{})
	tracer := otel.Tracer("serviceBGetWeather")
	_, span := tracer.Start(hCtx, "service_b:get_weather")
	defer span.End()

	u, urlErr := url.Parse(createWeatherEndpoint(wap.config.Temperature.URL))
	if urlErr != nil {
		fmt.Printf("Error parsing URL: %s\n", urlErr)
		return entity.Temperature{}, urlErr
	}
	apiKey := wap.config.Temperature.ApiKey
	if apiKey == "" {
		return entity.Temperature{}, entity.ErrEmptyAPIkey
	}

	q := u.Query()
	q.Set("key", wap.config.Temperature.ApiKey)
	q.Set("q", location)
	q.Set("aqi", "no")
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if reqErr != nil {
		fmt.Printf("Error creating request: %s\n", reqErr)
		return entity.Temperature{}, urlErr
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		fmt.Printf("Error making GET request: %s\n", doErr)
		return entity.Temperature{}, doErr
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Printf("Error reading response body: %s\n", readErr)
		return entity.Temperature{}, readErr
	}

	var weatherData dto.TemperatureResponseOut
	if unmErr := json.Unmarshal(bodyBytes, &weatherData); unmErr != nil {
		fmt.Printf("Error parsing JSON: %s\n", unmErr)
		return entity.Temperature{}, unmErr
	}

	return *entity.NewTemperature(weatherData.Current.TempC), nil
}
