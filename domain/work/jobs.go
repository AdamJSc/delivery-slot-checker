package work

import (
	"delivery-slot-checker/domain/apperrors"
	"fmt"
	"io"
	"os"
	"time"
)

const minInterval = 600

// WriterWithIdentifier represents a writer with a log-style prefix that appears after a timestamp
type WriterWithIdentifier struct {
	io.Writer
	Identifier string
}

func (w WriterWithIdentifier) Write(p []byte) (int, error) {
	input := string(p)
	ts := time.Now().Format("2006-01-02 15:04:05")

	value := fmt.Sprintf("%s [%s] %s", ts, w.Identifier, input)

	return w.Writer.Write([]byte(value))
}

// TaskPayload represents the data structure that will be passed to, and acted on by, a Task
type TaskPayload struct {
	Identifier string        `yaml:"identifier"`
	Interval   time.Duration `yaml:"interval"`
	Postcode   string        `yaml:"postcode"`
	Recipients []struct {
		Name   string `yaml:"name"`
		Mobile string `yaml:"mobile"`
	} `yaml:"recipients"`
}

// Task represents the function executed by a Job
type Task func(payload TaskPayload, state *TaskState, w WriterWithIdentifier) error

// Job represents a single unit of work
type Job struct {
	Identifier string
	Task       Task
	Payloads   []TaskPayload
}

// Runner represents a collection of Jobs to execute continuously
type Runner struct {
	Writer io.Writer
	Jobs   []Job
}

// runTask enables the concurrent execution of a Task
func runTask(task Task, payload TaskPayload, w WriterWithIdentifier, ch chan Task) {
	stateName := fmt.Sprintf("%s_%s", payload.Identifier, time.Now().Format("20060102"))

	state, err := LoadStateCreateIfMissing(stateName)
	if err != nil {
		fmt.Fprintln(w, fmt.Sprintf("error loading state: %s", stateName), err)
		os.Exit(1)
	}

	if err = task(payload, &state, w); err != nil {
		fmt.Fprintln(w, err)

		switch err.(type) {
		case apperrors.FatalError:
			os.Exit(1)
		}
	}

	if err = SaveState(stateName, state); err != nil {
		fmt.Fprintln(w, fmt.Sprintf("error saving state: %s", stateName), err)
		os.Exit(1)
	}

	time.Sleep(payload.Interval * time.Second)

	ch <- task
}

// Run executes all Jobs that belong to the Runner
func (r Runner) Run() {
	ch := make(chan Task)
	taskWriters := make(map[string]WriterWithIdentifier)

	for _, job := range r.Jobs {
		for _, payload := range job.Payloads {
			payload.Identifier = fmt.Sprintf("%s_%s", job.Identifier, payload.Identifier)
			taskWriters[payload.Identifier] = WriterWithIdentifier{Identifier: payload.Identifier, Writer: r.Writer}

			if payload.Interval < minInterval {
				payload.Interval = minInterval
			}

			go runTask(job.Task, payload, taskWriters[payload.Identifier], ch)

			for task := range ch {
				go runTask(task, payload, taskWriters[payload.Identifier], ch)
			}
		}
	}
}
