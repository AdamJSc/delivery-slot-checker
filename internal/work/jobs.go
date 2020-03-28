package work

import (
	"delivery-slot-checker/internal/apperrors"
	"log"
	"os"
	"time"
)

const minInterval = 600

// Task represents the function executed by a Job
type Task func(l *log.Logger) error

// Job represents a single unit of work
type Job struct {
	Name     string
	Task     Task
	Interval time.Duration
}

// Runner represents a collection of Jobs to execute continuously
type Runner struct {
	Logger *log.Logger
	Jobs   []Job
}

// runJob enables the concurrent execution of a Job
func runJob(job Job, l *log.Logger, ch chan Job) {
	prefixedLogger := log.New(l.Writer(), job.Name + ": ", l.Flags())
	defer func() {
		prefixedLogger = nil
	}()

	err := job.Task(prefixedLogger)

	if err != nil {
		prefixedLogger.Println(err)

		switch err.(type) {
		case apperrors.FatalError:
			os.Exit(1)
		}
	}

	time.Sleep(job.Interval * time.Second)

	ch <- job
}

// Run executes all Jobs that belong to the Runner
func (r Runner) Run() {
	ch := make(chan Job)

	for _, job := range r.Jobs {
		if job.Interval < minInterval {
			r.Logger.Printf("minimum interval %d: interval of %d too short for job '%s'\n", minInterval, job.Interval, job.Name)
			os.Exit(1)
		}

		go runJob(job, r.Logger, ch)
	}

	for job := range ch {
		go runJob(job, r.Logger, ch)
	}
}
