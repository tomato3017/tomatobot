package proxy

import "github.com/rs/zerolog"

func WithSendToChatChannels(sendToChatChannels bool) ProxyOption {
	return func(tgBotProxy *TGBotProxy) error {
		tgBotProxy.sendToChatChannels = sendToChatChannels
		return nil
	}
}

func WithLogger(logger zerolog.Logger) ProxyOption {
	return func(tgBotProxy *TGBotProxy) error {
		tgBotProxy.logger = logger
		return nil
	}
}
