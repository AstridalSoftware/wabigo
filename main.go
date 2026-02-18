package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3" // Importar el driver SQLite
	"github.com/mdp/qrterminal/v3"
)

func sendMessage(client *whatsmeow.Client, jid types.JID, text string) {
	msg := &waE2E.Message{
		Conversation: &text,
	}
	_, err := client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		fmt.Printf("Error enviando mensaje: %v\n", err)
	}
}

func main() {

	ctx := context.Background()

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

	// Agregar handler de eventos
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// Ignorar mensajes sin contenido
			if v.Message == nil {
				return
			}

			// Ignorar mensajes enviados por el propio bot
			if v.Info.IsFromMe {
				return
			}

			// Obtener texto del mensaje (normal o extendido)
			var text string
			if v.Message.GetConversation() != "" {
				text = v.Message.GetConversation()
			} else if v.Message.GetExtendedTextMessage() != nil {
				text = v.Message.GetExtendedTextMessage().GetText()
			} else {
				// Mensaje no texto (imagen, audio, sticker)
				fmt.Printf("Mensaje de %s recibido pero no es texto\n", v.Info.Sender.String())
				return
			}

			sender := v.Info.Sender.String()
			fmt.Printf("Mensaje de %s: %s\n", sender, text)

			// 🔹 Detectar @NombreNegocio
			reNegocio := regexp.MustCompile(`^@(\w+)`)
			if matches := reNegocio.FindStringSubmatch(text); len(matches) > 1 {
				negocio := matches[1]
				reply := fmt.Sprintf("Has seleccionado el negocio: %s", negocio)
				fmt.Println("PETICION DE NEGOCIO")
				sendMessage(client, v.Info.Sender, reply)
				return
			}

			// 🔹 Detectar pedidos #IDxCantidad
			rePedido := regexp.MustCompile(`#(\d+)x(\d+)`)
			pedidos := rePedido.FindAllStringSubmatch(text, -1)
			if len(pedidos) > 0 {
				reply := "Has ordenado:\n"
				for _, p := range pedidos {
					itemID := p[1]
					qty := p[2]
					reply += fmt.Sprintf("- Item #%s x%s\n", itemID, qty)
				}
				fmt.Println("PETICION DE PEDIDO")
				sendMessage(client, v.Info.Sender, reply)
				return
			}
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
	fmt.Println("Cliente desconectado.")
}
