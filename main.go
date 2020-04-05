package main

import (
	"delivery-slot-checker/domain/work"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/joho/godotenv"
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

	// retrieve and parse task payloads
	var taskPayloads []work.TaskPayload
	taskPayloadsFileContents, err := ioutil.ReadFile("./data/task/payloads.yml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(taskPayloadsFileContents, &taskPayloads)
	if err != nil {
		log.Fatal(err)
	}

	asdaCheckDeliverySlotsJob := work.Job{
		Name:     "asda-check-delivery-slots-job",
		Task:     work.AsdaCheckDeliverySlotsTask,
		Payloads: taskPayloads,
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
