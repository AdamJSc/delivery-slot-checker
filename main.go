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

	availableSlots := supermarket.FilterAvailableDeliverySlots(slots)
	manifest, err := supermarket.GetDeliveryManifestFromSlots(availableSlots)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Delivery Manifest:\n%+v\n", manifest)
	log.Printf("Found %d available slots\n", len(availableSlots))
}
