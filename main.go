package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
        log.Printf("No .env file found")
    }

	port := 9700

	mux := http.NewServeMux()

	mux.HandleFunc("/", withCORS(handleRootJSON))
	mux.HandleFunc("/sse", withCORS(handleSSE))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	stopBroadcaster := make(chan struct{})

	go broadcasterLoop(stopBroadcaster, 1*time.Second)

	log.Printf("SSE started on port: %d\n", port)

	shutdownSignals := make(chan os.Signal, 1)

	signal.Notify(shutdownSignals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownSignals

		close(stopBroadcaster)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		_ = server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("An error occurred while starting the server: %v", err)
	}
}

func handleRootJSON(responseWriter http.ResponseWriter, request *http.Request) {
	data, err := getSpotifyData(request.Context())

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadGateway)
		
		return
	}
	
	if data == nil {
		responseWriter.Header().Set("Content-Type", "application/json")
	
		responseWriter.WriteHeader(http.StatusOK)
	
		_, _ = responseWriter.Write([]byte("null"))
	
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(responseWriter).Encode(data)
}