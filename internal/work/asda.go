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

type checkDeliverySlotsTaskData struct {
	Postcode   string `yaml:"postcode"`
	Recipients []struct {
		Name   string `yaml:"name"`
		Mobile string `yaml:"mobile"`
	} `yaml:"recipients"`
}

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

	var taskData []checkDeliverySlotsTaskData

	err = yaml.Unmarshal(taskDataFileContents, &taskData)
	if err != nil {
		return apperrors.FatalError{Err: err}
	}

	// begin tasks...
	for _, data := range taskData {
		err = checkForDeliverySlots(data, state, w)
		if err != nil {
			return err
		}
	}

	return nil
})

func checkForDeliverySlots(data checkDeliverySlotsTaskData, state *JobState, w WriterWithIdentifier) error {
	client := merchant.NewClient()

	now := time.Now()
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return apperrors.FatalError{Err: err}
	}

	todayAtMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tsInSevenDays := todayAtMidnight.Add(7 * 24 * time.Hour)
	tsInTwentyOneDays := todayAtMidnight.Add(22 * 24 * time.Hour).Add(-time.Second)

	slots, err := client.GetDeliverySlots(data.Postcode, tsInSevenDays, tsInTwentyOneDays)
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
	for _, recipient := range data.Recipients {
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
