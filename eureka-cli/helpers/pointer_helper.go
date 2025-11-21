package helpers

import "time"

func StringPtr(value string) *string {
	return &value
}

func BoolPtr(value bool) *bool {
	return &value
}

func IntPtr(value int) *int {
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
