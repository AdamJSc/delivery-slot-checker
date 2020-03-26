package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

const asdaStatusAvailable = "AVAILABLE"
const asdaStatusUnavailable = "UNAVAILABLE"

type asdaDeliverySlot struct {
	Status string `json:"status"`
	StartTime time.Time `json:"start_time"`
}

func (ds asdaDeliverySlot) IsAvailable() bool {
	return ds.Status == asdaStatusAvailable
}

type asdaResponse struct {
	StatusCode string `json:"statusCode"`
	Data struct {
		SlotDays []struct {
			Slots []struct {
				SlotInfo asdaDeliverySlot `json:"slot_info"`
			} `json:"slots"`
		} `json:"slot_days"`
	} `json:"data"`
}

func main() {
	contents := getDataFromCache("cached_success")
	var response asdaResponse

	err := json.Unmarshal(contents, &response)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode == asdaStatusUnavailable {
		log.Fatal("site is down")
	}

	var availableSlots []asdaDeliverySlot
	for _, slotDay := range response.Data.SlotDays {
		for _, slot := range slotDay.Slots {
			if slot.SlotInfo.IsAvailable() {
				availableSlots = append(availableSlots, slot.SlotInfo)
			}
		}
	}

	log.Printf("%+v\nFound %d available slots", availableSlots, len(availableSlots))
}

func getDataFromCache(filename string) []byte {
	data, err := ioutil.ReadFile(fmt.Sprintf("./data/%s.txt", filename))
	if err != nil {
		log.Fatal(err)
	}

	return data
}
