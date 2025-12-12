package helpers

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

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

func IsVersionGreater(version1, version2 string) bool {
	semVer1, err1 := semver.NewVersion(version1)
	semVer2, err2 := semver.NewVersion(version2)
	if err1 != nil || err2 != nil {
		return version1 > version2
	}

	return semVer1.GreaterThan(semVer2)
}
