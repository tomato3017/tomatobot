package proxy

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TGBotSendable interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type TGBotImplementation interface {
	TGBotSendable
	InnerBotAPI() *tgbotapi.BotAPI
	SendPrivate(c tgbotapi.Chattable) (tgbotapi.Message, error)
	IsBotAdmin(userId int64) bool
	IdIsChat(chatID int64) bool
}
