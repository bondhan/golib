package log

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
)

type SafeJSONFormatter struct {
	sensitiveFields []string
	logrus.JSONFormatter
}

func (sf *SafeJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	for _, field := range sf.sensitiveFields {
		r, err := regexp.Compile(fmt.Sprintf(`%s:".*?"`, field))
		if err == nil {
			entry.Message = r.ReplaceAllString(entry.Message, fmt.Sprintf(`%s:******`, field))
		}

	}
	return sf.JSONFormatter.Format(entry)
}
