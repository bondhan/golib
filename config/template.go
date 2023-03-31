package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Template string template
type ConfigTemplate struct {
	template string
	params   []string
}

// NewTemplate new template
func NewTemplate(str string) *ConfigTemplate {
	r := regexp.MustCompile(`{([^{}]*)}`)
	matches := r.FindAllStringSubmatch(str, -1)
	params := make([]string, 0)
	if len(matches) > 0 {
		for _, v := range matches {
			params = append(params, v[1])
		}
	}

	return &ConfigTemplate{
		template: str,
		params:   params,
	}
}

func (t *ConfigTemplate) Render() string {

	if t.params == nil || len(t.params) == 0 {
		return t.template
	}

	out := t.template

	for _, k := range t.params {
		lk := k
		def := ""
		if dv := strings.Split(k, "|"); len(dv) > 1 {
			lk = dv[0]
			def = dv[1]
		}

		if !strings.HasPrefix(lk, "$") {
			continue
		}

		lk = strings.TrimPrefix(lk, "$")

		v := os.Getenv(lk)

		if v == "" && def != "" {
			v = def
		}

		out = strings.ReplaceAll(out, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return out
}
