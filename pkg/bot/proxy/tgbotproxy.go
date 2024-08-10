package proxy

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/config"
)

type TGBotProxy struct {
	sendToChatChannels bool
	tgbot              *tgbotapi.BotAPI
	logger             zerolog.Logger
	cfg                config.TomatoBot
}

func (t *TGBotProxy) SendPrivate(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	//TODO implement me
	panic("implement me")
}

var _ TGBotImplementation = &TGBotProxy{}

type ProxyOption func(*TGBotProxy) error

func NewTGBotProxy(tgbot *tgbotapi.BotAPI, options ...ProxyOption) (*TGBotProxy, error) {
	tgBotProxy := &TGBotProxy{
		tgbot:              tgbot,
		sendToChatChannels: true,
		logger:             zerolog.Nop(),
	}

	for _, option := range options {
		if err := option(tgBotProxy); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	tgBotProxy.logger.Trace().Msgf("Bot proxy set to send messages to chat channels: %t", tgBotProxy.sendToChatChannels)

	return tgBotProxy, nil
}

func (t *TGBotProxy) IsBotAdmin(userId int64) bool {
	return t.cfg.IsBotAdmin(userId)
}

func (t *TGBotProxy) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return t.handleSendable(c)
}

func (t *TGBotProxy) InnerBotAPI() *tgbotapi.BotAPI {
	return t.tgbot
}

func (t *TGBotProxy) handleSendable(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	switch v := c.(type) {
	case tgbotapi.MessageConfig:
		return t.handleMessageConfig(v)
	default:
		return tgbotapi.Message{}, fmt.Errorf("unsupported type: %T", v)
	}
}

func (t *TGBotProxy) handleMessageConfig(msgCfg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	if !t.idIsChat(msgCfg.ChatID) {
		return t.tgbot.Send(msgCfg)
	} else if !t.sendToChatChannels && t.idIsChat(msgCfg.ChatID) {
		t.logger.Trace().Msgf("Not sending message to chat channel: %+v", msgCfg)
		return tgbotapi.Message{}, nil
	}

	return t.tgbot.Send(msgCfg)
}

func (t *TGBotProxy) idIsChat(chatID int64) bool {
	return chatID < 0
}
