package work

import (
	"delivery-slot-checker/internal/merchant"
	"errors"
	"fmt"
)

var AsdaCheckDeliverySlotsTask = Task(func(w TaskWriter) error {
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
		"Found %d available slots, from %s to %s\n",
		manifest.GetSlotCount(),
		manifest.GetFirstDate().Format("Mon 2 Jan"),
		manifest.GetLastDate().Format("Mon 2 Jan"),
	)

	return nil
})
