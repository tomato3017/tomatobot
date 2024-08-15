package tgapi

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TGBotMsg struct {
	innerMsg      *tgbotapi.Message
	assumedChatID int64
	assumedUserID int64
}

func NewTGBotMsg(msg *tgbotapi.Message, assumedChatID, assumedUserID int64) TGBotMsg {
	return TGBotMsg{
		innerMsg:      msg,
		assumedChatID: assumedChatID,
		assumedUserID: assumedUserID,
	}
}

func (t *TGBotMsg) AssumedChatID() int64 {
	return t.assumedChatID
}

func (t *TGBotMsg) AssumedUserID() int64 {
	return t.assumedUserID
}

func (t *TGBotMsg) InnerMsg() *tgbotapi.Message {
	return t.innerMsg
}
