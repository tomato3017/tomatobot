package notifications

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
)

var ErrCancelNotification = errors.New("cancel notification")

// Hook is a per-recipient callback for the Predupe and OnSend lifecycle points.
type Hook func(ctx context.Context, msg Message, chatId int64) error

// EnrichHook is a once-per-message callback for the AfterDedupe lifecycle point.
// It may mutate the shared message body.
type EnrichHook func(ctx context.Context, msg *Message) error

type HookRegistrar interface {
	RegisterPreDupe(topicPattern string, hook Hook)
	RegisterAfterDedupe(topicPattern string, hook EnrichHook)
	RegisterOnSend(topicPattern string, hook Hook)
	FirePreDupe(ctx context.Context, msg Message, chatId int64) error
	FireAfterDedupe(ctx context.Context, msg *Message) error
	FireOnSend(ctx context.Context, msg Message, chatId int64) error
}

type hookEntry struct {
	pattern string
	hook    Hook
}

type enrichHookEntry struct {
	pattern string
	hook    EnrichHook
}

type hookRegistrar struct {
	mu          sync.RWMutex
	preDupe     []hookEntry
	afterDedupe []enrichHookEntry
	onSend      []hookEntry
}

func newHookRegistrar() *hookRegistrar {
	return &hookRegistrar{
		preDupe:     make([]hookEntry, 0),
		afterDedupe: make([]enrichHookEntry, 0),
		onSend:      make([]hookEntry, 0),
	}
}

func (r *hookRegistrar) RegisterPreDupe(topicPattern string, hook Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preDupe = append(r.preDupe, hookEntry{pattern: topicPattern, hook: hook})
}

func (r *hookRegistrar) RegisterAfterDedupe(topicPattern string, hook EnrichHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.afterDedupe = append(r.afterDedupe, enrichHookEntry{pattern: topicPattern, hook: hook})
}

func (r *hookRegistrar) RegisterOnSend(topicPattern string, hook Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onSend = append(r.onSend, hookEntry{pattern: topicPattern, hook: hook})
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

func (r *hookRegistrar) FirePreDupe(ctx context.Context, msg Message, chatId int64) error {
	r.mu.RLock()
	entries := make([]hookEntry, len(r.preDupe))
	copy(entries, r.preDupe)
	r.mu.RUnlock()

	for _, entry := range entries {
		re, err := compileHookPattern(entry.pattern)
		if err != nil {
			return err
		}
		if re.MatchString(msg.Topic) {
			if err := entry.hook(ctx, msg, chatId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *hookRegistrar) FireAfterDedupe(ctx context.Context, msg *Message) error {
	r.mu.RLock()
	entries := make([]enrichHookEntry, len(r.afterDedupe))
	copy(entries, r.afterDedupe)
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

func (r *hookRegistrar) FireOnSend(ctx context.Context, msg Message, chatId int64) error {
	r.mu.RLock()
	entries := make([]hookEntry, len(r.onSend))
	copy(entries, r.onSend)
	r.mu.RUnlock()

	for _, entry := range entries {
		re, err := compileHookPattern(entry.pattern)
		if err != nil {
			return err
		}
		if re.MatchString(msg.Topic) {
			if err := entry.hook(ctx, msg, chatId); err != nil {
				return err
			}
		}
	}
	return nil
}
