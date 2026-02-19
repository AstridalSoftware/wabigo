package helpers

import (
	"context"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func NotifyByWA(client *whatsmeow.Client, jid types.JID, text string) (bool, error) {

	msg := &waE2E.Message{
		Conversation: &text,
	}

	_, err := client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return false, err
	}
	return true, nil
}