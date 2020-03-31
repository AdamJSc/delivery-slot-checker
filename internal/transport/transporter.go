package transport

import "os"

// Message represent a message that we can send via SMS
type Message struct {
	From string
	To   string
	Text string
}

// Transporter represent
type Transporter interface {
	SendSMS(message Message) error
}

// NewTransporter instantiates the default Transporter
func NewTransporter() Transporter {
	return NexmoTransporter{
		apiKey:    os.Getenv("NEXMO_KEY"),
		apiSecret: os.Getenv("NEXMO_SECRET"),
	}
}
