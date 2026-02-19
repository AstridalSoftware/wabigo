package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
	"wabigo/api"
	"wabigo/config"
	"wabigo/helpers"
	"wabigo/services"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3" // Importar el driver SQLite
	"github.com/mdp/qrterminal/v3"
)


func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

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
	client.AddEventHandler(func(evt any) {
		switch v := evt.(type) {
		case *events.Message:
			
			img := v.Message.GetImageMessage()
			if img == nil && cfg.AppEnv != "development" {
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

			text := img.GetCaption()
			fmt.Println("Texto recibido:", text)
			rgxCURP := regexp.MustCompile(`(?i)\b[A-Z][AEIOU][A-Z]{2}\d{6}[HM][A-Z]{2}[B-DF-HJ-NP-TV-Z]{3}[A-Z0-9]\d\b`)
			if matches := rgxCURP.FindStringSubmatch(text); len(matches) > 0 {

				data, err := client.Download(context.Background(), img)
				if err != nil {
					fmt.Println("Error descargando imagen:", err)
					return
				}

				// fileName := fmt.Sprintf("image_%d.jpg", time.Now().Unix())
				// err = os.WriteFile(fileName, data, 0644)
				base64Image := base64.StdEncoding.EncodeToString(data)

				issuesService := &services.IssuesClient{Client:api.NewHttpClient(cfg.API_BASE_URL)}
				_, err = issuesService.AddIssue(context.Background(), 
					&services.IssueCreateRequest{
					Title:       v.Info.Sender.ToNonAD().User,
					Description: strings.ToUpper(text),
					ImageB64: base64Image,
				})
				if err != nil {
					fmt.Printf("Error obteniendo issues: %v\n", err)
				}
				fmt.Printf("SE ENVIO REQUEST")
				randomTimeOut := rand.IntN(3) + 1
				time.Sleep(time.Duration(randomTimeOut) * time.Second)
				_, err = helpers.NotifyByWA(
					client,
					v.Info.Sender.ToNonAD(),
					"Solicitud de contraseña recibida",
				)
				if err != nil {
					fmt.Printf("Error enviando notificación: %v\n", err)
				}
				fmt.Printf("Se envio notificacion mensaje")
				return
			}else{
				fmt.Printf("NO MATCH CURP")
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
