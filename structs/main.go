package structs

import (
	"context"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type RequestImageMessage struct {
	ID   int
	WAClient *whatsmeow.Client
	Sender types.JID
	Msg *waE2E.ImageMessage
}

type Dispatcher struct {
	Jobs chan RequestImageMessage
	WG   sync.WaitGroup
	Ctx  context.Context
}