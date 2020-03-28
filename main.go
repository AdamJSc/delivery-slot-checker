package main

import (
	"delivery-slot-checker/internal/work"
	"os"
)

func main() {
	asdaCheckDeliverySlotsJob := work.Job{
		Name:     "asda-check-delivery-slots-job",
		Task:     work.AsdaCheckDeliverySlotsTask,
		Interval: 3,
	}

	runner := work.Runner{
		Writer: os.Stdout,
		Jobs: []work.Job{
			asdaCheckDeliverySlotsJob,
		},
	}

	runner.Run()
}
