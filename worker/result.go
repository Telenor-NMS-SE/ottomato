package worker

import (
	"encoding/json"
	"errors"
	"time"
)

type Result struct {
	JobID         string         `json:"jobId"`
	WorkerID      string         `json:"workerId"`
	Tags          []string       `json:"tags"`
	Hostname      string         `json:"hostname"`
	Command       string         `json:"command"`
	Args          []string       `json:"args"`
	Kwargs        map[string]any `json:"kwargs"`
	Success       bool           `json:"success"`
	Error         error          `json:"error"`
	Return        any            `json:"return"`
	Timestamp     time.Time      `json:"timestamp"`
	ExecutionTime int64          `json:"executionTime"`
}

// Use a custom marshal to support *string errors
func (r Result) MarshalJSON() ([]byte, error) {
	type res Result

	return json.Marshal(&struct {
		res
		Error *string `json:"error"`
	}{
		res: res(r),
		Error: func() *string {
			if r.Error == nil {
				return nil
			}

			err := r.Error.Error()
			return &err
		}(),
	})
}

func (r *Result) UnmarshalJSON(input []byte) error {
	type res Result

	tmp := &struct {
		res
		Error *string `json:"error"`
	}{
		res: res(*r),
	}

	if err := json.Unmarshal(input, &tmp); err != nil {
		return err
	}

	*r = Result(tmp.res)
	if tmp.Error != nil {
		r.Error = errors.New(*tmp.Error)
	}

	return nil
}
