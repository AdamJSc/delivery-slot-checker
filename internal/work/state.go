package work

import (
	"delivery-slot-checker/internal/apperrors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const stateDir = "./data/jobstate"

type JobState struct {
	Bypass    bool      `json:"bypass"`
	Status    string    `json:"status"`
	LatestRun time.Time `json:"latest_run"`
	FirstRun  time.Time `json:"first_run"`
}

func LoadState(name string) (JobState, error) {
	contents, err := ioutil.ReadFile(getStateFullPath(name))
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

func SaveState(name string, state JobState) error {
	err := os.MkdirAll(stateDir, 0755)
	if err != nil {
		return apperrors.FatalError{Err: err}
	}

	data, err := json.Marshal(state)
	if err != nil {
		return apperrors.FatalError{Err: err}
	}

	return ioutil.WriteFile(getStateFullPath(name), data, 0755)
}

func LoadStateAndCreateIfMissing(name string) (JobState, error) {
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

func getStateFullPath(filename string) string {
	return fmt.Sprintf("%s/%s.txt", stateDir, filename)
}
