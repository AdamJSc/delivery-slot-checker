package main

import (
	"delivery-slot-checker/internal/work"
	"log"
	"os"
)

func main() {
	asdaCheckDeliverySlotsJob := work.Job{
		Name:     "asda-check-delivery-slots-job",
		Task:     work.AsdaCheckDeliverySlotsTask,
		Interval: 600,
	}

	runner := work.Runner{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		Jobs: []work.Job{
			asdaCheckDeliverySlotsJob,
		},
	}

	runner.Run()
}
