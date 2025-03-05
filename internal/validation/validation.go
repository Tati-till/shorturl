package validation

import "net/url"

func IsCorrectURL(s string) bool {
	if s == "" {
		return false
	}
	_, err := url.Parse(s)
	return err == nil
}
