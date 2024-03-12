package usecase

import (
	"context"

	"github.com/MatheusBenetti/opentelemetry/internal/temperature/dto"
	"github.com/MatheusBenetti/opentelemetry/internal/temperature/entity"
)

type GetWeather struct {
	locationRepo entity.LocationRepository
	tempRepo     entity.TemperatureRepository
}

func NewGetWeather(
	locationRepo entity.LocationRepository,
	tempRepo entity.TemperatureRepository,
) GetWeather {
	return GetWeather{
		locationRepo: locationRepo,
		tempRepo:     tempRepo,
	}
}

func (gw *GetWeather) Execute(
	ctx context.Context,
	input dto.LocationInput,
) (dto.TemperatureOutput, error) {
	if cepErr := entity.CEPValidation(input.CEP); cepErr != nil {
		return dto.TemperatureOutput{}, cepErr
	}

	location, err := gw.locationRepo.Get(ctx, input.CEP)
	if err != nil {
		return dto.TemperatureOutput{}, err
	}

	temperature, err := gw.tempRepo.Get(ctx, location.Localidade)
	if err != nil {
		return dto.TemperatureOutput{}, err
	}

	return dto.TemperatureOutput{
		Location: location.Localidade,
		TempC:    temperature.Celsius(),
		TempF:    temperature.Fahrenheit(),
		TempK:    temperature.Kelvin(),
	}, nil
}
