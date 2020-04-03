package work

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const stateDir = "./data/jobstate"

// JobState represents the latest state of our job
type JobState struct {
	Bypass    bool      `json:"bypass"`
	Status    string    `json:"status"`
	LatestRun time.Time `json:"latest_run"`
	FirstRun  time.Time `json:"first_run"`
}

// LoadState reads job state from disk
func LoadState(name string) (JobState, error) {
	contents, err := ioutil.ReadFile(getFullPathToStateFile(name))
	if err != nil {
		return JobState{}, err
	}

	var jobState JobState
	err = json.Unmarshal(contents, &jobState)
	if err != nil {
		return JobState{}, err
	}

	return jobState, err
}

// SaveState stores job state on disk
func SaveState(name string, state JobState) error {
	err := os.MkdirAll(stateDir, 0755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(getFullPathToStateFile(name), data, 0755)
}

// LoadStateCreateIfMissing will attempt to load state from disk, or create if not existing
func LoadStateCreateIfMissing(name string) (JobState, error) {
	state, err := LoadState(name)

	if err != nil {
		// save initial state
		state = JobState{
			FirstRun:  time.Now(),
			LatestRun: time.Now(),
		}
		if err := SaveState(name, state); err != nil {
			return JobState{}, err
		}
	}

	return state, nil
}

// getFullPathToStateFile gets full path to state file from name
func getFullPathToStateFile(name string) string {
	return fmt.Sprintf("%s/%s.txt", stateDir, name)
}
