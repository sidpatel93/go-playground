package longpolling_redis

import (
	"context"
	"encoding/json"
	"log"
	"long-polling/common"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var localBroker *common.Broker = common.NewBroker()

var ctx = context.Background()

const (
	RedisChannel = "longpolling_channel"
	ServerPort   = ":8080"
)

var rdb *redis.Client

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "192.168.2.156:6379",
		Username: "default",
		Password: "Periscope-Demeanor-Refold8-Preorder-Mutate",
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	log.Println("Connected to Redis")
}

func startRedisListener() {
	// Subscribe to the global topic
	pubsub := rdb.Subscribe(ctx, RedisChannel)

	// Get the Go channel for Redis messages
	ch := pubsub.Channel()

	log.Println("Listening for Redis updates...")

	// This loop runs forever in the background
	for msg := range ch {
		// When Redis tells us something happened...
		// We tell our local broker to wake up all waiting HTTP requests.
		localBroker.Broadcast(msg.Payload)
	}
}

func PollingHandler(w http.ResponseWriter, r *http.Request) {
	messageChan := make(chan string, 1)
	localBroker.Subscribe(messageChan)
	defer localBroker.Unsubscribe(messageChan)

	select {
	case msg := <-messageChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.Message{Content: msg})
	case <-time.After(10 * time.Second):
		http.Error(w, "Timeout", http.StatusRequestTimeout)
	case <-r.Context().Done():
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
	}
}

func PostHandlerRedis(w http.ResponseWriter, r *http.Request) {
	var msg common.Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// Publish the message to Redis
	err = rdb.Publish(ctx, RedisChannel, msg.Content).Err()
	if err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}
	log.Default().Println("Message broadcasted via Redis")
	w.WriteHeader(http.StatusNoContent)
}

func StartLongPollingrRedisServer() {
	initRedis()
	go startRedisListener()

	http.HandleFunc("/poll", PollingHandler)
	http.HandleFunc("/post", PostHandlerRedis)
	log.Printf("Starting server on port %s\n", ServerPort)
	if err := http.ListenAndServe(ServerPort, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
