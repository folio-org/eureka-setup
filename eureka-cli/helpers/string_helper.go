package helpers

import "strings"

func StripModuleVersion(name string) string {
	idx := strings.LastIndex(name, "-")
	if idx >= 0 && idx < len(name) && name[idx] == 45 {
		name = name[:idx]
	}

	return name
}

func FilterEmptyLines(input string) string {
	if input == "" {
		return ""
	}

	lines := strings.Split(input, "\n")
	filteredLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			filteredLines = append(filteredLines, line)
		}
	}
	if len(filteredLines) == 0 {
		return ""
	}

	return strings.Join(filteredLines, "\n")
}
