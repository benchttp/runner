package requester

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/runner/ansi"
)

var (
	tlBlock      = "◼︎"
	tlBlockGrey  = ansi.Grey(tlBlock)
	tlBlockGreen = ansi.Green(tlBlock)
	tlLen        = 10
)

func (r *Requester) state() string {
	r.mu.Lock()

	var (
		status    = r.status()
		reqcur    = len(r.records)
		reqmax    = strconv.Itoa(r.config.RunnerOptions.Requests)
		pctdone   = r.percentDone()
		elapsed   = time.Since(r.start)
		countdown = r.config.RunnerOptions.GlobalTimeout - elapsed
	)

	r.mu.Unlock()

	if reqmax == "-1" {
		reqmax = "∞"
	}
	if countdown < 0 {
		countdown = 0
	}

	tl := strings.Repeat(tlBlockGrey, tlLen)
	for i := 0; i < tlLen; i++ {
		if pctdone >= (tlLen * i) {
			tl = strings.Replace(tl, tlBlockGrey, tlBlockGreen, 1)
		}
	}

	return fmt.Sprintf(
		"%s %s %d%% | %d/%s requests | %.0fs timeout       \r",
		status, tl, pctdone, reqcur, reqmax, countdown.Seconds(),
	)
}

func (r *Requester) status() string {
	if !r.done {
		return ansi.Yellow("RUNNING")
	}
	switch r.runErr {
	case nil:
		return ansi.Green("DONE")
	case context.Canceled:
		return ansi.Cyan("CANCELED")
	case context.DeadlineExceeded:
		return ansi.Cyan("TIMEOUT")
	}
	return "" // should not occur
}

func (r *Requester) percentDone() int {
	var cur, max int
	if r.config.RunnerOptions.Requests == -1 {
		cur, max = int(time.Since(r.start)), int(r.config.RunnerOptions.GlobalTimeout)
	} else {
		cur, max = len(r.records), r.config.RunnerOptions.Requests
	}
	return capInt((100*cur)/max, 100)
}

func capInt(n, max int) int {
	if n > max {
		return max
	}
	return n
}
