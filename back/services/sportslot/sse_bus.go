package sportslot

import (
	"sync"
)

type EventType string

const (
	EventCategoriaUpsert EventType = "categoria"
	EventAmbienteUpsert  EventType = "ambiente"
	EventAmbienteInfo    EventType = "ambiente_info"
	EventAmbienteOcup     EventType = "ambiente_ocupaciones"
)

type SportSlotTopic string

const (
	TopicCategorias SportSlotTopic = "categorias"
	TopicAmbientes  SportSlotTopic = "ambientes"
	TopicOcupaciones SportSlotTopic = "ocupaciones"
)

type SSEBus struct {
	mu      sync.RWMutex
	clients map[chan SportSlotSSE]struct{}
}

type SportSlotSSE struct {
	Topic SportSlotTopic `json:"topic"`
	Event string         `json:"event"`
	Data  any            `json:"data"`
}

func NewSSEBus() *SSEBus {
	return &SSEBus{clients: make(map[chan SportSlotSSE]struct{})}
}

func (b *SSEBus) Register() chan SportSlotSSE {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan SportSlotSSE, 64)
	b.clients[ch] = struct{}{}
	return ch
}

func (b *SSEBus) Unregister(ch chan SportSlotSSE) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, ch)
	close(ch)
}

func (b *SSEBus) Broadcast(evt SportSlotSSE) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- evt:
		default:
		}
	}
}
