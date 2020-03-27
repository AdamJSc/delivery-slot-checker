package main

import (
	"delivery-slot-checker/internal/supermarket"
	"log"
)

func main() {
	client := supermarket.NewClient()

	slots, err := client.GetDeliverySlots()
	if err != nil {
		switch err.(type) {
		case supermarket.ServiceUnavailableError:
			log.Fatal(err)
		default:
			log.Fatalf("unexpected error: %s", err)
		}
	}

	availableSlots := supermarket.FilterDeliverySlotsByAvailable(slots)
	manifest, err := supermarket.GetAvailabilityManifestFromSlots(client.GetChain(), availableSlots)
	if err != nil {
		log.Fatal(err)
	}

	if manifest.GetSlotCount() == 0 {
		log.Fatal("no available slots :(")
	}

	manifest.SortByDate(true)

	log.Printf("Delivery Manifest:\n%+v\n", manifest)
	log.Printf(
		"Found %d available slots, from %s to %s\n",
		manifest.GetSlotCount(),
		manifest.GetFirstDate().Format("Mon 2 Jan"),
		manifest.GetLastDate().Format("Mon 2 Jan"),
	)
}
