package helpers

import "time"

func StringP(value string) *string {
	return &value
}

func BoolP(value bool) *bool {
	return &value
}

func IntP(value int) *int {
	return &value
}

func DefaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}

	return value
}

func DefaultDuration(value, defaultValue time.Duration) time.Duration {
	if value == 0 {
		return defaultValue
	}

	return value
}
