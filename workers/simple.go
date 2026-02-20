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

type SimpleWorkerDispatcher struct {
	*structs.Dispatcher
}

func NewSimpleWorkerDispatcher(ctx context.Context, workerCount int, bufferSize int) *SimpleWorkerDispatcher {
	d := &SimpleWorkerDispatcher{
		Dispatcher: &structs.Dispatcher{
			Jobs: make(chan structs.RequestImageMessage, bufferSize),
			Ctx:  ctx,
		},
	}

	for i := 1; i <= workerCount; i++ {
		d.WG.Add(1)
		go d.worker(i)
	}

	return d
}

func (d *SimpleWorkerDispatcher) worker(id int) {
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	defer d.WG.Done()

	for {
		select {
		case req := <-d.Jobs:
			// Lógica de negocio + llamadas a APIs
			fmt.Printf("Worker %d procesando mensaje %d\n", id, req.ID)
			text := req.Msg.GetCaption()
			fmt.Println("Texto recibido:", text)
			rgxCURP := regexp.MustCompile(`(?i)\b[A-Z][AEIOU][A-Z]{2}\d{6}[HM][A-Z]{2}[B-DF-HJ-NP-TV-Z]{3}[A-Z0-9]\d\b`)
			if matches := rgxCURP.FindStringSubmatch(text); len(matches) > 0 {

				// En tu event handler:
				data, err := req.WAClient.Download(context.Background(), req.Msg)
				if err != nil {
					fmt.Println("Error descargando imagen:", err)
					break
				}

				base64Image := base64.StdEncoding.EncodeToString(data)

				issuesService := &services.IssuesClient{Client:api.NewHttpClient(cfg.API_BASE_URL)}
				_, err = issuesService.AddIssue(context.Background(), 
					&services.IssueCreateRequest{
					Title:       req.Sender.ToNonAD().User,
					Description: strings.ToUpper(text),
					ImageB64: base64Image,
				})
				if err != nil {
					fmt.Printf("Error obteniendo issues: %v\n", err)
					break
				}
				
				fmt.Printf("SE ENVIO REQUEST")

				randomTimeOut := rand.IntN(3) + 1
				time.Sleep(time.Duration(randomTimeOut) * time.Second)
				_, err = helpers.NotifyByWA(
					req.WAClient,
					req.Sender.ToNonAD(),
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

func (d *SimpleWorkerDispatcher) Enqueue(msg structs.RequestImageMessage) {
	select {
	case d.Jobs <- msg:
		// Encolado exitoso
	default:
		// Buffer lleno → estrategia de backpressure
		log.Println("Cola llena, descartando mensaje o aplicando retry")
	}
}

func (d *SimpleWorkerDispatcher) Shutdown() {
	d.WG.Wait()
}
