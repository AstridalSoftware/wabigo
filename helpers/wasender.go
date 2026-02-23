package helpers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"wabigo/api"
	"wabigo/config"
	"wabigo/structs"
)

func ParseWebhookPayload(r *http.Request) (*structs.WASenderWebhookPayload, error) {
	var payload structs.WASenderWebhookPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

var cfg = config.Load()
var WASenderAPI *api.Client = api.CreateHttpClient(cfg.WASENDER_API_BASE_URL)

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

	res, err := WASenderAPI.DoPostRequest(ctx, "/decrypt-media", jsonPayload)
	if err != nil {
		return nil, err
	}
	var apiResponse structs.WASenderDecryptMediaResponse
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	fileResponse, err := WASenderAPI.DoGetRequest(ctx, apiResponse.PublicURL+"/"+msg.Data.Messages.Key.ID)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(fileResponse.Body)
}

func WASenderSendMessage(ctx context.Context, to string, text string) (*structs.WASenderSendMessageResponse, error) {

	payloadMap := map[string]interface{}{
		"to":   to,
		"text": text,
	}
	payload, _ := json.Marshal(payloadMap)

	req, err := WASenderAPI.DoPostRequest(ctx, "/send-message", payload)
	if err != nil {
		panic(err)
	}
	var apiResponse structs.WASenderSendMessageResponse
	if err := json.NewDecoder(req.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}
	return &apiResponse, nil
}
