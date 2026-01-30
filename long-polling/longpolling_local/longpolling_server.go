package longpolling

import (
	"encoding/json"
	"fmt"
	"long-polling/common"
	"net/http"
	"time"
)

var broker *common.Broker = common.NewBroker()

func PollingHandler(w http.ResponseWriter, r *http.Request) {
	messageChan := make(chan string, 1)
	broker.Subscribe(messageChan)
	defer broker.Unsubscribe(messageChan)

	select {
	case msg := <-messageChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.Message{Content: msg})
	case <-time.After(30 * time.Second):
		http.Error(w, "Timeout", http.StatusRequestTimeout)
	case <-r.Context().Done():
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	var msg common.Message
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
