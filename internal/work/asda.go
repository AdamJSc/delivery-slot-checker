package work

import (
	"delivery-slot-checker/internal/merchant"
	"errors"
	"fmt"
	"time"
)

var AsdaCheckDeliverySlotsTask = Task(func(w TaskWriter) error {
	now := time.Now()
	stateName := fmt.Sprintf("%s_%s", w.GetFormattedTaskName(), now.Format("20060102"))

	state, err := LoadStateAndCreateIfMissing(stateName)
	if err != nil {
		return err
	}

	state.LatestRun = now

	if state.Bypass == true {
		if err = SaveState(stateName, state); err != nil {
			return err
		}
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

	err = SaveState(stateName, state)
	if err != nil {
		return err
	}

	return nil
})
