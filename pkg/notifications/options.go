package notifications

import "github.com/rs/zerolog"

type PublisherOptions func(p *NotificationPublisher)

func WithLogger(logger zerolog.Logger) PublisherOptions {
	return func(p *NotificationPublisher) {
		p.logger = logger
	}
}

// WithBusSize sets the size of the bus channel, defaults to unbuffered
func WithBusSize(size int) PublisherOptions {
	return func(p *NotificationPublisher) {
		p.bus = make(chan Message, size)
	}
}
