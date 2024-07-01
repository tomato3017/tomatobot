package util

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func NewMessagePrivate(msg tgbotapi.Message, text string) tgbotapi.MessageConfig {
	if msg.Chat.IsPrivate() {
		return tgbotapi.NewMessage(msg.Chat.ID, text)
	} else {
		return tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: msg.From.ID,
			},
			Text:                  text,
			DisableWebPagePreview: false,
		}
	}
}

func NewMessageReply(msg *tgbotapi.Message, text string) tgbotapi.MessageConfig {
	return tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: msg.MessageID,
		},
		Text:                  text,
		DisableWebPagePreview: false,
	}
}
