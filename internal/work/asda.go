package work

import (
	"delivery-slot-checker/internal/supermarket"
	"errors"
	"log"
)

var AsdaCheckDeliverySlotsTask = Task(func(l *log.Logger) error {
	client := supermarket.NewClient()

	slots, err := client.GetDeliverySlots()
	if err != nil {
		return err
	}

	availableSlots := supermarket.FilterDeliverySlotsByAvailable(slots)
	manifest, err := supermarket.GetAvailabilityManifestFromSlots(client.GetChain(), availableSlots)
	if err != nil {
		return err
	}

	if manifest.GetSlotCount() == 0 {
		return errors.New("no available slots :(")
	}

	manifest.SortByDate(true)

	l.Printf(
		"Found %d available slots, from %s to %s\n",
		manifest.GetSlotCount(),
		manifest.GetFirstDate().Format("Mon 2 Jan"),
		manifest.GetLastDate().Format("Mon 2 Jan"),
	)

	return nil
})
