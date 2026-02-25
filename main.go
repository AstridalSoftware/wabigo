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
	"wabigo/helpers"
	"wabigo/structs"
	"wabigo/workers"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// Canal explícito para señales
	sigChan := make(chan os.Signal, 1)

	signal.Notify(
		sigChan,
		os.Interrupt,    // Ctrl+C
		syscall.SIGTERM, // kill
		syscall.SIGQUIT, // kill -3
	)

	dispatcher := workers.NewWASenderWorkerDispatcher(context.Background(), 2, 100)

	mux := http.NewServeMux()

	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("WEBHOOK RECEIVED")

		if !helpers.IsWebhookSignatureValid(r) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		payload, err := helpers.ParseWebhookPayload(r)
		if err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		b, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(b))
		fmt.Println("WEBHOOK ENQUEUED")
		dispatcher.Enqueue(structs.WASenderRequest{
			ID:      uuid.New(),
			Payload: payload,
		})

	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		fmt.Println("STARTING SERVER ON :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()

	sig := <-sigChan
	fmt.Printf("Signal received: %v\n", sig)

	switch sig {

	case os.Interrupt, syscall.SIGTERM:
		fmt.Println("Starting graceful drain...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown error: %v\n", err)
		}

		dispatcher.Drain()

	case syscall.SIGQUIT:
		fmt.Println("Immediate abort...")

		dispatcher.Abort()
	}

	fmt.Println("Proceso finalizado.")
}
