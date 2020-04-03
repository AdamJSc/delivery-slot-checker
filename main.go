package main

import (
	"delivery-slot-checker/domain/work"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// requiredEnv represents the env keys required by our program
var requiredEnv = []string{
	"NEXMO_KEY",
	"NEXMO_SECRET",
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	for _, key := range requiredEnv {
		if os.Getenv(key) == "" {
			log.Fatal(fmt.Errorf("missing env value: %s", key))
		}
	}

	asdaCheckDeliverySlotsJob := work.Job{
		Name:     "asda-check-delivery-slots-job",
		Task:     work.AsdaCheckDeliverySlotsTask,
		Interval: 600,
	}

	runner := work.Runner{
		Writer: os.Stdout,
		Jobs: []work.Job{
			asdaCheckDeliverySlotsJob,
		},
	}

	runner.Run()
}
