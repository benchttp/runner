package output

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/benchttp/runner/requester"
)

// applyTemplate applies Report to a template using given pattern and returns
// the result as a string. If pattern == "", it returns errTemplateEmpty.
// If an error occurs parsing the pattern or executing the template,
// it returns errTemplateSyntax.
func (rep *Report) applyTemplate(pattern string) (string, error) {
	if pattern == "" {
		return "", errTemplateEmpty
	}

	t, err := template.
		New("report").
		Funcs(rep.templateFuncs()).
		Parse(pattern)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errTemplateSyntax, err)
	}

	var b strings.Builder
	if err := t.Execute(&b, rep); err != nil {
		return "", fmt.Errorf("%w: %s", errTemplateSyntax, err)
	}

	return b.String(), nil
}

func (rep *Report) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"stats": func() basicStats {
			if rep.stats.isZero() {
				rep.stats.Min, rep.stats.Max, rep.stats.Mean = rep.Benchmark.Stats()
			}
			return rep.stats
		},

		"event": func(rec requester.Record, name string) time.Duration {
			for _, e := range rec.Events {
				if e.Name == name {
					return e.Time
				}
			}
			return 0
		},

		"fail": func(a ...interface{}) string {
			if rep.errTplFailTriggered == nil {
				rep.errTplFailTriggered = fmt.Errorf(
					"%w: %s",
					ErrTplFailTriggered, fmt.Sprint(a...),
				)
			}
			return ""
		},
	}
}
