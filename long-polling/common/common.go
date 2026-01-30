package common

import (
	"sync"
)

type Broker struct {
	clients map[chan string]bool
	mu      sync.Mutex
}

type Message struct {
	Content string `json:"content"`
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan string]bool),
	}
}

func (b *Broker) Subscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[ch] = true
}

func (b *Broker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, ch)
	close(ch)
}

func (b *Broker) Broadcast(message string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		ch <- message
	}
}
