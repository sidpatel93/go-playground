package main


import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
	"fmt"
)

type Message struct {
	Content string `json:"content"`
}

type Broker struct {
	clients map[chan string]bool
	mu      sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan string]bool),
	}
}

var broker *Broker = NewBroker()

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

func PollingHandler(w http.ResponseWriter, r *http.Request) {
	messageChan := make(chan string, 1)
	broker.Subscribe(messageChan)
	defer broker.Unsubscribe(messageChan)

	select {
	case msg := <-messageChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{Content: msg})
	case <-time.After(30 * time.Second):
		http.Error(w, "Timeout", http.StatusRequestTimeout)
	case <-r.Context().Done():
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	broker.Broadcast(msg.Content)
	fmt.Fprintln(w, "Message broadcasted")
	w.WriteHeader(http.StatusNoContent)
}

func main() {

	http.HandleFunc("/poll", PollingHandler)
	http.HandleFunc("/post", PostHandler)
	http.ListenAndServe(":8080", nil)

}
