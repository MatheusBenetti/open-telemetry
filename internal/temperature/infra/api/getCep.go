package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MatheusBenetti/opentelemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/MatheusBenetti/opentelemetry/internal/temperature/dto"
	"github.com/MatheusBenetti/opentelemetry/internal/temperature/entity"
)

var createCepEndpoint = func(baseUrl, cep string) string {
	return strings.Join([]string{baseUrl, "ws", cep, "json"}, "/")
}

type CEPFromAPI struct {
	config *config.Config
}

func NewCEPFromAPI(config *config.Config) *CEPFromAPI {
	return &CEPFromAPI{
		config: config,
	}
}

func (cap *CEPFromAPI) Get(ctx context.Context, cep string) (entity.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	hCtx := otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier{})
	tracer := otel.Tracer("serviceBGetCEP")
	_, span := tracer.Start(hCtx, "service_b:get_CEP")
	defer span.End()
	defer cancel()

	req, reqErr := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		createCepEndpoint(cap.config.CEP.URL, cep),
		nil,
	)
	if reqErr != nil {
		return entity.Location{}, reqErr
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
		return entity.Location{}, doErr
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return entity.Location{}, readErr
	}

	var location dto.LocationOut
	if unmErr := json.Unmarshal(bodyBytes, &location); unmErr != nil {
		return entity.Location{}, unmErr
	}

	if location.CEP == "" {
		return entity.Location{}, entity.ErrCEPNotFound
	}

	return entity.Location{
		Cep:        location.CEP,
		Localidade: location.Localidade,
	}, nil
}
