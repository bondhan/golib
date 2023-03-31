package validation

import (
	"regexp"
	"strings"
)

func IsValidEmail(email string, validDomains []string) bool {

	rule := regexp.MustCompile(`^[\w-.]+@([\w-]+.)+[\w-]{2,4}$`)
	if !rule.MatchString(email) {
		return false
	}

	// if list is empty, ignore it
	if len(validDomains) == 0 {
		return true
	}

	splited := strings.Split(email, "@")
	if len(splited) != 2 {
		return false
	}

	var domain = splited[1]
	var valid = false
	for _, validDomain := range validDomains {
		if validDomain == domain {
			valid = true
			break
		}
	}
	return valid
}
