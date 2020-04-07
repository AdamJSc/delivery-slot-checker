package work

import (
	"delivery-slot-checker/domain/apperrors"
	"delivery-slot-checker/domain/merchant"
	"delivery-slot-checker/domain/transport"
	"errors"
	"fmt"
	"strings"
	"time"
)

var AsdaDeliverySlotsTask = Task(func(payload TaskPayload, state *TaskState, w WriterWithIdentifier) error {
	return checkForDeliverySlots(merchant.AsdaClient{}, payload, state, w)
})

func checkForDeliverySlots(client merchant.Client, payload TaskPayload, state *TaskState, w WriterWithIdentifier) error {
	state.LatestRun = time.Now()

	if state.Bypass {
		return errors.New("bypassing task...")
	}

	now := time.Now()
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return apperrors.FatalError{Err: err}
	}

	todayAtMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tsInSevenDays := todayAtMidnight.Add(7 * 24 * time.Hour)
	tsInTwentyOneDays := todayAtMidnight.Add(22 * 24 * time.Hour).Add(-time.Second)

	slots, err := client.GetDeliverySlots(payload.Postcode, tsInSevenDays, tsInTwentyOneDays)
	if err != nil {
		return err
	}

	manifest, err := merchant.NewDeliveryManifest(client.GetName(), slots)
	if err != nil {
		return err
	}
	manifest.FilterByAvailability(true)

	if manifest.GetSlotCount() == 0 {
		return errors.New("no available slots :(")
	}

	manifest.SortByDate(true)

	from := manifest.From.Format("Mon 2 Jan")
	until := manifest.Until.Format("Mon 2 Jan")
	fmt.Fprintf(
		w,
		"found %d available slots, from %s until %s\n",
		manifest.GetSlotCount(),
		strings.ToLower(from),
		strings.ToLower(until),
	)

	transporter := transport.NewTransporter()
	for _, recipient := range payload.Recipients {
		message := transport.Message{
			From: "SlotChecker",
			To:   recipient.Mobile,
			Text: manifest.AsMessageText(recipient.Name),
		}

		if err = transporter.SendSMS(message); err != nil {
			return fmt.Errorf("error sending sms: %s", err)
		}
	}

	state.Bypass = true

	return nil
}
