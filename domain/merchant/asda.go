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

var asdaURL = "https://groceries.asda.com/api/v3"

// AsdaClient defines a client for interacting with the Asda supermarket.
type AsdaClient struct {
	Client
}

// GetName returns the supermarket name.
func (c AsdaClient) GetName() string {
	return "Asda"
}

// GetDeliverySlots returns delivery slots from Asda within the given date range, for a given postcode.
func (c AsdaClient) GetDeliverySlots(postcode string, from, to time.Time) ([]DeliverySlot, error) {
	var response asdaAPIResponse

	httpRequestBody, err := json.Marshal(newAsdaAPIRequest(postcode, from, to))
	if err != nil {
		return []DeliverySlot{}, apperrors.FatalError{Err: err}
	}

	httpResponse, err := http.Post(
		fmt.Sprintf("%s/slot/view", asdaURL),
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
		return []DeliverySlot{}, apperrors.OfflineError{
			Merchant: c.GetName(),
		}
	}

	var slots []DeliverySlot
	for _, slotDay := range response.Data.SlotDays {
		for _, slot := range slotDay.Slots {
			slots = append(slots, slot.SlotInfo)
		}
	}

	return slots, nil
}

// AsdaDeliverySlot represents a delivery slot for Asda.
type AsdaDeliverySlot struct {
	DeliverySlot
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
}

// IsAvailable checks whether a delivery slot can still be booked.
func (ds AsdaDeliverySlot) IsAvailable() bool {
	return ds.Status == asdaStatusAvailable
}

// GetTime returns the start time of the delivery slot
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
	Data Data `json:"data"`
}

// Data defines the data within an asdaAPIrequest.
type Data struct {
	ServiceInfo    ServiceInfo    `json:"service_info"`
	ServiceAddress ServiceAddress `json:"service_address"`
	CustomerInfo   CustomerInfo   `json:"customer_info"`
	OrderInfo      OrderInfo      `json:"order_info"`
	StartDate      time.Time      `json:"start_date"`
	EndDate        time.Time      `json:"end_date"`
}

// ServiceInfo defines the service provided by asda.
type ServiceInfo struct {
	FulfillmentType string `json:"fulfillment_type"`
}

// ServiceAddress holds the postcode requested.
type ServiceAddress struct {
	Postcode string `json:"postcode"`
}

// CustomerInfo holds account information for the customer.
type CustomerInfo struct {
	AccountID string `json:"account_id"`
}

// OrderInfo holds order information relevant to the request.
type OrderInfo struct {
	OrderID string `json:"order_id"`
}

func newAsdaAPIRequest(postcode string, start, end time.Time) asdaAPIRequest {
	return asdaAPIRequest{
		Data: Data{
			ServiceInfo: ServiceInfo{
				FulfillmentType: "DELIVERY",
			},
			ServiceAddress: ServiceAddress{
				Postcode: postcode,
			},
			CustomerInfo: CustomerInfo{
				AccountID: "0",
			},
			OrderInfo: OrderInfo{
				OrderID: "0",
			},
			StartDate: start,
			EndDate:   end,
		},
	}
}
