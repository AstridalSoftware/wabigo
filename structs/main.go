package structs

import (
	"context"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type RequestImageMessage struct {
	ID       int
	WAClient *whatsmeow.Client
	Sender   types.JID
	Msg      *waE2E.ImageMessage
}

type WASenderRequest struct {
	ID      int
	Payload WASenderWebhookPayload
}

type Dispatcher struct {
	Jobs chan RequestImageMessage
	WG   sync.WaitGroup
	Ctx  context.Context
}

type WASenderDispatcher struct {
	Jobs chan WASenderRequest
	WG   sync.WaitGroup
	Ctx  context.Context
}

type WASenderWebhookPayload struct {
	Event     string `json:"event"`
	Timestamp int64  `json:"timestamp"`
	Data      Data   `json:"data"`
}

type Data struct {
	Messages MessageWrapper `json:"messages"`
}

type MessageWrapper struct {
	Key             MessageKey   `json:"key"`
	CleanedSenderPn string       `json:"cleanedSenderPn"`
	MessageBody     string       `json:"messageBody"`
	Message         InnerMessage `json:"message"`
}

type MessageKey struct {
	ID              string `json:"id"`
	FromMe          bool   `json:"fromMe"`
	RemoteJid       string `json:"remoteJid"`
	AddressingMode  string `json:"addressingMode"`
	SenderPn        string `json:"senderPn"`
	CleanedSenderPn string `json:"cleanedSenderPn"`
	SenderLid       string `json:"senderLid"`
}

type InnerMessage struct {
	Conversation string       `json:"conversation,omitempty"`
	ImageMessage ImageMessage `json:"imageMessage,omitempty"`
}

type ImageMessage struct {
	URL      string `json:"url"`
	FileName string `json:"fileName"`
	Mimetype string `json:"mimetype"`
	MediaKey string `json:"mediaKey"`
}

type WASenderDecryptMediaRequest struct{}

type WASenderDecryptMediaResponse struct {
	Success   bool   `json:"success"`
	PublicURL string `json:"publicUrl"`
}

type WASenderSendMessageResponse struct {
	Success bool `json:"success"`
	Data    struct {
		MsgID  int    `json:"msgId"`
		JID    string `json:"jid"`
		Status string `json:"status"`
	} `json:"data"`
}
