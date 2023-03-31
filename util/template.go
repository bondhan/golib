package util

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

// Template string template
type Template struct {
	template string
	params   []string
	evals    map[string]*Evaluator
}

// NewTemplate new template
func NewTemplate(str string) *Template {
	r := regexp.MustCompile(`{([^{:"}]*)}`)
	matches := r.FindAllStringSubmatch(str, -1)
	params := make([]string, 0)
	evals := make(map[string]*Evaluator)
	if len(matches) > 0 {
		for _, v := range matches {
			if strings.HasPrefix(v[1], "[") && strings.HasSuffix(v[1], "]") {
				exp := strings.TrimSuffix(strings.TrimPrefix(v[1], "["), "]")
				ev, err := NewEvaluator(exp)
				if err != nil {
					continue
				}
				evals[v[1]] = ev
				continue
			}
			params = append(params, v[1])
		}
	}

	return &Template{
		template: str,
		params:   params,
		evals:    evals,
	}
}

func (t *Template) render(obj interface{}, strict bool) (string, error) {

	if len(t.params) == 0 && len(t.evals) == 0 {
		return t.template, nil
	}

	if obj == nil && len(t.params) > 0 {
		return "", errors.New("null object")
	}

	out := t.template

	for k, v := range t.evals {
		val, err := v.Eval(obj)
		if err != nil {
			continue
		}
		out = strings.ReplaceAll(out, "{"+k+"}", fmt.Sprintf("%v", val))
	}

	for _, k := range t.params {
		lk := k
		def := ""
		unquote := false
		if dv := strings.Split(k, "|"); len(dv) > 1 {
			lk = dv[0]
			def = dv[1]
		}
		if strings.HasPrefix(lk, "'") {
			unquote = true
			lk = strings.TrimPrefix(lk, "'")
		}
		v, ok := Lookup(lk, obj)
		if !ok && strict {
			if def == "" {
				return "", errors.New("value not found")
			}
			v = def
		}
		if v == nil {
			v = def
		}
		if unquote {
			out = strings.ReplaceAll(out, `"{`+k+`}"`, fmt.Sprintf("%v", v))
			continue
		}
		out = strings.ReplaceAll(out, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return out, nil
}

// Render render template
func (t *Template) Render(obj interface{}) string {
	o, _ := t.render(obj, false)
	return o
}

// StrictRender render template and return error if one of value is missing
func (t *Template) StrictRender(obj interface{}) (string, error) {
	return t.render(obj, true)
}

func (t *Template) RenderFromStringJSON(json string) string {
	if t.params == nil || len(t.params) == 0 || json == "" {
		return t.template
	}

	out := t.template

	for k, v := range t.evals {
		val, err := v.EvalStringJSON(json)
		if err != nil {
			continue
		}
		out = strings.ReplaceAll(out, "{"+k+"}", fmt.Sprintf("%v", val))
	}

	for _, k := range t.params {
		lk := k
		def := ""
		unquote := false
		if dv := strings.Split(k, "|"); len(dv) > 1 {
			lk = dv[0]
			def = dv[1]
		}
		if strings.HasPrefix(lk, "'") {
			unquote = true
			lk = strings.TrimPrefix(lk, "'")
		}
		v := gjson.Get(json, lk).Value()
		if v == nil {
			v = def
		}
		if unquote {
			out = strings.ReplaceAll(out, `"{`+k+`}"`, fmt.Sprintf("%v", v))
			continue
		}
		out = strings.ReplaceAll(out, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return out
}
