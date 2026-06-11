package notifications

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
)

var ErrCancelNotification = errors.New("cancel notification")

type Hook func(ctx context.Context, msg Message) error

type HookRegistrar interface {
	Register(topicPattern string, hook Hook)
	Fire(ctx context.Context, msg Message) error
}

type hookEntry struct {
	pattern string
	hook    Hook
}

type hookRegistrar struct {
	mu      sync.RWMutex
	entries []hookEntry
}

func newHookRegistrar() *hookRegistrar {
	return &hookRegistrar{entries: make([]hookEntry, 0)}
}

func (r *hookRegistrar) Register(topicPattern string, hook Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, hookEntry{pattern: topicPattern, hook: hook})
}

// compileHookPattern builds an anchored regex for hook topic matching.
// Intentionally diverges from getChatIdsForTopic (unanchored) so that
// "weather.warning" matches only "weather.warning", not "weather.warning.severe".
func compileHookPattern(pattern string) (*regexp.Regexp, error) {
	tokenized := strings.ReplaceAll(pattern, "*", "<star>")
	escaped := regexp.QuoteMeta(tokenized)
	detokenized := strings.ReplaceAll(escaped, "<star>", ".*")
	return regexp.Compile("^" + detokenized + "$")
}

func (r *hookRegistrar) Fire(ctx context.Context, msg Message) error {
	r.mu.RLock()
	entries := make([]hookEntry, len(r.entries))
	copy(entries, r.entries)
	r.mu.RUnlock()

	for _, entry := range entries {
		re, err := compileHookPattern(entry.pattern)
		if err != nil {
			return err
		}
		if re.MatchString(msg.Topic) {
			if err := entry.hook(ctx, msg); err != nil {
				return err
			}
		}
	}
	return nil
}
