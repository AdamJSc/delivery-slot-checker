package work

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const stateDir = "./data/taskstate"

// TaskState represents the latest state of our task
type TaskState struct {
	BypassUntil time.Time `json:"bypass_until"`
	LatestRun   time.Time `json:"latest_run"`
	FirstRun    time.Time `json:"first_run"`
}

// LoadState reads task state from disk
func LoadState(name string) (TaskState, error) {
	contents, err := ioutil.ReadFile(getFullPathToStateFile(name))
	if err != nil {
		return TaskState{}, err
	}

	var taskState TaskState
	err = json.Unmarshal(contents, &taskState)
	if err != nil {
		return TaskState{}, err
	}

	return taskState, err
}

// SaveState stores task state on disk
func SaveState(name string, state TaskState) error {
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
func LoadStateCreateIfMissing(name string) (TaskState, error) {
	state, err := LoadState(name)

	if err != nil {
		// save initial state
		state = TaskState{
			FirstRun:  time.Now(),
			LatestRun: time.Now(),
		}
		if err := SaveState(name, state); err != nil {
			return TaskState{}, err
		}
	}

	return state, nil
}

// getFullPathToStateFile gets full path to state file from name
func getFullPathToStateFile(name string) string {
	return fmt.Sprintf("%s/%s.txt", stateDir, name)
}
