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

// FilterDeliverySlotsByAvailability returns only those DeliverySlots within a slice that match the provided availability
func FilterDeliverySlotsByAvailability(slots []DeliverySlot, isAvailable bool) (filtered []DeliverySlot) {
	for _, slot := range slots {
		if slot.IsAvailable() == isAvailable {
			filtered = append(filtered, slot)
		}
	}

	return filtered
}

// DailySchedule represents a collection of DeliverySlots that pertain to the same single day
type DailySchedule struct {
	Date  time.Time
	Slots []DeliverySlot
}

// AvailabilityManifest represents a series of DailySchedules
type AvailabilityManifest struct {
	MerchantName   string
	Created        time.Time
	DailySchedules []DailySchedule
}

// AsMessageText renders object as a string to be sent as a message
func (m AvailabilityManifest) AsMessageText(name string) string {
	from := m.GetFirstDate()
	to := m.GetLastDate()
	dateRange := int(to.Sub(from).Hours() / 24)

	var summaries []string
	for _, schedule := range m.DailySchedules {
		summaries = append(summaries, fmt.Sprintf("%s (%d)", schedule.Date.Format("Mon 2"), len(schedule.Slots)))
	}

	return fmt.Sprintf(
		"Hi %s, %s slots next %d days as at %s: %s",
		name,
		m.MerchantName,
		dateRange,
		time.Now().Format("3:04pm"),
		strings.Join(summaries, ", "),
	)
}

func (m AvailabilityManifest) MarshalJSON() ([]byte, error) {
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
func (m *AvailabilityManifest) GetSlotCount() int {
	var count int

	for _, schedule := range m.DailySchedules {
		count += len(schedule.Slots)
	}

	return count
}

// SortByDate orders all DailySchedules ascending by date if asc is true, otherwise descending
func (m *AvailabilityManifest) SortByDate(asc bool) {
	sort.Slice(m.DailySchedules, func(i int, j int) bool {
		return asc == m.DailySchedules[i].Date.Before(m.DailySchedules[j].Date)
	})
}

// GetFirstDate returns the date of the first DailySchedule in an AvailabilityManifest
func (m *AvailabilityManifest) GetFirstDate() time.Time {
	if len(m.DailySchedules) == 0 {
		return time.Time{}
	}

	return m.DailySchedules[0].Date
}

// GetLastDate returns the date of the last DailySchedule in an AvailabilityManifest
func (m *AvailabilityManifest) GetLastDate() time.Time {
	schedulesCount := len(m.DailySchedules)
	if schedulesCount == 0 {
		return time.Time{}
	}

	return m.DailySchedules[schedulesCount-1].Date
}

func GetAvailabilityManifestFromSlots(merchantName string, slots []DeliverySlot) (AvailabilityManifest, error) {
	scheduleMap := make(map[string]DailySchedule)

	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return AvailabilityManifest{}, err
	}

	for _, slot := range slots {
		ts := slot.GetTime()
		yyyymmdd := ts.Format("20060102")

		schedule := scheduleMap[yyyymmdd]

		schedule.Date = time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc)
		schedule.Slots = append(schedule.Slots, slot)

		scheduleMap[yyyymmdd] = schedule
	}

	manifest := AvailabilityManifest{
		MerchantName: merchantName,
		Created:      time.Now(),
	}
	for _, schedule := range scheduleMap {
		manifest.DailySchedules = append(manifest.DailySchedules, schedule)
	}

	return manifest, nil
}

// Client represents a single merchant
type Client interface {
	GetName() string
	GetDeliverySlots() ([]DeliverySlot, error)
}

// NewClient instantiates the default Client
func NewClient() Client {
	return AsdaClient{}
}
