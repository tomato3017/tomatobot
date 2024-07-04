package notifications

import (
	"github.com/jellydator/ttlcache/v3"
	"github.com/rs/zerolog"
	"time"
)

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

func WithCacheTTL(ttl time.Duration) PublisherOptions {
	return func(p *NotificationPublisher) {
		p.cache = ttlcache.New[string, []int64](
			ttlcache.WithTTL[string, []int64](ttl))
	}
}
