package helpers

import "strings"

func TrimModuleName(name string) string {
	charIndex := strings.LastIndex(name, "-")
	if charIndex >= 0 && charIndex < len(name) && name[charIndex] == 45 {
		name = name[:charIndex]
	}

	return name
}
