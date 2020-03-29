package work

import (
	"delivery-slot-checker/internal/merchant"
	"errors"
	"fmt"
	"time"
)

var AsdaCheckDeliverySlotsTask = Task(func(state *JobState, w WriterWithIdentifier) error {
	state.LatestRun = time.Now()

	if state.Bypass == true {
		return errors.New("bypassing job...")
	}

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

	fmt.Fprintf(
		w,
		"found %d available slots, from %s to %s\n",
		manifest.GetSlotCount(),
		manifest.GetFirstDate().Format("Mon 2 Jan"),
		manifest.GetLastDate().Format("Mon 2 Jan"),
	)

	state.Bypass = true

	return nil
})
