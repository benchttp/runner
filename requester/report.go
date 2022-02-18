package requester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Report represents the collected results of a benchmark test.
type Report struct {
	Records  []Record      `json:"records"`
	Length   int           `json:"length"`
	Success  int           `json:"success"`
	Fail     int           `json:"fail"`
	Duration time.Duration `json:"duration"`
}

// String returns an indented JSON representation of the report.
func (rep Report) String() string {
	b, _ := json.MarshalIndent(rep, "", "  ")
	return string(b)
}

// Stats returns basic stats about the report's records:
// min duration, max duration, and mean duration.
// It does not replace the remote computing and should only be used
// when a local reporting is needed.
func (rep Report) Stats() (min, max, mean time.Duration) {
	var sum time.Duration
	for _, rec := range rep.Records {
		d := rec.Time
		if d < min || min == 0 {
			min = d
		}
		if d > max {
			max = d
		}
		sum += rec.Time
	}
	return min, max, sum / time.Duration(rep.Length)
}

// makeReport generates and returns a Report from a previous Run.
func makeReport(records []Record, numErr int, d time.Duration) Report {
	return Report{
		Records:  records,
		Length:   len(records),
		Success:  len(records) - numErr,
		Fail:     numErr,
		Duration: d,
	}
}

// SendReport sends the report to url. Returns any non-nil error that occurred.;
//
// TODO: move from requester
func SendReport(url string, report Report) error {
	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(report); err != nil {
		return fmt.Errorf("%w: %s", ErrReporting, err)
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrReporting, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrReporting, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: %s", ErrReporting, resp.Status)
	}

	return nil
}
