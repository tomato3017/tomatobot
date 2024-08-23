package tgapi

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TGBotAssumedIds struct {
	ChatID int64
	UserID int64
}

type TGBotMsg struct {
	innerMsg   *tgbotapi.Message
	assumedIds TGBotAssumedIds

	normalizedTextData []SerializableTextData
}

func NewTGBotMsg(msg *tgbotapi.Message, assumedIds TGBotAssumedIds, normalizedData []SerializableTextData) TGBotMsg {
	return TGBotMsg{
		innerMsg:           msg,
		assumedIds:         assumedIds,
		normalizedTextData: normalizedData,
	}
}

func (t *TGBotMsg) AssumedChatID() int64 {
	return t.assumedIds.ChatID
}

func (t *TGBotMsg) AssumedUserID() int64 {
	return t.assumedIds.UserID
}

func (t *TGBotMsg) InnerMsg() *tgbotapi.Message {
	return t.innerMsg
}

func (t *TGBotMsg) NormalizedTextData() []SerializableTextData {
	return t.normalizedTextData
}
