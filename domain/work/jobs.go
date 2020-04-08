package work

import (
	"delivery-slot-checker/domain/apperrors"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

// minInterval determines the minimum permitted number of seconds' delay before a task is next run
const minInterval = 600

// offset determines the amount of interval flux +/-
// a value of 7 with an interval setting of 600 would result in a real-life delay of
// somewhere between 593 and 607 seconds
const offset = 7

// bypassDuration represents a default duration in minutes for which bypassed tasks should be ignored by the runner
const bypassDuration = 120

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
	stateName := fmt.Sprintf("%s_%s", time.Now().Format("20060102"), payload.Identifier)

	state, err := LoadStateCreateIfMissing(stateName)
	if err != nil {
		fmt.Fprintln(w, fmt.Sprintf("error loading state: %s", stateName), err)
		os.Exit(1)
	}

	var checkForBypassAndRun = func() error {
		if time.Now().Before(state.BypassUntil) {
			return errors.New("bypassing task...")
		}
		return task(payload, &state, w)
	}

	if err = checkForBypassAndRun(); err != nil {
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

	time.Sleep(getRandomisedInterval(payload.Interval) * time.Second)

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

			go func(payload TaskPayload) {
				// initial randomised interval
				time.Sleep(getRandomisedInterval(payload.Interval) * time.Second)
				go runTask(job.Task, payload, taskWriters[payload.Identifier], ch)

				for task := range ch {
					go runTask(task, payload, taskWriters[payload.Identifier], ch)
				}
			}(payload)
		}
	}

	// block indefinitely to allow runTask() goroutines to run within per-payload goroutines
	// if any of the task runs return an apperrors.FatalError, current program will exit
	func() { select {} }()
}

// getRandomisedInterval returns a random duration based on the provided interval
func getRandomisedInterval(interval time.Duration) time.Duration {
	var base, lowerLimit, upperLimit time.Duration

	base = interval
	if interval < minInterval {
		base = minInterval
	}

	lowerLimit = base - offset
	if base-offset <= 0 {
		lowerLimit = base
	}

	upperLimit = base + offset

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomInterval := r.Intn(int(upperLimit)-int(lowerLimit)-1) + int(lowerLimit)

	return time.Duration(randomInterval)
}

func getDefaultBypassDuration() time.Duration {
	return bypassDuration * time.Minute
}
