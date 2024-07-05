package owm

import "net/http"

type Option func(c *OpenWeatherMapClient) error

func WithHTTPClient(client *http.Client) Option {
	return func(c *OpenWeatherMapClient) error {
		c.client = client
		return nil
	}
}

func WithBaseURL(url string) Option {
	return func(c *OpenWeatherMapClient) error {
		c.url = url
		return nil
	}
}
