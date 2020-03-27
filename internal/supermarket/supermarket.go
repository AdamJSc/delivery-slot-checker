package supermarket

type Client interface {
	GetDeliverySlots() ([]DeliverySlot, error)
}

func NewClient() Client {
	return AsdaClient{}
}

type DeliverySlot interface {
	IsAvailable() bool
}

func FilterAvailableDeliverySlots(slots []DeliverySlot) (filtered []DeliverySlot) {
	for _, slot := range slots {
		if slot.IsAvailable() {
			filtered = append(filtered, slot)
		}
	}

	return filtered
}
