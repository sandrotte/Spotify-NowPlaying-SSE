package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SSEClient struct {
	id      int
	writer  http.ResponseWriter
	flusher http.Flusher
	done    <-chan struct{}
}

var (
	sseMutex    sync.Mutex
	clients      = map[int]*SSEClient{}
	nextClientID = 1
)

func handleSSE(responseWriter http.ResponseWriter, request *http.Request) {
	flusher, ok := responseWriter.(http.Flusher)

	if !ok {
		http.Error(responseWriter, "Streaming unsupported", http.StatusInternalServerError)

		return
	}

	responseWriter.Header().Set("Content-Type", "text/event-stream")
	responseWriter.Header().Set("Cache-Control", "no-store")
	responseWriter.Header().Set("Connection", "keep-alive")
	
	// If behind nginx, uncomment:
	// responseWriter.Header().Set("X-Accel-Buffering", "no")

	clientID := addClient(responseWriter, flusher, request.Context().Done())

	defer removeClient(clientID)

	writeSSEComment(responseWriter, flusher, "connected")

	<-request.Context().Done()
}

func broadcasterLoop(stop <-chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	keepAliveTicker := time.NewTicker(15 * time.Second)

	defer keepAliveTicker.Stop()

	for {
		select {
			case <-stop:
				return

			case <-ticker.C:
				data, err := getSpotifyData(context.Background())

				if err != nil || data == nil {
					continue
				}

				broadcastSpotifyData(data)

			case <-keepAliveTicker.C:
				broadcastKeepAlive()
		}
	}
}

func broadcastSpotifyData(data *SpotifyData) {
	payload, err := json.Marshal(data)
	
	if err != nil {
		return
	}

	sseMutex.Lock()

	defer sseMutex.Unlock()

	for _, client := range clients {
		select {
			case <-client.done:
				continue

			default:
		}

		writeSSEEvent(client.writer, client.flusher, "spotify", payload)
	}
}

func broadcastKeepAlive() {
	sseMutex.Lock()
	
	defer sseMutex.Unlock()

	for _, client := range clients {
		select {
			case <-client.done:
				continue
				
			default:
		}

		writeSSEComment(client.writer, client.flusher, "keep-alive")
	}
}

func writeSSEEvent(responseWriter http.ResponseWriter, flusher http.Flusher, eventName string, data []byte) {
	_, _ = fmt.Fprintf(responseWriter, "event: %s\n", eventName)
	_, _ = fmt.Fprintf(responseWriter, "data: %s\n\n", data)

	flusher.Flush()
}

func writeSSEComment(responseWriter http.ResponseWriter, flusher http.Flusher, comment string) {
	_, _ = fmt.Fprintf(responseWriter, ": %s\n\n", comment)

	flusher.Flush()
}

func addClient(responseWriter http.ResponseWriter, flusher http.Flusher, done <-chan struct{}) int {
	sseMutex.Lock()

	defer sseMutex.Unlock()

	id := nextClientID

	nextClientID++

	clients[id] = &SSEClient{
		id:      id,
		writer:  responseWriter,
		flusher: flusher,
		done:    done,
	}

	return id
}

func removeClient(id int) {
	sseMutex.Lock()

	defer sseMutex.Unlock()

	delete(clients, id)
}