package workers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand/v2"
	"regexp"
	"strings"
	"time"
	"wabigo/api"
	"wabigo/config"
	"wabigo/helpers"
	"wabigo/services"
	"wabigo/structs"
)

type WASenderWorkerDispatcher struct {
	*structs.WASenderDispatcher
}

func NewWASenderWorkerDispatcher(ctx context.Context, workerCount int, bufferSize int) *WASenderWorkerDispatcher {
	ctx, cancel := context.WithCancel(ctx)
	d := &WASenderWorkerDispatcher{
		WASenderDispatcher: &structs.WASenderDispatcher{
			Jobs:   make(chan structs.WASenderRequest, bufferSize),
			Ctx:    ctx,
			Cancel: cancel,
		},
	}

	for i := 1; i <= workerCount; i++ {
		d.WG.Add(1)
		go d.worker(i)
	}

	return d
}

func (d *WASenderWorkerDispatcher) worker(id int) {

	cfg := config.Load()
	DnetSoporteAPI := api.CreateHttpClient(cfg.DNET_SOPORTE_API_BASE_URL, cfg.API_KEY)

	defer d.WG.Done()

	for {
		select {
		case req, ok := <-d.Jobs:
			if !ok {
				log.Printf("Worker %d drained y terminado\n", id)
				return
			}
			// Lógica de negocio + llamadas a APIs
			fmt.Printf("Worker %d procesando mensaje %d\n", id, req.ID)
			text := req.Payload.Data.Messages.MessageBody
			fmt.Println("Texto recibido:", text)
			rgxCURP := regexp.MustCompile(`(?i)\b[A-Z][AEIOU][A-Z]{2}\d{6}[HM][A-Z]{2}[B-DF-HJ-NP-TV-Z]{3}[A-Z0-9]\d\b`)
			if matches := rgxCURP.FindStringSubmatch(text); len(matches) > 0 {

				data, err := helpers.WASenderDownloadMedia(d.Ctx, &req.Payload)
				if err != nil {
					fmt.Println("Error descargando imagen:", err)
					break
				}

				base64Image := base64.StdEncoding.EncodeToString(data)

				issuesService := services.NewIssuesService(DnetSoporteAPI)
				_, err = issuesService.AddIssue(d.Ctx,
					&services.IssueCreateRequest{
						Title:       req.Payload.Data.Messages.CleanedSenderPn,
						Description: strings.ToUpper(text),
						ImageB64:    base64Image,
					})
				if err != nil {
					fmt.Printf("Error obteniendo issues: %v\n", err)
					break
				}

				fmt.Printf("SE ENVIO REQUEST")

				randomTimeOut := rand.IntN(3) + 1
				time.Sleep(time.Duration(randomTimeOut) * time.Second)
				_, err = helpers.WASenderSendMessage(
					d.Ctx,
					req.Payload.Data.Messages.Key.RemoteJid,
					"Solicitud de contraseña recibida",
				)
				if err != nil {
					fmt.Printf("Error enviando notificación: %v\n", err)
					break
				}
				fmt.Printf("Se envio notificacion mensaje")
			}

		case <-d.Ctx.Done():
			log.Printf("Worker %d detenido\n", id)
			return
		}
	}
}

func (d *WASenderWorkerDispatcher) Enqueue(msg structs.WASenderRequest) {
	select {
	case d.Jobs <- msg:
		// Encolado exitoso
	default:
		// Buffer lleno → estrategia de backpressure
		log.Println("Cola llena, descartando mensaje o aplicando retry")
	}
}

func (d *WASenderWorkerDispatcher) Abort() {
	d.Cancel()
	d.WG.Wait()
}

func (d *WASenderWorkerDispatcher) Drain() {
	close(d.Jobs)
	d.WG.Wait()
}
