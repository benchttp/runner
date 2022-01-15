package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Record struct {
	cost  time.Duration
	code  string
	error string
	bytes int
}

func newRecord(resp *http.Response, t time.Duration) Record {
	var code string

	switch resp.StatusCode / 100 {
	case 1:
		code = "1xx"
	case 2:
		code = "2xx"
	case 3:
		code = "3xx"
	case 4:
		code = "4xx"
	case 5:
		code = "5xx"
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	r := Record{
		code:  code,
		cost:  t,
		bytes: len(body),
	}

	if err != nil {
		r.error = fmt.Sprint(err)
	}

	return r
}

func (r Record) String() string {
	return fmt.Sprintf("status %s, took %s", r.code, r.cost)
}

func Do(url string, timeout time.Duration) Record {
	client := http.Client{
		// Timeout includes connection time, any redirects, and reading the response body.
		// We may want exclude reading the response body in our benchmark tool.
		Timeout: timeout,
	}

	t0 := time.Now()

	resp, err := client.Get(url)
	if err != nil {
		return Record{error: fmt.Sprint(err)}
	}

	return newRecord(resp, time.Since(t0))
}
