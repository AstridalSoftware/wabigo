package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wabigo/api"
	"wabigo/config"
	"wabigo/structs"
)

var cfg = config.Load()
var WASenderAPI *api.Client = api.CreateHttpClient(cfg.WASENDER_API_BASE_URL, cfg.WASENDER_API_KEY)

func IsWebhookSignatureValid(r *http.Request) bool {
	if r.Header.Get("X-Webhook-Signature") == "" {
		return false
	}

	if r.Header.Get("X-Webhook-Signature") != cfg.WEBHOOK_SECRET {
		return false
	}
	// Aquí iría la lógica de validación usando el secret
	return true
}

func ParseWebhookPayload(r *http.Request) (*structs.WASenderWebhookPayload, error) {
	var payload structs.WASenderWebhookPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func WASenderDownloadMedia(ctx context.Context, msg *structs.WASenderWebhookPayload) ([]byte, error) {

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"messages": map[string]interface{}{
				"key": map[string]interface{}{
					"id": msg.Data.Messages.Key.ID,
				},
				"message": map[string]interface{}{
					"imageMessage": map[string]interface{}{
						"url":      msg.Data.Messages.Message.ImageMessage.URL,
						"mimetype": msg.Data.Messages.Message.ImageMessage.Mimetype,
						"mediaKey": msg.Data.Messages.Message.ImageMessage.MediaKey,
					},
				},
			},
		},
	}
	jsonPayload, _ := json.Marshal(payload)

	res, err := WASenderAPI.DoPostRequest(ctx, "/decrypt-media", cfg.WASENDER_API_KEY, jsonPayload)
	if err != nil {
		return nil, err
	}
	var apiResponse structs.WASenderDecryptMediaResponse
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}
	res.Body.Close()
	fmt.Printf("URL de descarga: %s\n", apiResponse.PublicURL+"/"+msg.Data.Messages.Key.ID)
	fileResponse, err := WASenderAPI.DoGetRequest(ctx, apiResponse.PublicURL+"/"+msg.Data.Messages.Key.ID, cfg.WASENDER_API_KEY)
	if err != nil {
		return nil, err
	}
	file, err := io.ReadAll(fileResponse.Body)
	if err != nil {
		return nil, err
	}
	fileResponse.Body.Close()
	return file, nil
}

func WASenderSendMessage(ctx context.Context, to string, text string) (*structs.WASenderSendMessageResponse, error) {

	payloadMap := map[string]interface{}{
		"to":   to,
		"text": text,
	}
	payload, _ := json.Marshal(payloadMap)

	req, err := WASenderAPI.DoPostRequest(ctx, "/send-message", cfg.WASENDER_API_KEY, payload)
	if err != nil {
		return nil, err
	}

	var apiResponse structs.WASenderSendMessageResponse
	if err := json.NewDecoder(req.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}
	req.Body.Close()
	return &apiResponse, nil
}
