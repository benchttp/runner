package output

import (
	"errors"
	"testing"
)

func TestReport_applyTemplate(t *testing.T) {
	testcases := []struct {
		label   string
		pattern string
		expStr  string
		expErr  error
	}{
		{
			label:   "return errTemplateEmpty if pattern is empty",
			pattern: "",
			expStr:  "",
			expErr:  errTemplateEmpty,
		},
		{
			label:   "return errTemplateSyntaxt if pattern has syntax error",
			pattern: "{{ else }}",
			expStr:  "",
			expErr:  errTemplateSyntax,
		},
		{
			label:   "return errTemplateSyntaxt if pattern doesn't match report values",
			pattern: "{{ .Foo }}", // Report.Foo doesn't exist
			expStr:  "",
			expErr:  errTemplateSyntax,
		},
		{
			label:   "happy path with custom template functions",
			pattern: "{{ stats.Min }},{{ stats.Max }},{{ stats.Mean }}",
			expStr:  "0s,0s,0s",
			expErr:  nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			r := &Report{}
			gotStr, gotErr := r.applyTemplate(tc.pattern)
			if !errors.Is(gotErr, tc.expErr) {
				t.Errorf("unexpected error: %v", gotErr)
			}
			if gotStr != tc.expStr {
				t.Errorf("unexpected string: %q", gotStr)
			}
		})
	}
}
