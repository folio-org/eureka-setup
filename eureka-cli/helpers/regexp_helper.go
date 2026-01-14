package helpers

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
)

var (
	colonDelimited = regexp.MustCompile(constant.ColonDelimitedPattern)
	moduleId       = regexp.MustCompile(constant.ModuleIDPattern)
	newLine        = regexp.MustCompile(constant.NewLinePattern)
	protocol       = regexp.MustCompile(constant.ProtocolPattern)
)

// ==================== Vault ====================

func GetVaultRootTokenFromLogs(logLine string) string {
	return strings.TrimSpace(colonDelimited.ReplaceAllString(logLine, `$1`))
}

// ==================== Hostname ====================

func GetPortFromURL(url string) (int, error) {
	url = strings.TrimSpace(url)
	if !strings.Contains(url, ":") && !strings.Contains(url, "/") {
		return strconv.Atoi(url)
	}

	url = protocol.ReplaceAllString(url, "")
	if strings.HasPrefix(url, ":") {
		portStr := url[1:]
		if slashIdx := strings.IndexAny(portStr, "/?#"); slashIdx != -1 {
			portStr = portStr[:slashIdx]
		}
		return strconv.Atoi(strings.TrimSpace(portStr))
	}

	idx := strings.LastIndex(url, ":")
	if idx == -1 {
		return 0, strconv.ErrSyntax
	}

	portStr := url[idx+1:]
	if slashIdx := strings.IndexAny(portStr, "/?#"); slashIdx != -1 {
		portStr = portStr[:slashIdx]
	}

	return strconv.Atoi(strings.TrimSpace(portStr))
}

func GetHostnameFromURL(url string) string {
	url = protocol.ReplaceAllString(url, "")
	if idx := strings.IndexAny(url, ":/?#"); idx != -1 {
		return url[:idx]
	}

	return url
}

// ==================== Module ====================

func GetModuleNameFromID(id string) string {
	return StripModuleVersion(moduleId.ReplaceAllString(id, `$1`))
}

func GetModuleVersionFromID(id string) string {
	return moduleId.ReplaceAllString(id, `$2$3`)
}

func GetOptionalModuleVersion(id string) *string {
	version := GetModuleVersionFromID(id)
	if version == "" {
		return nil
	}
	return &version
}

func MatchesModuleName(moduleID string, moduleName string) bool {
	pattern := regexp.MustCompile("^" + regexp.QuoteMeta(moduleName) + `-\d`)
	return pattern.MatchString(moduleID)
}

// ==================== Kafka ====================

func GetKafkaConsumerLagFromLogLine(stdout bytes.Buffer) string {
	if stdout.Len() == 0 {
		return "0"
	}

	cleaned := newLine.ReplaceAllString(stdout.String(), " ")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "0"
	}

	parts := strings.Fields(cleaned)
	totalLag := 0
	for _, part := range parts {
		if part == "-" {
			continue
		}
		if lag, err := strconv.Atoi(part); err == nil {
			totalLag += lag
		}
	}

	return strconv.Itoa(totalLag)
}
