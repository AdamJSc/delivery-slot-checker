package supermarket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	asdaStatusAvailable   = "AVAILABLE"
	asdaStatusUnavailable = "UNAVAILABLE"
)

type AsdaClient struct {
	Client
}

func (c AsdaClient) GetDeliverySlots() ([]DeliverySlot, error) {
	var (
		response asdaResponse
	)

	contents, err := getDataFromCache("cached_success")
	if err != nil {
		return []DeliverySlot{}, err
	}

	if err = json.Unmarshal(contents, &response); err != nil {
		return []DeliverySlot{}, err
	}

	if response.StatusCode == asdaStatusUnavailable {
		return []DeliverySlot{}, ServiceUnavailableError{}
	}

	var slots []DeliverySlot
	for _, slotDay := range response.Data.SlotDays {
		for _, slot := range slotDay.Slots {
			slots = append(slots, slot.SlotInfo)
		}
	}

	return slots, nil
}

type AsdaDeliverySlot struct {
	DeliverySlot
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
}

func (ds AsdaDeliverySlot) IsAvailable() bool {
	return ds.Status == asdaStatusAvailable
}

func (ds AsdaDeliverySlot) GetTime() time.Time {
	return ds.StartTime
}

type asdaResponse struct {
	StatusCode string `json:"statusCode"`
	Data       struct {
		SlotDays []struct {
			Slots []struct {
				SlotInfo AsdaDeliverySlot `json:"slot_info"`
			} `json:"slots"`
		} `json:"slot_days"`
	} `json:"data"`
}

func getDataFromCache(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("./data/%s.txt", filename))
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}
