package supermarket

import (
	"time"
)

type DeliverySlot interface {
	GetTime() time.Time
	IsAvailable() bool
}

type DeliverySchedule struct {
	Date  time.Time
	Slots []DeliverySlot
}

type DeliveryManifest []DeliverySchedule

type Client interface {
	GetDeliverySlots() ([]DeliverySlot, error)
}

func FilterAvailableDeliverySlots(slots []DeliverySlot) (filtered []DeliverySlot) {
	for _, slot := range slots {
		if slot.IsAvailable() {
			filtered = append(filtered, slot)
		}
	}

	return filtered
}

func GetDeliveryManifestFromSlots(slots []DeliverySlot) (DeliveryManifest, error) {
	scheduleMap := make(map[string]DeliverySchedule)

	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return DeliveryManifest{}, err
	}

	for _, slot := range slots {
		ts := slot.GetTime()
		yyyymmdd := ts.Format("20060102")

		schedule := scheduleMap[yyyymmdd]

		schedule.Date = time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc)
		schedule.Slots = append(schedule.Slots, slot)

		scheduleMap[yyyymmdd] = schedule
	}

	var manifest DeliveryManifest
	for _, schedule := range scheduleMap {
		manifest = append(manifest, schedule)
	}

	return manifest, nil
}

func NewClient() Client {
	return AsdaClient{}
}
