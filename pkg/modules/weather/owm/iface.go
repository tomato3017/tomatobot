package owm

type OpenWeatherMapIClient interface {
	CurrentWeatherByLocation(location Location) (OneCallCurrentResponse, error)
}
