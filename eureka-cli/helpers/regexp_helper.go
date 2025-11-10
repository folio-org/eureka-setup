package helpers

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/constant"
)

func GetVaultRootTokenFromLogs(logLine string) string {
	pattern := regexp.MustCompile(constant.ColonDelimitedPattern)
	return strings.TrimSpace(pattern.ReplaceAllString(logLine, `$1`))
}

func GetPortFromURL(url string) (int, error) {
	pattern := regexp.MustCompile(constant.ColonDelimitedPattern)
	return strconv.Atoi(strings.TrimSpace(pattern.ReplaceAllString(url, `$1`)))
}

func GetModuleNameFromID(id string) string {
	pattern := getModuleIDRegexp()
	return TrimModuleName(pattern.ReplaceAllString(id, `$1`))
}

func getModuleIDRegexp() *regexp.Regexp {
	return regexp.MustCompile(constant.ModuleIDPattern)
}

func GetOptionalModuleVersion(id string) *string {
	version := GetModuleVersionFromID(id)
	if version == "" {
		return nil
	}
	return &version
}

func GetModuleVersionFromID(id string) string {
	pattern := getModuleIDRegexp()
	return pattern.ReplaceAllString(id, `$2$3`)
}
func GetKafkaConsumerLagFromLogLine(stdout bytes.Buffer) string {
	pattern := regexp.MustCompile(constant.NewLinePattern)
	return pattern.ReplaceAllString(stdout.String(), "")
}

func MatchesModuleName(moduleID string, moduleName string) bool {
	pattern := regexp.MustCompile("^" + regexp.QuoteMeta(moduleName) + `-\d`)
	return pattern.MatchString(moduleID)
}
