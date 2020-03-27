package supermarket

import (
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
	DailySchedules []DailySchedule
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

	manifest := AvailabilityManifest{Chain: chain}
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
