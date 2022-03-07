package output

import (
	"testing"
	"time"
)

func TestGenFilename(t *testing.T) {
	testcases := []struct {
		label string
		in    time.Time
		exp   string
	}{
		{
			label: "return timestamped filename",
			in:    time.Date(1234, time.December, 13, 14, 15, 16, 17, time.UTC),
			exp:   "./benchttp.report.12341213141516.json",
		},
		{
			label: "return timestamped filename with added zeros",
			in:    time.Date(1, time.January, 1, 1, 1, 1, 1, time.UTC),
			exp:   "./benchttp.report.00010101010101.json",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			got := genFilename(tc.in)
			if got != tc.exp {
				t.Errorf("\nexp %s\ngot %s", tc.exp, got)
			}
		})
	}
}
