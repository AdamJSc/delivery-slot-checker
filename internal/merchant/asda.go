package merchant

import (
	"delivery-slot-checker/internal/apperrors"
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

func(c AsdaClient) GetName() string {
	return "Asda"
}

func (c AsdaClient) GetDeliverySlots() ([]DeliverySlot, error) {
	var (
		response asdaAPIResponse
	)

	contents, err := getDataFromCache("cached_success")
	if err != nil {
		return []DeliverySlot{}, apperrors.FatalError{Err: err}
	}

	if err = json.Unmarshal(contents, &response); err != nil {
		return []DeliverySlot{}, err
	}

	if response.StatusCode == asdaStatusUnavailable {
		return []DeliverySlot{}, apperrors.OfflineError(c.GetName())
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

type asdaAPIResponse struct {
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
	data, err := ioutil.ReadFile(fmt.Sprintf("./data/cache/%s.txt", filename))
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}
