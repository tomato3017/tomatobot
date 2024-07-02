package util

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func NewMessagePrivate(msg tgbotapi.Message, parseMode string, text string) tgbotapi.MessageConfig {
	if msg.Chat.IsPrivate() {
		return tgbotapi.NewMessage(msg.Chat.ID, text)
	} else {
		return tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: msg.From.ID,
			},
			ParseMode:             parseMode,
			Text:                  text,
			DisableWebPagePreview: false,
		}
	}
}

func NewMessageReply(msg *tgbotapi.Message, parseMode string, text string) tgbotapi.MessageConfig {
	return tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: msg.MessageID,
		},
		ParseMode:             parseMode,
		Text:                  text,
		DisableWebPagePreview: false,
	}
}

func EscapeStrings(a ...interface{}) []interface{} {
	escaped := make([]interface{}, len(a))
	for i, v := range a {
		if str, ok := v.(string); ok {
			escaped[i] = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, str)
		} else {
			escaped[i] = v
		}
	}
	return escaped
}
