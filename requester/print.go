package requester

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/runner/ansi"
)

var (
	status = map[bool]string{
		false: ansi.Yellow("Running"),
		true:  ansi.Green("Done"),
	}
	tlBlock     = "◼︎"
	tlBlockGrey = ansi.Grey(tlBlock)
	tlLen       = 10
)

func (r *Requester) state() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var (
		reqcur    = len(r.records)
		reqmax    = strconv.Itoa(r.config.RunnerOptions.Requests)
		pctdone   = r.percentDone()
		elapsed   = time.Since(r.start)
		countdown = r.config.RunnerOptions.GlobalTimeout - elapsed
	)

	if reqmax == "-1" {
		reqmax = "∞"
	}
	if countdown < 0 {
		countdown = 0
	}

	tl := strings.Repeat(tlBlockGrey, tlLen)
	for i := 0; i < tlLen; i++ {
		if pctdone >= (tlLen * i) {
			tl = strings.Replace(tl, tlBlockGrey, tlBlock, 1)
		}
	}

	return fmt.Sprintf(
		"%s %d%% %s | %d/%s requests | %.0fs timeout       \r",
		status[r.done], pctdone, tl, reqcur, reqmax, countdown.Seconds(),
	)
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
