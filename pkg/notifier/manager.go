package notifier

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages all registered notification channels
type Manager struct {
	notifiers map[string]Notifier
	mu        sync.RWMutex
}

// NewManager creates a new notification manager
func NewManager() *Manager {
	return &Manager{
		notifiers: make(map[string]Notifier),
	}
}

// Register registers a new notification channel
func (m *Manager) Register(notifier Notifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := notifier.Name()
	if _, exists := m.notifiers[name]; exists {
		return fmt.Errorf("notifier %s already registered", name)
	}

	m.notifiers[name] = notifier
	return nil
}

// Send sends a notification through the specified channel
func (m *Manager) Send(ctx context.Context, channel string, msg *Message) (*Response, error) {
	m.mu.RLock()
	notifier, exists := m.notifiers[channel]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("notifier %s not found", channel)
	}

	if err := notifier.Validate(msg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return notifier.Send(ctx, msg)
}

// Broadcast sends a notification through multiple channels
func (m *Manager) Broadcast(ctx context.Context, channels []string, msg *Message) map[string]*Response {
	results := make(map[string]*Response)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, channel := range channels {
		wg.Add(1)
		go func(ch string) {
			defer wg.Done()
			resp, err := m.Send(ctx, ch, msg)
			mu.Lock()
			if err != nil {
				results[ch] = &Response{
					Success: false,
					Error:   err.Error(),
				}
			} else {
				results[ch] = resp
			}
			mu.Unlock()
		}(channel)
	}

	wg.Wait()
	return results
}

// ListChannels returns a list of all registered channels
func (m *Manager) ListChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channels := make([]string, 0, len(m.notifiers))
	for name := range m.notifiers {
		channels = append(channels, name)
	}
	return channels
}

// Get returns a notifier by name
func (m *Manager) Get(channel string) (Notifier, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	notifier, exists := m.notifiers[channel]
	if !exists {
		return nil, fmt.Errorf("notifier %s not found", channel)
	}
	return notifier, nil
}
