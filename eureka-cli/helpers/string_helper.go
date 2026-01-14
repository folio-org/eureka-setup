package helpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
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

func IncrementSnapshotVersion(version string) (string, error) {
	if version == "" {
		return "", errors.VersionEmpty()
	}
	if !IsSnapshot(version) {
		return "", errors.NotSnapshotVersion(version)
	}

	parts := strings.Split(version, "-SNAPSHOT.")
	if len(parts) != 2 {
		return "", errors.InvalidSnapshotFormat(version)
	}
	baseVersion := parts[0]
	buildNum := parts[1]

	num, err := strconv.Atoi(buildNum)
	if err != nil {
		return "", errors.InvalidBuildNumber(version, err)
	}

	return fmt.Sprintf("%s-SNAPSHOT.%d", baseVersion, num+1), nil
}

func IsSnapshot(version string) bool {
	return strings.Contains(version, "-SNAPSHOT.")
}

func IsFolioNamespace(namespace string) bool {
	return namespace == constant.SnapshotNamespace || namespace == constant.ReleaseNamespace
}
