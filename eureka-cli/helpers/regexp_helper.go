package helpers

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/constant"
)

func GetVaultRootTokenFromLogs(logLine string) string {
	return strings.TrimSpace(regexp.MustCompile(constant.ColonDelimitedPattern).ReplaceAllString(logLine, `$1`))
}

func GetPortFromURL(url string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(regexp.MustCompile(constant.ColonDelimitedPattern).ReplaceAllString(url, `$1`)))
}

func GetModuleName(id string) string {
	return TrimModuleName(getModuleIDRegexp().ReplaceAllString(id, `$1`))
}

func GetModuleVersionP(id string) *string {
	return StringP(GetModuleVersion(id))
}

func GetModuleVersion(id string) string {
	return getModuleIDRegexp().ReplaceAllString(id, `$2$3`)
}

func getModuleIDRegexp() *regexp.Regexp {
	return regexp.MustCompile(constant.ModuleIDPattern)
}

func GetKafkaConsumerLag(stdout bytes.Buffer) string {
	return regexp.MustCompile(constant.NewLinePattern).ReplaceAllString(stdout.String(), "")
}
