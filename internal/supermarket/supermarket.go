package supermarket

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type DeliverySlot interface {
	GetTime() time.Time
	IsAvailable() bool
}

func FilterDeliverySlotsByAvailable(slots []DeliverySlot) (filtered []DeliverySlot) {
	for _, slot := range slots {
		if slot.IsAvailable() {
			filtered = append(filtered, slot)
		}
	}

	return filtered
}

type DailySchedule struct {
	Date  time.Time
	Slots []DeliverySlot
}

type AvailabilityManifest struct {
	Chain          string
	Created        time.Time
	DailySchedules []DailySchedule
}

func (m AvailabilityManifest) MarshalJSON() ([]byte, error) {
	simplified := struct {
		Chain    string              `json:"chain"`
		Created  string              `json:"created"`
		Schedule map[string][]string `json:"schedule"`
	}{
		Chain:    m.Chain,
		Created:  m.Created.Format("3:04pm on Mon 2 Jan"),
		Schedule: make(map[string][]string),
	}

	for idx, schedule := range m.DailySchedules {
		key := fmt.Sprintf("%02d_%s", idx + 1, schedule.Date.Format("Mon 2 Jan"))

		for _, slot := range schedule.Slots {
			simplified.Schedule[key] = append(simplified.Schedule[key], slot.GetTime().Format("3:04pm"))
		}
	}

	return json.Marshal(simplified)
}

func (m *AvailabilityManifest) GetSlotCount() int {
	var count int

	for _, schedule := range m.DailySchedules {
		count += len(schedule.Slots)
	}

	return count
}

func (m *AvailabilityManifest) SortByDate(asc bool) {
	sort.Slice(m.DailySchedules, func(i int, j int) bool {
		return asc == m.DailySchedules[i].Date.Before(m.DailySchedules[j].Date)
	})
}

func (m *AvailabilityManifest) GetFirstDate() time.Time {
	if len(m.DailySchedules) == 0 {
		return time.Time{}
	}

	return m.DailySchedules[0].Date
}

func (m *AvailabilityManifest) GetLastDate() time.Time {
	schedulesCount := len(m.DailySchedules)
	if schedulesCount == 0 {
		return time.Time{}
	}

	return m.DailySchedules[schedulesCount-1].Date
}

func GetAvailabilityManifestFromSlots(chain string, slots []DeliverySlot) (AvailabilityManifest, error) {
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
		Chain:   chain,
		Created: time.Now(),
	}
	for _, schedule := range scheduleMap {
		manifest.DailySchedules = append(manifest.DailySchedules, schedule)
	}

	return manifest, nil
}

type Client interface {
	GetChain() string
	GetDeliverySlots() ([]DeliverySlot, error)
}

func NewClient() Client {
	return AsdaClient{}
}
