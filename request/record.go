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
	return fmt.Sprintf("took %s", r.cost)
}
