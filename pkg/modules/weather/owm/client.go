package owm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tomato3017/tomatobot/pkg/util"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	OWMAPIHOST        = "api.openweathermap.org"
	OWMAPIONECALLPATH = "/data/3.0/onecall"
	OWMAPIONECALLURL  = "https://" + OWMAPIHOST + OWMAPIONECALLPATH
)

type Location struct {
	Latitude  float64
	Longitude float64
}

type OpenWeatherMapClient struct {
	client *http.Client

	apiKey string
	url    string
}

func (c *OpenWeatherMapClient) GetLocationDataForZipCode(ctx context.Context, zipCode string) (GeolocationResponse, error) {
	rawUrl := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/zip?zip=%s,us&appid=%s", zipCode, c.apiKey)

	req, err := http.NewRequest(http.MethodGet, rawUrl, nil)
	if err != nil {
		return GeolocationResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return GeolocationResponse{}, fmt.Errorf("failed to perform request: %w", err)
	}
	defer util.CloseSafely(res.Body)

	if res.StatusCode != http.StatusOK {
		return GeolocationResponse{}, fmt.Errorf("failed to get location data for zip code %s: %s", zipCode, res.Status)
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return GeolocationResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response GeolocationResponse
	err = json.Unmarshal(rawBody, &response)
	if err != nil {
		return GeolocationResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

func (c *OpenWeatherMapClient) CurrentWeatherByLocation(ctx context.Context, location Location) (OneCallCurrentResponse, error) {
	finalURL, err := c.getRenderedURL_CurrentLoc(location)
	if err != nil {
		return OneCallCurrentResponse{}, err
	}

	req, err := http.NewRequest(http.MethodGet, finalURL, nil)
	if err != nil {
		return OneCallCurrentResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return OneCallCurrentResponse{}, fmt.Errorf("failed to perform request: %w", err)
	}
	defer util.CloseSafely(res.Body)

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return OneCallCurrentResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response OneCallCurrentResponse
	err = json.Unmarshal(rawBody, &response)
	if err != nil {
		return OneCallCurrentResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

func (c *OpenWeatherMapClient) getRenderedURL_CurrentLoc(location Location) (string, error) {
	baseURL, err := url.Parse(c.url)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	qParams := url.Values{}
	qParams.Add("lat", strconv.FormatFloat(location.Latitude, 'f', -1, 64))
	qParams.Add("lon", strconv.FormatFloat(location.Longitude, 'f', -1, 64))
	qParams.Add("exclude", "minutely,hourly,daily,current")
	qParams.Add("appid", c.apiKey)

	baseURL.RawQuery = qParams.Encode()

	return baseURL.String(), nil
}

func NewOpenWeatherMapClient(apiKey string, options ...Option) (*OpenWeatherMapClient, error) {
	c := &OpenWeatherMapClient{
		apiKey: apiKey,
		client: http.DefaultClient,
		url:    OWMAPIONECALLURL,
	}

	err := setOptions(c, options...)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func setOptions(c *OpenWeatherMapClient, options ...Option) error {
	for _, option := range options {
		err := option(c)
		if err != nil {
			return err
		}
	}

	return nil
}
