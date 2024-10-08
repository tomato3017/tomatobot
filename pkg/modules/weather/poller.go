package weather

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/modules/weather/owm"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

const WeatherPollerTopic = "weather"

const (
	topicFmtWatch    = "%s.%s.watch"
	topicFmtWarning  = "%s.%s.warning"
	topicFmtAdvisory = "%s.%s.advisory"
	topicFmtUnknown  = "%s.%s.unknown"
)

var numberedStormRegex = regexp.MustCompile(`(?m)((\w+)\s(WARNING|WATCH)\s\d+)\s`)

//go:embed weatheralert.tmpl
var msgTemplateStr string

type poller struct {
	publisher notifications.Publisher

	locations []dbmodels.WeatherPollingLocations

	ctxCf func()
	wg    sync.WaitGroup
	cfg   config.WeatherConfig

	logger      zerolog.Logger
	client      owm.OpenWeatherMapIClient
	dbConn      bun.IDB
	msgTemplate *template.Template
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

func (p *poller) isLowerAdvisory(alertNameUpper string) bool {
	switch {
	case
		strings.Contains(alertNameUpper, "RED FLAG"),
		strings.Contains(alertNameUpper, "FIRE WEATHER"),
		strings.Contains(alertNameUpper, "EXCESSIVE HEAT"),
		strings.Contains(alertNameUpper, "WIND CHILL"),
		strings.Contains(alertNameUpper, "FREEZE"):
		return true
	}
	return false
}

func (p *poller) topicName(location dbmodels.WeatherPollingLocations, alert owm.Alerts) string {
	alertNameUpper := strings.ToUpper(alert.Event)

	switch {
	case p.isLowerAdvisory(alertNameUpper):
		return fmt.Sprintf(topicFmtAdvisory, WeatherPollerTopic, location.ZipCode)
	case strings.Contains(alertNameUpper, "WATCH"):
		return fmt.Sprintf(topicFmtWatch, WeatherPollerTopic, location.ZipCode)
	case strings.Contains(alertNameUpper, "WARNING"):
		return fmt.Sprintf(topicFmtWarning, WeatherPollerTopic, location.ZipCode)
	case strings.Contains(alertNameUpper, "ADVISORY"):
		return fmt.Sprintf(topicFmtAdvisory, WeatherPollerTopic, location.ZipCode)
	default:
		p.logger.Error().Msgf("Unknown alert type: %s", alert.Event)
		return fmt.Sprintf(topicFmtUnknown, WeatherPollerTopic, location.ZipCode)
	}
}

func (p *poller) getDedupeKey(location dbmodels.WeatherPollingLocations, alert owm.Alerts) string {
	matches := numberedStormRegex.FindAllStringSubmatch(alert.Description, -1)
	if len(matches) > 0 {
		return fmt.Sprintf("%s_%s", util.FirstNonZero(location.Name, location.ZipCode), matches[0][1])
	}

	return fmt.Sprintf("%s_%s_%d", util.FirstNonZero(location.Name, location.ZipCode), alert.Event, alert.End)
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

		renderedMsg, err := p.getRenderedWeatherAlert(alert, location)
		if err != nil {
			return fmt.Errorf("failed to render weather alert: %w", err)
		}
		p.publisher.Publish(notifications.Message{
			Topic:   topicName,
			Msg:     renderedMsg,
			DupeTTL: p.getDedupeTTL(alert),
			DupeKey: p.getDedupeKey(location, alert),
		})
	}

	return nil
}

func (p *poller) getRenderedWeatherAlert(alert owm.Alerts, location dbmodels.WeatherPollingLocations) (string, error) {
	msgBuffer := bytes.Buffer{}
	err := p.msgTemplate.Execute(&msgBuffer, tgWeatherAlert{
		Alerts: owm.Alerts{
			Event:       alert.Event,
			Start:       alert.Start,
			End:         alert.End,
			Description: util.TruncateString(alert.Description, 512, "..."),
		},
		WeatherPollingLocations: dbmodels.WeatherPollingLocations{
			Name: location.Name,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to render weather alert: %w", err)
	}

	return msgBuffer.String(), nil
}

type pollerNewArgs struct {
	publisher notifications.Publisher
	locations []dbmodels.WeatherPollingLocations
	cfg       config.WeatherConfig
	logger    zerolog.Logger
	dbConn    bun.IDB
}

func newPoller(args pollerNewArgs) *poller {
	client, err := owm.NewOpenWeatherMapClient(args.cfg.APIKey)
	if err != nil {
		args.logger.Fatal().Err(err).Msg("Failed to create OpenWeatherMap client")
	}

	tmplFuncMap := template.FuncMap{
		"int64ToTime": int64ToTime,
	}

	msgTemplate, err := template.New("weatheralert").Funcs(tmplFuncMap).Parse(msgTemplateStr)
	if err != nil {
		args.logger.Fatal().Err(err).Msg("Failed to parse weather alert template")
	}

	return &poller{
		publisher:   args.publisher,
		locations:   args.locations,
		cfg:         args.cfg,
		logger:      args.logger,
		client:      client,
		dbConn:      args.dbConn,
		msgTemplate: msgTemplate,
	}
}
func int64ToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}
