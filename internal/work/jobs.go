package work

import (
	"delivery-slot-checker/internal/apperrors"
	"fmt"
	"io"
	"os"
	"time"
)

const minInterval = 1

// TaskWriter represents a Task-specific writer
type TaskWriter struct {
	io.Writer
	TaskName string
}

func (w TaskWriter) Write(p []byte) (int, error) {
	input := string(p)
	ts := time.Now().Format("2006-01-02 15:04:05")

	value := fmt.Sprintf("%s [%s] %s", ts, w.TaskName, input)

	return w.Writer.Write([]byte(value))
}

// Task represents the function executed by a Job
type Task func(w TaskWriter) error

// Job represents a single unit of work
type Job struct {
	Name     string
	Task     Task
	Interval time.Duration
}

// Runner represents a collection of Jobs to execute continuously
type Runner struct {
	Writer io.Writer
	Jobs   []Job
}

// runJob enables the concurrent execution of a Job
func runJob(job Job, w io.Writer, ch chan Job) {
	taskWriter := &TaskWriter{TaskName: job.Name, Writer: w}
	defer func() {
		taskWriter = nil
	}()

	err := job.Task(*taskWriter)

	if err != nil {
		fmt.Fprintln(taskWriter, err)

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
			fmt.Fprintf(r.Writer, "minimum interval %d: interval of %d too short for job '%s'\n", minInterval, job.Interval, job.Name)
			os.Exit(1)
		}

		go runJob(job, r.Writer, ch)
	}

	for job := range ch {
		go runJob(job, r.Writer, ch)
	}
}
