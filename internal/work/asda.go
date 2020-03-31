package work

import (
	"delivery-slot-checker/internal/apperrors"
	"delivery-slot-checker/internal/merchant"
	"delivery-slot-checker/internal/transport"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"time"
)

var AsdaCheckDeliverySlotsTask = Task(func(state *JobState, w WriterWithIdentifier) error {
	state.LatestRun = time.Now()

	if state.Bypass == true {
		return errors.New("bypassing job...")
	}

	// retrieve and parse recipients data
	taskDataFileContents, err := ioutil.ReadFile("./data/tasks/asda-check-delivery-slots.yml")
	if err != nil {
		return apperrors.FatalError{Err: err}
	}

	var taskData struct {
		Recipients []struct {
			Name   string `yaml:"name"`
			Mobile string `yaml:"mobile"`
		} `yaml:"recipients"`
	}

	err = yaml.Unmarshal(taskDataFileContents, &taskData)
	if err != nil {
		return apperrors.FatalError{Err: err}
	}

	// begin task...

	client := merchant.NewClient()

	slots, err := client.GetDeliverySlots()
	if err != nil {
		return err
	}

	availableSlots := merchant.FilterDeliverySlotsByAvailability(slots, true)
	manifest, err := merchant.GetAvailabilityManifestFromSlots(client.GetName(), availableSlots)
	if err != nil {
		return err
	}

	if manifest.GetSlotCount() == 0 {
		return errors.New("no available slots :(")
	}

	manifest.SortByDate(true)

	from := manifest.GetFirstDate().Format("Mon 2 Jan")
	to := manifest.GetLastDate().Format("Mon 2 Jan")
	fmt.Fprintf(
		w,
		"found %d available slots, from %s to %s\n",
		manifest.GetSlotCount(),
		strings.ToLower(from),
		strings.ToLower(to),
	)

	transporter := transport.NewTransporter()
	for _, recipient := range taskData.Recipients {
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
})
