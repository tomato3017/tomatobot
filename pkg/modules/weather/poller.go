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
	"github.com/uptrace/bun"
	"strings"
	"sync"
	"time"
)

const WeatherPollerTopic = "weather"

type poller struct {
	publisher notifications.Publisher

	locations []dbmodels.WeatherPollingLocations

	ctxCf func()
	wg    sync.WaitGroup
	cfg   config.WeatherConfig

	logger zerolog.Logger
	client owm.OpenWeatherMapIClient
	dbConn bun.IDB
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
			if err := p.updateWeatherLocations(ctx); err != nil {
				p.logger.Fatal().Msgf("Failed to update weather locations: %s", err.Error())
			}
			p.publishWeatherForLocations(ctx)
		}
	}
}

func (p *poller) updateWeatherLocations(ctx context.Context) error {
	locations := make([]dbmodels.WeatherPollingLocations, 0)

	err := p.dbConn.NewSelect().Model(&locations).Where("polling = ?", true).Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to get weather polling locations: %w", err)
	}

	p.locations = locations
	return nil
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

func (p *poller) topicName(location dbmodels.WeatherPollingLocations, alert owm.Alerts) string {
	alertNameUpper := strings.ToUpper(alert.Event)

	switch {
	case strings.Contains(alertNameUpper, "WATCH"):
		return fmt.Sprintf("%s.%s.watch", WeatherPollerTopic, location.ZipCode)
	case strings.Contains(alertNameUpper, "WARNING"):
		return fmt.Sprintf("%s.%s.warning", WeatherPollerTopic, location.ZipCode)
	case strings.Contains(alertNameUpper, "ADVISORY"):
		return fmt.Sprintf("%s.%s.advisory", WeatherPollerTopic, location.ZipCode)
	default:
		panic("Unknown alert type")
	}
}

func (p *poller) getDedupeTTL(alert owm.Alerts) time.Duration {
	return time.Until(time.Unix(alert.End, 0).Add(10 * time.Minute))
}

func (p *poller) publishWeatherForLocation(ctx context.Context, location dbmodels.WeatherPollingLocations) error {
	res, err := p.client.CurrentWeatherByLocation(ctx, owm.Location{
		Latitude:  location.Lat,
		Longitude: location.Lon,
	})
	if err != nil {
		return err
	}

	for _, alert := range res.Alerts {
		p.logger.Trace().Msgf("Publishing weather alert for location %s, event %s, start %d, end %d",
			location.ZipCode, alert.Event, alert.Start, alert.End)

		topicName := p.topicName(location, alert)
		p.logger.Trace().Msgf("Topic name: %s", topicName)

		p.publisher.Publish(notifications.Message{
			Topic: topicName,
			Msg: fmt.Sprintf("Weather Alert: Location %s, Event %s, Start %s, End %s", location.ZipCode,
				alert.Event,
				time.Unix(alert.Start, 0).Format(time.RFC3339),
				time.Unix(alert.End, 0).Format(time.RFC3339)),
			DupeTTL: p.getDedupeTTL(alert),
		})
	}

	return nil
}

type pollerNewArgs struct {
	publisher notifications.Publisher
	locations []dbmodels.WeatherPollingLocations
	cfg       config.WeatherConfig
	logger    zerolog.Logger
	dbConn    bun.IDB
}

func newPoller(args pollerNewArgs) *poller {
	httpClient := debughttp.NewClient(nil) //TODO remove
	client, err := owm.NewOpenWeatherMapClient(args.cfg.APIKey, owm.WithHTTPClient(httpClient))
	if err != nil {
		args.logger.Fatal().Err(err).Msg("Failed to create OpenWeatherMap client")
	}

	return &poller{
		publisher: args.publisher,
		locations: args.locations,
		cfg:       args.cfg,
		logger:    args.logger,
		client:    client,
		dbConn:    args.dbConn}
}
