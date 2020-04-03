package merchant

import (
	"bytes"
	"delivery-slot-checker/domain/apperrors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	asdaStatusAvailable   = "AVAILABLE"
	asdaStatusUnavailable = "UNAVAILABLE"
)

type AsdaClient struct {
	Client
	URL string
}

func (c AsdaClient) GetName() string {
	return "Asda"
}

func (c AsdaClient) GetDeliverySlots(postcode string, from, to time.Time) ([]DeliverySlot, error) {
	var response asdaAPIResponse

	httpRequestBody, err := json.Marshal(newAsdaAPIRequest(postcode, from, to))
	if err != nil {
		return []DeliverySlot{}, apperrors.FatalError{Err: err}
	}

	httpResponse, err := http.Post(
		fmt.Sprintf("%s/slot/view", c.URL),
		"application/json",
		bytes.NewReader(httpRequestBody),
	)
	if err != nil {
		return []DeliverySlot{}, err
	}

	httpResponseBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return []DeliverySlot{}, apperrors.FatalError{Err: err}
	}

	if err = json.Unmarshal(httpResponseBody, &response); err != nil {
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

type asdaAPIRequest struct {
	Data struct {
		ServiceInfo struct {
			FulfillmentType string `json:"fulfillment_type"`
		} `json:"service_info"`
		StartDate      time.Time `json:"start_date"`
		EndDate        time.Time `json:"end_date"`
		ServiceAddress struct {
			Postcode string `json:"postcode"`
		} `json:"service_address"`
		CustomerInfo struct {
			AccountID string `json:"account_id"`
		} `json:"customer_info"`
		OrderInfo struct {
			OrderID string `json:"order_id"`
		} `json:"order_info"`
	} `json:"data"`
}

func newAsdaAPIRequest(postcode string, start, end time.Time) asdaAPIRequest {
	var request asdaAPIRequest

	request.Data.ServiceInfo.FulfillmentType = "DELIVERY"
	request.Data.StartDate = start
	request.Data.EndDate = end
	request.Data.ServiceAddress.Postcode = postcode
	request.Data.CustomerInfo.AccountID = "0"
	request.Data.OrderInfo.OrderID = "0"

	return request
}
