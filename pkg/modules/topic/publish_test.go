package topic

import (
	"context"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tomato3017/tomatobot/pkg/bot/models/tgapi"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/notifications"
)

type TopicPublishCmdSuite struct {
	suite.Suite
}

func (s *TopicPublishCmdSuite) newParams(args []string) models.CommandParams {
	msg := tgapi.NewTGBotMsg(&tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 123},
	}, tgapi.TGBotAssumedIds{ChatID: 123, UserID: 456}, nil)
	return models.CommandParams{
		CommandName: "publish",
		Args:        args,
		Message:     msg,
	}
}

func (s *TopicPublishCmdSuite) Test_WithDedupeKey_SubmitsExpectedMessage() {
	mockPub := notifications.NewMockPublisher(s.T())
	mockPub.EXPECT().Publish(notifications.Message{
		Topic:   "weather.warning",
		Msg:     "big storm",
		DupeKey: "key1",
	}).Return()

	mockBot := proxy.NewMockTGBotImplementation(s.T())
	mockBot.EXPECT().Send(mock.Anything).Return(tgbotapi.Message{}, nil)

	cmd := newTopicPublishCmd(mockPub, mockBot, zerolog.Nop())
	err := cmd.Execute(context.Background(), s.newParams([]string{"weather.warning", "big storm", "key1"}))
	require.NoError(s.T(), err)
}

func (s *TopicPublishCmdSuite) Test_WithoutDedupeKey_EmptyDupeKey() {
	mockPub := notifications.NewMockPublisher(s.T())
	mockPub.EXPECT().Publish(notifications.Message{
		Topic:   "weather.warning",
		Msg:     "test",
		DupeKey: "",
	}).Return()

	mockBot := proxy.NewMockTGBotImplementation(s.T())
	mockBot.EXPECT().Send(mock.Anything).Return(tgbotapi.Message{}, nil)

	cmd := newTopicPublishCmd(mockPub, mockBot, zerolog.Nop())
	err := cmd.Execute(context.Background(), s.newParams([]string{"weather.warning", "test"}))
	require.NoError(s.T(), err)
}

func (s *TopicPublishCmdSuite) Test_TooFewArgs_Rejected() {
	mockPub := notifications.NewMockPublisher(s.T())
	mockBot := proxy.NewMockTGBotImplementation(s.T())

	cmd := newTopicPublishCmd(mockPub, mockBot, zerolog.Nop())
	err := cmd.RunMiddleware(context.Background(), s.newParams([]string{"weather.warning"}))
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "at least 2")
}

func (s *TopicPublishCmdSuite) Test_TooManyArgs_Rejected() {
	mockPub := notifications.NewMockPublisher(s.T())
	mockBot := proxy.NewMockTGBotImplementation(s.T())

	cmd := newTopicPublishCmd(mockPub, mockBot, zerolog.Nop())
	// 4 args: topic, word1, word2, word3 — unquoted multi-word body
	err := cmd.RunMiddleware(context.Background(), s.newParams([]string{"weather.warning", "big", "storm", "key1"}))
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "at most 3")
}

func (s *TopicPublishCmdSuite) Test_EmptyMessage_Rejected() {
	mockPub := notifications.NewMockPublisher(s.T())
	mockBot := proxy.NewMockTGBotImplementation(s.T())

	cmd := newTopicPublishCmd(mockPub, mockBot, zerolog.Nop())
	err := cmd.Execute(context.Background(), s.newParams([]string{"weather.warning", ""}))
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "empty")
}

func (s *TopicPublishCmdSuite) Test_MalformedTopic_Rejected() {
	mockPub := notifications.NewMockPublisher(s.T())
	mockBot := proxy.NewMockTGBotImplementation(s.T())

	cmd := newTopicPublishCmd(mockPub, mockBot, zerolog.Nop())
	err := cmd.Execute(context.Background(), s.newParams([]string{"bad topic!", "some message"}))
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "invalid topic format")
}

func (s *TopicPublishCmdSuite) Test_Success_ConfirmationContainsTopicNotBody() {
	const topic = "news.daily"
	const body = "hello there world"

	mockPub := notifications.NewMockPublisher(s.T())
	mockPub.EXPECT().Publish(notifications.Message{
		Topic: topic,
		Msg:   body,
	}).Return()

	var sentMsg tgbotapi.Chattable
	mockBot := proxy.NewMockTGBotImplementation(s.T())
	mockBot.EXPECT().Send(mock.Anything).
		Run(func(c tgbotapi.Chattable) { sentMsg = c }).
		Return(tgbotapi.Message{}, nil)

	cmd := newTopicPublishCmd(mockPub, mockBot, zerolog.Nop())
	err := cmd.Execute(context.Background(), s.newParams([]string{topic, body}))
	require.NoError(s.T(), err)

	msgConfig, ok := sentMsg.(tgbotapi.MessageConfig)
	require.True(s.T(), ok, "expected tgbotapi.MessageConfig")
	require.Contains(s.T(), msgConfig.Text, tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, topic))
	require.NotContains(s.T(), msgConfig.Text, body)
}

func TestTopicPublishCmdSuite(t *testing.T) {
	suite.Run(t, new(TopicPublishCmdSuite))
}
