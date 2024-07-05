package weather

import (
	"context"
	"fmt"
	"github.com/rclone/debughttp"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/modules/weather/owm"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"sync"
	"time"
)

const WeatherPollerTopic = "weather"

type poller struct {
	publisher notifications.Publisher

	locations []dbmodels.WeatherPollingLocation

	ctxCf func()
	wg    sync.WaitGroup
	cfg   config.WeatherConfig

	logger zerolog.Logger
	client owm.OpenWeatherMapIClient
}

func (p *poller) poll(ctx context.Context) {
	p.logger.Debug().Msgf("Polling weather every %s", p.cfg.PollingInterval.String())
	tick := time.NewTicker(p.cfg.PollingInterval)
	defer tick.Stop()

	for {
		p.logger.Debug().Msg("Polling weather")
		select {
		case <-ctx.Done():
			p.logger.Debug().Msg("Context done, stopping poller")
			return
		case <-tick.C:
			p.publishWeatherForLocations(ctx)
		}
	}
}

func (p *poller) publishWeatherForLocations(ctx context.Context) {
	for _, location := range p.locations {
		p.logger.Debug().Msgf("Publishing weather for location %s, lat %f, long %f", location.ZipCode, location.Lat, location.Lon)
		if err := p.publishWeatherForLocation(ctx, location); err != nil {
			p.logger.Error().Err(err).Msgf("Failed to publish weather for location %s", location.ZipCode)
			continue
		}
	}
}

func (p *poller) Start(ctx context.Context) {
	ctx, cf := context.WithCancel(ctx)
	p.ctxCf = cf

	p.wg.Add(1)
	go func(ctx context.Context) {
		defer p.wg.Done()
		p.poll(ctx)
	}(ctx)
}

func (p *poller) Stop() {
	p.ctxCf()
	p.wg.Wait()
}

func (p *poller) topicName(location dbmodels.WeatherPollingLocation) string {
	return fmt.Sprintf("%s.%s.alerts", WeatherPollerTopic, location.ZipCode)
}

func (p *poller) publishWeatherForLocation(ctx context.Context, location dbmodels.WeatherPollingLocation) error {
	res, err := p.client.CurrentWeatherByLocation(owm.Location{
		Latitude:  location.Lat,
		Longitude: location.Lon,
	})
	if err != nil {
		return err
	}

	for _, alert := range res.Alerts {
		p.logger.Trace().Msgf("Publishing weather alert for location %s, event %s, start %d, end %d",
			location.ZipCode, alert.Event, alert.Start, alert.End)
		p.publisher.Publish(notifications.Message{
			Topic: p.topicName(location),
			Msg: fmt.Sprintf("Weather Alert: Location %s, Event %s, Start %s, End %s", location.ZipCode,
				alert.Event,
				time.Unix(int64(alert.Start), 0).Format(time.RFC3339),
				time.Unix(int64(alert.End), 0).Format(time.RFC3339)),
			DupeTTL: time.Hour * 6, //TODO use alert end or max duration
		})
	}

	return nil
}

type pollerNewArgs struct {
	publisher notifications.Publisher
	locations []dbmodels.WeatherPollingLocation
	cfg       config.WeatherConfig
	logger    zerolog.Logger
}

func newPoller(args pollerNewArgs) *poller {
	httpClient := debughttp.NewClient(nil)
	client, err := owm.NewOpenWeatherMapClient(args.cfg.APIKey, owm.WithHTTPClient(httpClient))
	if err != nil {
		args.logger.Fatal().Err(err).Msg("Failed to create OpenWeatherMap client")
	}

	return &poller{
		publisher: args.publisher,
		locations: args.locations,
		cfg:       args.cfg,
		logger:    args.logger,
		client:    client}
}
