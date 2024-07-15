package owm

import "context"

type OpenWeatherMapIClient interface {
	CurrentWeatherByLocation(ctx context.Context, location Location) (OneCallCurrentResponse, error)
	GetLocationDataForZipCode(ctx context.Context, zipCode string) (GeolocationResponse, error)
}
