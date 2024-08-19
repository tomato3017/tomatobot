package birthday

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/uptrace/bun"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"sync"
	"text/template"
	"time"
)

const BirthdayPollerTopic = "birthday"

// TODO stopped here. Needs to add the birthday check and announcement logic
//
//go:embed birthdaymsg.tmpl
var msgTemplateStr string

type birthdayMsg struct {
	Name string
	Age  int
}

type poller struct {
	publisher notifications.Publisher

	ctxCf  context.CancelFunc
	dbConn bun.IDB
	logger zerolog.Logger
	wg     *sync.WaitGroup

	msgTemplate *template.Template
}

func newPoller(publisher notifications.Publisher, dbConn bun.IDB, logger zerolog.Logger) (*poller, error) {
	tmpl, err := template.New("birthdaymsg").Parse(msgTemplateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message template: %w", err)
	}

	return &poller{
		wg:          &sync.WaitGroup{},
		publisher:   publisher,
		dbConn:      dbConn,
		logger:      logger.With().Str("thread", "poller").Logger(),
		msgTemplate: tmpl,
	}, nil
}

func (p *poller) Start(ctx context.Context) {
	ctx, cf := context.WithCancel(ctx)
	p.ctxCf = cf

	p.wg.Add(1)
	go func(ctx context.Context) {
		defer p.wg.Done()
		if err := p.poll(ctx); err != nil {
			p.logger.Error().Err(err).Msg("Failed to check birthdays")
		}
	}(ctx)
}

func (p *poller) Stop() {
	p.ctxCf()
	p.wg.Wait()
}

func (p *poller) poll(ctx context.Context) error {
	bTicker := time.NewTicker(time.Minute)
	defer bTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug().Msg("Context done, stopping poller")
			return nil
		case <-bTicker.C:
			p.logger.Debug().Msg("Polling birthdays")
			if err := p.birthdayCheck(ctx); err != nil {
				return err
			}
		}
	}
}

func (p *poller) birthdayCheck(ctx context.Context) error {
	//Get all the birthdays that are today
	currentTime := time.Now()
	_, month, day := currentTime.Date()
	birthdays, err := p.getBirthdays(ctx, day, month)
	if err != nil {
		return fmt.Errorf("failed to get birthdays: %w", err)
	}

	if len(birthdays) == 0 {
		return nil
	}

	if err := p.announceBirthdays(ctx, birthdays); err != nil {
		return fmt.Errorf("failed to announce birthdays: %w", err)
	}

	return nil
}

func (p *poller) getBirthdays(ctx context.Context, day int, month time.Month) ([]dbmodels.Birthdays, error) {
	var birthdays []dbmodels.Birthdays
	err := p.dbConn.NewSelect().Model(&birthdays).
		Where("day = ? AND month = ? AND last_announced_at< ?", day, month, time.Now().Add(time.Hour*-24)).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get birthdays: %w", err)
	}

	return birthdays, nil
}

func (p *poller) announceBirthdays(ctx context.Context, birthdays []dbmodels.Birthdays) error {
	p.logger.Debug().Msg("Announcing birthdays")
	for _, birthday := range birthdays {
		p.logger.Trace().Msgf("Announcing birthday for %s with age %d", birthday.Name, p.getAge(birthday))
		msg, err := p.generateBirthdayMessage(birthday)
		if err != nil {
			return fmt.Errorf("failed to generate birthday message: %w", err)
		}

		birthdayTopic := fmt.Sprintf("%s.%d", BirthdayPollerTopic, birthday.ChatId)

		p.publisher.Publish(notifications.Message{
			Topic:   birthdayTopic,
			Msg:     msg,
			DupeTTL: time.Hour * 24,
		})

		if err := p.setLastAnnouncedBirthday(ctx, birthday); err != nil {
			return fmt.Errorf("failed to set last announced birthday: %w", err)
		}
	}

	return nil
}

func (p *poller) setLastAnnouncedBirthday(ctx context.Context, birthday dbmodels.Birthdays) error {
	_, err := p.dbConn.NewUpdate().Model(&birthday).Set("last_announced_at = ?", time.Now()).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set last announced birthday: %w", err)
	}

	return nil
}

func (p *poller) getAge(birthday dbmodels.Birthdays) int {
	if birthday.Year == 0 {
		return 0
	}

	return time.Now().Year() - birthday.Year
}

func (p *poller) generateBirthdayMessage(birthday dbmodels.Birthdays) (string, error) {
	caser := cases.Title(language.AmericanEnglish)
	newBirthdayMsg := birthdayMsg{
		Name: caser.String(birthday.Name),
		Age:  p.getAge(birthday),
	}

	var msg strings.Builder
	err := p.msgTemplate.Execute(&msg, newBirthdayMsg)
	if err != nil {
		return "", fmt.Errorf("failed to execute message template: %w", err)
	}

	return msg.String(), nil
}
