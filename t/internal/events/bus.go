// Package events defines a simple event handler system
package events

import "sync"

type Handler func(event any)

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

func (b *Bus) On(eventType any, h Handler) {
	key := typeKey(eventType)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[key] = append(b.handlers[key], h)
}

func (b *Bus) Emit(event any) {
	key := typeKey(event)

	b.mu.RLock()
	handlers := append([]Handler{}, b.handlers[key]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		h(event)
	}
}
