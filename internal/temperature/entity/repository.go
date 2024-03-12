package entity

import (
	"context"
)

type LocationRepository interface {
	Get(ctx context.Context, cep string) (Location, error)
}

type TemperatureRepository interface {
	Get(ctx context.Context, location string) (Temperature, error)
}
