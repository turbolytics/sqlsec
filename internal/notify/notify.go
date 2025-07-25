package notify

import (
	"context"
	"errors"
	"sync"
)

type ChannelType string

const (
	SlackChannel ChannelType = "slack"
	// Future: EmailChannel, PagerDutyChannel, etc.
)

type Notifier interface {
	Test(ctx context.Context, cfg map[string]string) error
	Send(ctx context.Context, cfg map[string]string, msg Message) error
}

type Message struct {
	Title string
	Body  string
}

type Registry struct {
	channels map[ChannelType]Notifier
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		channels: make(map[ChannelType]Notifier),
	}
}

func (r *Registry) Register(t ChannelType, impl Notifier) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channels[t] = impl
}

func (r *Registry) Get(t ChannelType) (Notifier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	impl, ok := r.channels[t]
	if !ok {
		return nil, errors.New("unsupported notification channel type")
	}
	return impl, nil
}

func (r *Registry) IsEnabled(t ChannelType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.channels[t]
	return ok
}

var DefaultRegistry = NewRegistry()
