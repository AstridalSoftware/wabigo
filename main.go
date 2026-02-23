package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wabigo/helpers"
	"wabigo/structs"
	"wabigo/workers"

	_ "github.com/mattn/go-sqlite3" // Importar el driver SQLite
)

func main() {

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Crear dispatcher de workers
	dispatcher := workers.NewWASenderWorkerDispatcher(ctx, 2, 100) // 2 workers, buffer 100
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		payload, err := helpers.ParseWebhookPayload(r)
		if err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		dispatcher.Enqueue(structs.WASenderRequest{
			ID:      123,
			Payload: *payload,
		})

	})
	http.ListenAndServe(":8080", nil)

	// Esperar Ctrl+C para desconectar correctamente
	<-ctx.Done()
	fmt.Println("Shutting down...")
	dispatcher.Shutdown()
	fmt.Println("Cliente desconectado.")
}
