package work

import (
	"delivery-slot-checker/domain/apperrors"
	"fmt"
	"io"
	"os"
	"strings"
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

// Task represents the function executed by a Job
type Task func(state *JobState, w WriterWithIdentifier) error

// Job represents a single unit of work
type Job struct {
	Name     string
	Task     Task
	Interval time.Duration
}

// GetIdentifier returns a formatted ID
func (j Job) GetIdentifier() string {
	lower := strings.Trim(strings.ToLower(j.Name), " ")
	return strings.Replace(lower, " ", "-", -1)
}

// Runner represents a collection of Jobs to execute continuously
type Runner struct {
	Writer io.Writer
	Jobs   []Job
}

// runJob enables the concurrent execution of a Job
func runJob(job Job, w WriterWithIdentifier, ch chan Job) {
	stateName := fmt.Sprintf("%s_%s", job.GetIdentifier(), time.Now().Format("20060102"))

	state, err := LoadStateCreateIfMissing(stateName)
	if err != nil {
		fmt.Fprintln(w, fmt.Sprintf("error loading state: %s", stateName), err)
		os.Exit(1)
	}

	if err = job.Task(&state, w); err != nil {
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

	time.Sleep(job.Interval * time.Second)

	ch <- job
}

// Run executes all Jobs that belong to the Runner
func (r Runner) Run() {
	ch := make(chan Job)
	taskWriters := make(map[string]WriterWithIdentifier)

	for _, job := range r.Jobs {
		if job.Interval < minInterval {
			fmt.Fprintf(r.Writer, "minimum interval %d: interval of %d too short for job '%s'\n", minInterval, job.Interval, job.Name)
			os.Exit(1)
		}

		jobID := job.GetIdentifier()
		taskWriters[jobID] = WriterWithIdentifier{Identifier: jobID, Writer: r.Writer}
		go runJob(job, taskWriters[jobID], ch)
	}

	for job := range ch {
		go runJob(job, taskWriters[job.GetIdentifier()], ch)
	}
}
