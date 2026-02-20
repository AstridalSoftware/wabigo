package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"wabigo/config"
	"wabigo/helpers"
	"wabigo/structs"
	"wabigo/workers"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3" // Importar el driver SQLite
	"github.com/mdp/qrterminal/v3"
)


func main() {

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
    defer stop()
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Crear logger para la base de datos
	dbLog := waLog.Stdout("Database", "DEBUG", true)

	// Inicializar SQLite store
	container, err := sqlstore.New(ctx, "sqlite3", "file:example.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	// Obtener el primer dispositivo registrado (si no existe, se creará con QR)
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	// Logger para el cliente
	clientLog := waLog.Stdout("Client", "DEBUG", true)

	// Crear cliente de WhatsApp
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Crear dispatcher de workers
	dispatcher := workers.NewSimpleWorkerDispatcher(ctx, 2, 100) // 2 workers, buffer 100

	// Agregar handler de eventos
	client.AddEventHandler(func(evt any) {
		switch v := evt.(type) {
		case *events.Message:
			
			msg := v.Message.GetImageMessage()
			if msg == nil && cfg.AppEnv != "development" {
				_, err := helpers.NotifyByWA(
					client,
					v.Info.Sender.ToNonAD(),
					"¡Hola! Para recuperar tu contraseña, por favor manda una foto de tu INE y en el mensaje escribe tu CURP",
				)
				if err != nil {
					fmt.Printf("Error enviando notificación: %v\n", err)
				}
				return
			}

			dispatcher.Enqueue(structs.RequestImageMessage{
					ID:         123,
					WAClient: client,
					Sender: v.Info.Sender,
					Msg:	   msg,
			})

		}
	})

	// Conectar cliente
	if client.Store.ID == nil {
		// No hay sesión -> login nuevo con QR
		qrChan, _ := client.GetQRChannel(ctx)
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				// Renderizar QR en consola
				qrterminal.Generate(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("Escanea este QR con WhatsApp Web")
			} else {
				fmt.Println("Evento de login:", evt.Event)
			}
		}

	} else {
		// Ya hay sesión -> solo conectar
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Esperar Ctrl+C para desconectar correctamente
	<-ctx.Done()

    fmt.Println("Shutting down...")

	dispatcher.Shutdown()
    client.Disconnect()

    fmt.Println("Cliente desconectado.")
}
