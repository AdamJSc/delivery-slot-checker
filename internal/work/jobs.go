package work

import (
	"delivery-slot-checker/internal/apperrors"
	"log"
	"os"
	"time"
)

const minInterval = 600

type Task func(l *log.Logger) error

type Job struct {
	Name     string
	Task     Task
	Interval time.Duration
}

type Runner struct {
	Logger *log.Logger
	Jobs   []Job
}

func runJob(job Job, l *log.Logger) {
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
	runJob(job, l)
}

func (r Runner) Run() {
	for _, job := range r.Jobs {
		if job.Interval < minInterval {
			r.Logger.Printf("minimum interval %d: interval of %d too short for job '%s'\n", minInterval, job.Interval, job.Name)
			os.Exit(1)
		}

		runJob(job, r.Logger)
	}
}
