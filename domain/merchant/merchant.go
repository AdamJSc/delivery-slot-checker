package merchant

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// DeliverySlot represents a single delivery slot
type DeliverySlot interface {
	GetTime() time.Time
	IsAvailable() bool
}

// DailySchedule represents a collection of DeliverySlots that pertain to the same single day
type DailySchedule struct {
	Date  time.Time
	Slots []DeliverySlot
}

// DeliveryManifest represents a series of DailySchedules
type DeliveryManifest struct {
	MerchantName   string
	Postcode       string
	From           time.Time
	Until          time.Time
	DailySchedules []DailySchedule
	Created        time.Time
}

// AsMessageText renders object as a string to be sent as a message
func (m DeliveryManifest) AsMessageText(name string) string {
	var summaries []string
	for _, schedule := range m.DailySchedules {
		summaries = append(summaries, fmt.Sprintf("%s (%d)", schedule.Date.Format("Mon 2"), len(schedule.Slots)))
	}

	return fmt.Sprintf(
		"Hi %s, forthcoming slots for %s (%s) correct at %s: %s",
		name,
		m.MerchantName,
		m.Postcode,
		time.Now().Format("3:04pm"),
		strings.Join(summaries, ", "),
	)
}

func (m DeliveryManifest) MarshalJSON() ([]byte, error) {
	simplified := struct {
		Merchant string              `json:"merchant"`
		Created  string              `json:"created"`
		Schedule map[string][]string `json:"schedule"`
	}{
		Merchant: m.MerchantName,
		Created:  m.Created.Format("3:04pm on Mon 2 Jan"),
		Schedule: make(map[string][]string),
	}

	for idx, schedule := range m.DailySchedules {
		key := fmt.Sprintf("%02d_%s", idx+1, schedule.Date.Format("Mon 2 Jan"))

		for _, slot := range schedule.Slots {
			simplified.Schedule[key] = append(simplified.Schedule[key], slot.GetTime().Format("3:04pm"))
		}
	}

	return json.Marshal(simplified)
}

// GetSlotCount returns the total number of slots across all DailySchedules
func (m DeliveryManifest) GetSlotCount() int {
	var count int

	for _, schedule := range m.DailySchedules {
		count += len(schedule.Slots)
	}

	return count
}

// SortByDate orders all DailySchedules ascending by date if asc is true, otherwise descending
func (m DeliveryManifest) SortByDate(asc bool) {
	sort.Slice(m.DailySchedules, func(i int, j int) bool {
		return asc == m.DailySchedules[i].Date.Before(m.DailySchedules[j].Date)
	})
}

// FilterByAvailability removes all slots from all daily schedules whose availability does not match the provided flag
func (m *DeliveryManifest) FilterByAvailability(isAvailable bool) {
	filteredManifest := DeliveryManifest{
		MerchantName:   m.MerchantName,
		Postcode:       m.Postcode,
		From:           m.From,
		Until:          m.Until,
		DailySchedules: []DailySchedule{},
		Created:        m.Created,
	}

	for _, schedule := range m.DailySchedules {
		filteredSchedule := DailySchedule{
			Date: schedule.Date,
		}

		for _, slot := range schedule.Slots {
			if slot.IsAvailable() == isAvailable {
				filteredSchedule.Slots = append(filteredSchedule.Slots, slot)
			}
		}

		if len(filteredSchedule.Slots) > 0 {
			filteredManifest.DailySchedules = append(filteredManifest.DailySchedules, filteredSchedule)
		}
	}

	*m = filteredManifest
}

// NewDeliveryManifest populates a new DeliveryManifest from the provided merchantName and DeliverySlots
func NewDeliveryManifest(merchantName string, postcode string, slots []DeliverySlot) (DeliveryManifest, error) {
	scheduleMap := make(map[string]DailySchedule)

	for _, slot := range slots {
		ts := slot.GetTime()
		yyyymmdd := ts.Format("20060102")

		schedule := scheduleMap[yyyymmdd]

		slotDateAtMidnight, err := time.Parse("20060102", yyyymmdd)
		if err != nil {
			return DeliveryManifest{}, err
		}
		schedule.Date = slotDateAtMidnight
		schedule.Slots = append(schedule.Slots, slot)

		scheduleMap[yyyymmdd] = schedule
	}

	manifest := DeliveryManifest{
		MerchantName: merchantName,
		Postcode:     postcode,
		Created:      time.Now(),
	}
	for _, schedule := range scheduleMap {
		manifest.DailySchedules = append(manifest.DailySchedules, schedule)
	}

	count := len(manifest.DailySchedules)
	if count > 0 {
		manifest.From = manifest.DailySchedules[0].Date
		manifest.Until = manifest.DailySchedules[count-1].Date
	}

	return manifest, nil
}

// Client represents a single merchant
type Client interface {
	GetName() string
	GetDeliverySlots(locationID string, from, to time.Time) ([]DeliverySlot, error)
}
