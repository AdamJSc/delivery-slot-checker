package transport

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/nexmo-community/nexmo-go"
)

type NexmoTransporter struct {
	Transporter
	apiKey    string
	apiSecret string
}

func (n NexmoTransporter) SendSMS(message Message) error {
	auth := nexmo.NewAuthSet()
	auth.SetAPISecret(n.apiKey, n.apiSecret)

	client := nexmo.NewClient(http.DefaultClient, auth)
	smsReq := nexmo.SendSMSRequest{
		From: message.From,
		To:   message.To,
		Text: message.Text,
	}

	resp, _, err := client.SMS.SendSMS(smsReq)
	if err != nil {
		return err
	}

	if resp.MessageCount == "0" {
		return errors.New("no messages returned")
	}

	for _, message := range resp.Messages {
		if message.Status != "0" {
			return fmt.Errorf("%s (status %s)", strings.ToLower(message.ErrorText), message.Status)
		}
	}

	return nil
}
